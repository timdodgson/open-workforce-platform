// policy_exports.go re-exports the optimisation/policy subpackage for stable call sites.
package optimisation

import (
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation/policy"
)

type (
	Policy                      = policy.Policy
	PolicyContext               = policy.PolicyContext
	PolicyDecision              = policy.PolicyDecision
	PolicyMetadata              = policy.PolicyMetadata
	Rule                        = policy.Rule
	RulePolicy                  = policy.RulePolicy
	PolicyModel                 = policy.PolicyModel
	ModelPrediction             = policy.ModelPrediction
	LearnedPolicy               = policy.LearnedPolicy
	LearnedPolicyConfig         = policy.LearnedPolicyConfig
	HybridPolicy                = policy.HybridPolicy
	PolicyProvider              = policy.PolicyProvider
	PolicyHierarchy             = policy.PolicyHierarchy
	PolicyLevel                 = policy.PolicyLevel
	HierarchicalDecision        = policy.HierarchicalDecision
	TransferPolicy              = policy.TransferPolicy
	TransferPolicyConfig        = policy.TransferPolicyConfig
	HierarchyEntry              = policy.HierarchyEntry
	FeatureVector               = policy.FeatureVector
	FeatureStore                = policy.FeatureStore
	FeatureRecord               = policy.FeatureRecord
	FeatureOutcome              = policy.FeatureOutcome
	SklearnTree                 = policy.SklearnTree
	ImprovementCurveModel       = policy.ImprovementCurveModel
	StagnationClassifierEntry   = policy.StagnationClassifierEntry
	ImprovementCurveEntry       = policy.ImprovementCurveEntry
	StagnationAssessment        = policy.StagnationAssessment
	StagnationPolicyConfig      = policy.StagnationPolicyConfig
	LearnedStagnationDetector   = policy.LearnedStagnationDetector
	RestartModel                = policy.RestartModel
	RestartModelEntry           = policy.RestartModelEntry
	RestartDecision             = policy.RestartDecision
	RestartPolicyConfig         = policy.RestartPolicyConfig
	RestartPolicy               = policy.RestartPolicy
	RestartEffectivenessRecord  = policy.RestartEffectivenessRecord
	PolicySearchDecision        = policy.PolicySearchDecision
	PolicySearchConfig          = policy.PolicySearchConfig
	ShadowRecord                = policy.ShadowRecord
	ShadowMetrics               = policy.ShadowMetrics
	PolicyShadowRunner          = policy.PolicyShadowRunner
	FeatureContribution         = policy.FeatureContribution
	PolicyExplanation           = policy.PolicyExplanation
	ExplanationBuilder          = policy.ExplanationBuilder
	PolicyEvaluationRecord      = policy.PolicyEvaluationRecord
	PolicyEvaluator             = policy.PolicyEvaluator
	PolicyMetrics               = policy.PolicyMetrics
	CalibrationBucket           = policy.CalibrationBucket
	PolicyEvaluationInput       = policy.PolicyEvaluationInput
	PolicyStatus                = policy.PolicyStatus
	PolicyVersionRecord         = policy.PolicyVersionRecord
	PolicyLifecycleRegistry     = policy.PolicyLifecycleRegistry
	PolicyComparison            = policy.PolicyComparison
	ContinuousLearningConfig    = policy.ContinuousLearningConfig
	LearningState               = policy.LearningState
	ContinuousLearner           = policy.ContinuousLearner
	LearningRecommendation      = policy.LearningRecommendation
	PromotionRules              = policy.PromotionRules
	PromotionResult             = policy.PromotionResult
	PolicyPromoter              = policy.PolicyPromoter
	GateResult                  = policy.GateResult
	PostRunPolicyConfig         = policy.PostRunPolicyConfig
	PostRunPolicyReport         = policy.PostRunPolicyReport
	TrainingPipelineConfig      = policy.TrainingPipelineConfig
	TrainingDataset             = policy.TrainingDataset
	TrainingSample              = policy.TrainingSample
	TrainingResult              = policy.TrainingResult
	TrainingPipeline            = policy.TrainingPipeline
	PolicyReport                = policy.PolicyReport
)

const (
	FeatureSchemaVersion   = policy.FeatureSchemaVersion
	StagnationPolicyFile   = policy.StagnationPolicyFile
	RestartPolicyFile      = policy.RestartPolicyFile
	PolicyStatusTraining   = policy.PolicyStatusTraining
	PolicyStatusShadow   = policy.PolicyStatusShadow
	PolicyStatusActive   = policy.PolicyStatusActive
	PolicyStatusRetired  = policy.PolicyStatusRetired
	LevelInstance        = policy.LevelInstance
	LevelDomain          = policy.LevelDomain
	LevelGlobal          = policy.LevelGlobal
	LevelNone            = policy.LevelNone
)

func NewRulePolicy(id, version, domain, decisionType string, rules []Rule) *RulePolicy {
	return policy.NewRulePolicy(id, version, domain, decisionType, rules)
}

func NewLearnedPolicy(cfg LearnedPolicyConfig) *LearnedPolicy {
	return policy.NewLearnedPolicy(cfg)
}

func NewHybridPolicy(learned *LearnedPolicy, fallback *RulePolicy) *HybridPolicy {
	return policy.NewHybridPolicy(learned, fallback)
}

func NewPolicyProvider() *PolicyProvider { return policy.NewPolicyProvider() }

func NewPolicyHierarchy() *PolicyHierarchy { return policy.NewPolicyHierarchy() }

func NewTransferPolicy(cfg TransferPolicyConfig) *TransferPolicy {
	return policy.NewTransferPolicy(cfg)
}

func NewFeatureStore(dir string) *FeatureStore { return policy.NewFeatureStore(dir) }

func LoadImprovementCurveModel(path string) (*ImprovementCurveModel, error) {
	return policy.LoadImprovementCurveModel(path)
}

func DefaultStagnationPolicyConfig() StagnationPolicyConfig {
	return policy.DefaultStagnationPolicyConfig()
}

func NewLearnedStagnationDetector(model *ImprovementCurveModel, config StagnationPolicyConfig) *LearnedStagnationDetector {
	return policy.NewLearnedStagnationDetector(model, config)
}

func LoadRestartModel(path string) (*RestartModel, error) {
	return policy.LoadRestartModel(path)
}

func DefaultRestartPolicyConfig() RestartPolicyConfig {
	return policy.DefaultRestartPolicyConfig()
}

func NewRestartPolicy(model *RestartModel, config RestartPolicyConfig) *RestartPolicy {
	return policy.NewRestartPolicy(model, config)
}

func WritePolicyDecisionsCSV(path string, decisions []PolicySearchDecision) error {
	return policy.WritePolicyDecisionsCSV(path, decisions)
}

func NewPolicyShadowRunner(dir string) *PolicyShadowRunner {
	return policy.NewPolicyShadowRunner(dir)
}

func NewExplanationBuilder() *ExplanationBuilder {
	return policy.NewExplanationBuilder()
}

func NewPolicyEvaluator() *PolicyEvaluator { return policy.NewPolicyEvaluator() }

func BuildPolicyEvaluationRecords(in PolicyEvaluationInput) []PolicyEvaluationRecord {
	return policy.BuildPolicyEvaluationRecords(in)
}

func WritePolicyEvaluationCSV(dir string, in PolicyEvaluationInput) error {
	return policy.WritePolicyEvaluationCSV(dir, in)
}

func DefaultPolicyDir() string { return policy.DefaultPolicyDir() }

func ResolvePolicyDir(policyMode, policyDir string) string {
	return policy.ResolvePolicyDir(policyMode, policyDir)
}

func LoadPolicyRegistry(path string) (*PolicyLifecycleRegistry, error) {
	return policy.LoadPolicyRegistry(path)
}

func NewContinuousLearner(config ContinuousLearningConfig) *ContinuousLearner {
	return policy.NewContinuousLearner(config)
}

func NewPolicyPromoter(rules PromotionRules, registry *PolicyLifecycleRegistry) *PolicyPromoter {
	return policy.NewPolicyPromoter(rules, registry)
}

func DefaultPromotionRules() PromotionRules { return policy.DefaultPromotionRules() }

func RunPostRunPolicyPipeline(cfg PostRunPolicyConfig) *PostRunPolicyReport {
	return policy.RunPostRunPolicyPipeline(cfg)
}

func FormatPostRunSummary(report *PostRunPolicyReport) string {
	return policy.FormatPostRunSummary(report)
}

func DefaultTrainingPipelineConfig() TrainingPipelineConfig {
	return policy.DefaultTrainingPipelineConfig()
}

func NewTrainingPipeline(config TrainingPipelineConfig) (*TrainingPipeline, error) {
	return policy.NewTrainingPipeline(config)
}

func SaveReport(report PolicyReport, dir string) error {
	return policy.SaveReport(report, dir)
}
