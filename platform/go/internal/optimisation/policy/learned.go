// policy_learned.go — LearnedPolicy implementation.
//
// A model-based policy that maps FeatureVectors to decisions.
// Defers (returns Action="defer") when confidence is below threshold.
// Callers must handle the defer case (typically by using HybridPolicy).
package policy

import "time"

// PolicyModel is the interface for any trained model that can produce decisions.
// Implementations may use decision trees, gradient boosting, neural networks, etc.
type PolicyModel interface {
	// Predict returns an action and confidence for the given features.
	Predict(features FeatureVector) ModelPrediction
}

// ModelPrediction is the raw output of a trained model.
type ModelPrediction struct {
	Action        string
	Confidence    float64
	ExpectedValue float64
	Reason        string
}

// LearnedPolicy wraps a trained model in the Policy interface.
type LearnedPolicy struct {
	id        string
	version   string
	domain    string
	decType   string
	model     PolicyModel
	threshold float64 // minimum confidence to produce a decision
	trained   int
	created   time.Time
	valScore  float64
}

// LearnedPolicyConfig holds construction parameters.
type LearnedPolicyConfig struct {
	ID              string
	Version         string
	Domain          string
	DecisionType    string
	Model           PolicyModel
	Threshold       float64
	TrainedSamples  int
	CreatedAt       time.Time
	ValidationScore float64
}

// NewLearnedPolicy creates a model-based policy.
func NewLearnedPolicy(cfg LearnedPolicyConfig) *LearnedPolicy {
	threshold := cfg.Threshold
	if threshold <= 0 {
		threshold = 0.60 // default confidence threshold
	}
	return &LearnedPolicy{
		id:        cfg.ID,
		version:   cfg.Version,
		domain:    cfg.Domain,
		decType:   cfg.DecisionType,
		model:     cfg.Model,
		threshold: threshold,
		trained:   cfg.TrainedSamples,
		created:   cfg.CreatedAt,
		valScore:  cfg.ValidationScore,
	}
}

// Decide uses the trained model to produce a decision.
// Returns Action="defer" if confidence is below threshold.
func (p *LearnedPolicy) Decide(ctx PolicyContext) PolicyDecision {
	prediction := p.model.Predict(ctx.Features)

	if prediction.Confidence < p.threshold {
		return PolicyDecision{
			Action:         "defer",
			Confidence:     prediction.Confidence,
			Reason:         "learned_low_confidence",
			PolicyID:       p.id,
			PolicyVersion:  p.version,
			IsFallback:     false,
			FallbackReason: "",
		}
	}

	return PolicyDecision{
		Action:        prediction.Action,
		Confidence:    prediction.Confidence,
		Reason:        prediction.Reason,
		PolicyID:      p.id,
		PolicyVersion: p.version,
		IsFallback:    false,
	}
}

// Metadata returns identity information for this learned policy.
func (p *LearnedPolicy) Metadata() PolicyMetadata {
	return PolicyMetadata{
		ID:              p.id,
		Version:         p.version,
		Type:            "learned",
		Domain:          p.domain,
		DecisionType:    p.decType,
		TrainedSamples:  p.trained,
		CreatedAt:       p.created,
		ValidationScore: p.valScore,
	}
}
