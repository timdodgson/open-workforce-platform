// post_run_policy.go — Post-run continuous learning and promotion evaluation.
package policy

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const defaultPolicyDir = "../ml/policies"

// DefaultPolicyDir is the default --policy-dir when --policy-mode is set.
func DefaultPolicyDir() string {
	return defaultPolicyDir
}

// ResolvePolicyDir returns policyDir or the default when policy mode is active.
func ResolvePolicyDir(policyMode, policyDir string) string {
	if policyMode == "" {
		return policyDir
	}
	if policyDir != "" {
		return policyDir
	}
	return defaultPolicyDir
}

// PostRunPolicyConfig configures the post-run SI 2.0 pipeline.
type PostRunPolicyConfig struct {
	PolicyMode            string
	PolicyDir             string
	OutputDir             string
	Domain                string
	PolicyDecisionCount   int
	AssistRecordCount     int
}

// PostRunPolicyReport captures post-run learning and promotion results.
type PostRunPolicyReport struct {
	LearningRecommendation LearningRecommendation `json:"learning_recommendation"`
	PromotionResults       []PromotionResult      `json:"promotion_results,omitempty"`
	SamplesAdded           int                    `json:"samples_added"`
}

// RunPostRunPolicyPipeline runs continuous learning and promotion checks after a solve.
func RunPostRunPolicyPipeline(cfg PostRunPolicyConfig) *PostRunPolicyReport {
	if cfg.PolicyMode == "" {
		return nil
	}

	policyDir := ResolvePolicyDir(cfg.PolicyMode, cfg.PolicyDir)
	samples := cfg.PolicyDecisionCount + cfg.AssistRecordCount
	if samples <= 0 {
		samples = 1
	}

	learner := NewContinuousLearner(ContinuousLearningConfig{
		RetrainThreshold: 50,
		DataDir:          filepath.Join(policyDir, "training_data"),
		PolicyDir:          policyDir,
		RegistryPath:       filepath.Join(policyDir, "policy_registry.json"),
	})

	report := &PostRunPolicyReport{
		LearningRecommendation: learner.OnRunCompleted(filepath.Base(cfg.OutputDir), cfg.Domain, samples),
		SamplesAdded:           samples,
	}

	registryPath := filepath.Join(policyDir, "policy_registry.json")
	if reg, err := LoadPolicyRegistry(registryPath); err == nil {
		promoter := NewPolicyPromoter(DefaultPromotionRules(), reg)
		report.PromotionResults = promoter.EvaluateAll()
		_ = reg.Save(registryPath)
	}

	if cfg.OutputDir != "" {
		data, _ := json.MarshalIndent(report, "", "  ")
		_ = os.WriteFile(filepath.Join(cfg.OutputDir, "policy_learning_report.json"), data, 0o644)
	}

	return report
}

// FormatPostRunSummary returns a one-line summary for stderr logging.
func FormatPostRunSummary(report *PostRunPolicyReport) string {
	if report == nil {
		return ""
	}
	return fmt.Sprintf("SI learning: %s (%s)", report.LearningRecommendation.Action, report.LearningRecommendation.Reason)
}
