// training.go — Policy Training Pipeline.
//
// Orchestrates the full lifecycle from raw telemetry to production policy:
//
//	Telemetry → Feature Engineering → Training → Validation → Candidate →
//	Shadow Deployment → Evaluation → Promotion → Production
//
// The pipeline does not train models itself — it orchestrates the steps and
// validates that each gate passes before advancing the candidate.
package policy

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// TrainingPipelineConfig defines thresholds and paths for the pipeline.
type TrainingPipelineConfig struct {
	MinOfflineAccuracy        float64
	MinShadowAccuracy         float64
	MinSamples                int
	MinShadowRuns             int
	MaxRegretForAutoPromotion float64
	OutputDir                 string
	RegistryPath              string
}

// DefaultTrainingPipelineConfig returns sensible defaults.
func DefaultTrainingPipelineConfig() TrainingPipelineConfig {
	return TrainingPipelineConfig{
		MinOfflineAccuracy:        0.65,
		MinShadowAccuracy:         0.60,
		MinSamples:                50,
		MinShadowRuns:             20,
		MaxRegretForAutoPromotion: 0.0,
		OutputDir:                 "policies",
		RegistryPath:              "policy_registry.json",
	}
}

// TrainingDataset represents the engineered features ready for model training.
type TrainingDataset struct {
	Domain       string
	DecisionType string
	Algorithm    string
	Samples      []TrainingSample
	Features     []string
	CreatedAt    time.Time
}

// TrainingSample is one labelled example: features + correct action + outcome.
type TrainingSample struct {
	Features FeatureVector
	Action   string
	Outcome  float64
	Correct  bool
}

// TrainingResult captures the output of one training run.
type TrainingResult struct {
	PolicyID          string
	Version           string
	Domain            string
	DecisionType      string
	Algorithm         string
	TrainingSamples   int
	Features          []string
	TrainedAt         time.Time
	OfflineAccuracy   float64
	OfflinePrecision  float64
	OfflineRecall     float64
	ValidationSamples int
	ModelPath         string
	PassedValidation  bool
	FailureReason     string
}

// TrainingPipeline orchestrates policy training, validation, and promotion.
type TrainingPipeline struct {
	config   TrainingPipelineConfig
	registry *PolicyLifecycleRegistry
}

// NewTrainingPipeline creates a pipeline with the given config.
func NewTrainingPipeline(config TrainingPipelineConfig) (*TrainingPipeline, error) {
	registry, err := LoadPolicyRegistry(config.RegistryPath)
	if err != nil {
		return nil, fmt.Errorf("training pipeline: %w", err)
	}
	return &TrainingPipeline{config: config, registry: registry}, nil
}

// ValidateTrainingData checks that the dataset meets minimum requirements.
func (p *TrainingPipeline) ValidateTrainingData(dataset TrainingDataset) GateResult {
	if len(dataset.Samples) < p.config.MinSamples {
		return GateResult{
			Gate: "min_samples", Passed: false,
			Reason:    fmt.Sprintf("need %d samples, have %d", p.config.MinSamples, len(dataset.Samples)),
			Value:     float64(len(dataset.Samples)),
			Threshold: float64(p.config.MinSamples),
		}
	}
	if len(dataset.Features) == 0 {
		return GateResult{
			Gate: "features", Passed: false,
			Reason: "no features defined",
		}
	}
	return GateResult{Gate: "min_samples", Passed: true, Value: float64(len(dataset.Samples)), Threshold: float64(p.config.MinSamples)}
}

// ValidateOfflineAccuracy checks that training result meets accuracy threshold.
func (p *TrainingPipeline) ValidateOfflineAccuracy(result TrainingResult) GateResult {
	if result.OfflineAccuracy < p.config.MinOfflineAccuracy {
		return GateResult{
			Gate: "offline_accuracy", Passed: false,
			Reason:    fmt.Sprintf("accuracy %.2f below threshold %.2f", result.OfflineAccuracy, p.config.MinOfflineAccuracy),
			Value:     result.OfflineAccuracy,
			Threshold: p.config.MinOfflineAccuracy,
		}
	}
	return GateResult{Gate: "offline_accuracy", Passed: true, Value: result.OfflineAccuracy, Threshold: p.config.MinOfflineAccuracy}
}

// ValidateShadowPerformance checks shadow results meet promotion threshold.
func (p *TrainingPipeline) ValidateShadowPerformance(shadowAccuracy float64, shadowRuns int) GateResult {
	if shadowRuns < p.config.MinShadowRuns {
		return GateResult{
			Gate: "shadow_runs", Passed: false,
			Reason:    fmt.Sprintf("need %d shadow runs, have %d", p.config.MinShadowRuns, shadowRuns),
			Value:     float64(shadowRuns),
			Threshold: float64(p.config.MinShadowRuns),
		}
	}
	if shadowAccuracy < p.config.MinShadowAccuracy {
		return GateResult{
			Gate: "shadow_accuracy", Passed: false,
			Reason:    fmt.Sprintf("shadow accuracy %.2f below threshold %.2f", shadowAccuracy, p.config.MinShadowAccuracy),
			Value:     shadowAccuracy,
			Threshold: p.config.MinShadowAccuracy,
		}
	}
	return GateResult{Gate: "shadow_accuracy", Passed: true, Value: shadowAccuracy, Threshold: p.config.MinShadowAccuracy}
}

// RegisterCandidate registers a validated training result as a candidate policy.
func (p *TrainingPipeline) RegisterCandidate(result TrainingResult) error {
	p.registry.Register(PolicyVersionRecord{
		ID:                 result.PolicyID,
		Version:            result.Version,
		Domain:             result.Domain,
		DecisionType:       result.DecisionType,
		Algorithm:          result.Algorithm,
		CreatedAt:          result.TrainedAt,
		TrainingSamples:    result.TrainingSamples,
		TrainingDate:       result.TrainedAt,
		Features:           result.Features,
		OfflineAccuracy:    result.OfflineAccuracy,
		ShadowAccuracy:     -1,
		ProductionAccuracy: -1,
		ModelPath:          result.ModelPath,
	})
	return p.registry.Save(p.config.RegistryPath)
}

// DeployToShadow promotes a candidate to shadow mode.
func (p *TrainingPipeline) DeployToShadow(policyID string, version string) error {
	if err := p.registry.PromoteToShadow(policyID, version); err != nil {
		return err
	}
	return p.registry.Save(p.config.RegistryPath)
}

// PromoteToProduction promotes a shadow policy to active production.
func (p *TrainingPipeline) PromoteToProduction(policyID string, version string, shadowAccuracy float64, shadowRuns int) (GateResult, error) {
	gate := p.ValidateShadowPerformance(shadowAccuracy, shadowRuns)
	if !gate.Passed {
		return gate, fmt.Errorf("promotion gate failed: %s", gate.Reason)
	}

	if err := p.registry.UpdateShadowAccuracy(policyID, version, shadowAccuracy); err != nil {
		return gate, err
	}
	if err := p.registry.PromoteToActive(policyID, version); err != nil {
		return gate, err
	}
	return gate, p.registry.Save(p.config.RegistryPath)
}

// Rollback reverts to a previous policy version.
func (p *TrainingPipeline) Rollback(domain string, decisionType string, targetVersion string, reason string) error {
	if err := p.registry.Rollback(domain, decisionType, targetVersion, reason); err != nil {
		return err
	}
	return p.registry.Save(p.config.RegistryPath)
}

// Registry returns the underlying registry for inspection.
func (p *TrainingPipeline) Registry() *PolicyLifecycleRegistry {
	return p.registry
}

// PolicyReport is a structured summary of a training pipeline run.
type PolicyReport struct {
	PolicyID          string       `json:"policy_id"`
	Version           string       `json:"version"`
	Domain            string       `json:"domain"`
	DecisionType      string       `json:"decision_type"`
	GeneratedAt       time.Time    `json:"generated_at"`
	TotalSamples      int          `json:"total_samples"`
	TrainingSamples   int          `json:"training_samples"`
	ValidationSamples int          `json:"validation_samples"`
	Features          []string     `json:"features"`
	OfflineAccuracy   float64      `json:"offline_accuracy"`
	ShadowAccuracy    float64      `json:"shadow_accuracy"`
	RegretVsRules     float64      `json:"regret_vs_rules"`
	Gates             []GateResult `json:"gates"`
	FinalStatus       string       `json:"final_status"`
	StatusReason      string       `json:"status_reason"`
}

// GenerateReport creates a training report from pipeline state.
func (p *TrainingPipeline) GenerateReport(result TrainingResult, shadowAccuracy float64, shadowRuns int) PolicyReport {
	gates := []GateResult{
		p.ValidateTrainingData(TrainingDataset{
			Samples:  make([]TrainingSample, result.TrainingSamples),
			Features: result.Features,
		}),
		p.ValidateOfflineAccuracy(result),
	}

	if shadowRuns > 0 {
		gates = append(gates, p.ValidateShadowPerformance(shadowAccuracy, shadowRuns))
	}

	allPassed := true
	var failReason string
	for _, g := range gates {
		if !g.Passed {
			allPassed = false
			failReason = g.Reason
			break
		}
	}

	status := "pending"
	if !allPassed {
		status = "rejected"
	} else if shadowAccuracy > 0 && shadowRuns >= p.config.MinShadowRuns {
		status = "promoted"
	} else if result.PassedValidation {
		status = "shadow"
	}

	return PolicyReport{
		PolicyID:          result.PolicyID,
		Version:           result.Version,
		Domain:            result.Domain,
		DecisionType:      result.DecisionType,
		GeneratedAt:       time.Now(),
		TotalSamples:      result.TrainingSamples + result.ValidationSamples,
		TrainingSamples:   result.TrainingSamples,
		ValidationSamples: result.ValidationSamples,
		Features:          result.Features,
		OfflineAccuracy:   result.OfflineAccuracy,
		ShadowAccuracy:    shadowAccuracy,
		Gates:             gates,
		FinalStatus:       status,
		StatusReason:      failReason,
	}
}

// SaveReport writes a policy report to JSON.
func SaveReport(report PolicyReport, dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("save report: mkdir: %w", err)
	}
	filename := fmt.Sprintf("policy_report_%s_%s.json", report.PolicyID, report.Version)
	path := filepath.Join(dir, filename)
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("save report: marshal: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}
