// policy_exports.go re-exports the optimisation/policy subpackage for stable call sites.
package optimisation

import (
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation/policy"
)

type (
	PolicyStatus             = policy.PolicyStatus
	PolicyVersionRecord      = policy.PolicyVersionRecord
	PolicyLifecycleRegistry  = policy.PolicyLifecycleRegistry
	PolicyComparison         = policy.PolicyComparison
	ContinuousLearningConfig = policy.ContinuousLearningConfig
	LearningState            = policy.LearningState
	ContinuousLearner        = policy.ContinuousLearner
	LearningRecommendation   = policy.LearningRecommendation
	PromotionRules           = policy.PromotionRules
	PromotionResult          = policy.PromotionResult
	PolicyPromoter           = policy.PolicyPromoter
	GateResult               = policy.GateResult
	PostRunPolicyConfig      = policy.PostRunPolicyConfig
	PostRunPolicyReport      = policy.PostRunPolicyReport
)

const (
	PolicyStatusTraining = policy.PolicyStatusTraining
	PolicyStatusShadow   = policy.PolicyStatusShadow
	PolicyStatusActive   = policy.PolicyStatusActive
	PolicyStatusRetired  = policy.PolicyStatusRetired
)

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
