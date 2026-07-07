// policy_lifecycle.go — Policy Lifecycle Management.
//
// Manages multiple policy versions with full provenance tracking.
// Supports: promotion, rollback, comparison, and retirement.
//
// Lifecycle:
//
//	training → shadow → active → retired
//
// Each version records:
//   - Training data (samples, date, features used)
//   - Offline accuracy
//   - Shadow accuracy (once shadow-tested)
//   - Production accuracy (once promoted)
//   - Domain and algorithm scope
//
// The registry persists as policy_registry.json.
package optimisation

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"
)

// ───────────────────────────────────────────────────────────────
// Policy Status
// ───────────────────────────────────────────────────────────────

// PolicyStatus represents where a policy is in its lifecycle.
type PolicyStatus string

const (
	PolicyStatusTraining PolicyStatus = "training"
	PolicyStatusShadow   PolicyStatus = "shadow"
	PolicyStatusActive   PolicyStatus = "active"
	PolicyStatusRetired  PolicyStatus = "retired"
)

// ───────────────────────────────────────────────────────────────
// Policy Version Record
// ───────────────────────────────────────────────────────────────

// PolicyVersionRecord captures full provenance for one policy version.
type PolicyVersionRecord struct {
	// Identity
	ID           string `json:"id"`
	Version      string `json:"version"`
	Domain       string `json:"domain"`
	DecisionType string `json:"decision_type"`
	Algorithm    string `json:"algorithm"` // "*" if domain-wide

	// Lifecycle
	Status     PolicyStatus `json:"status"`
	CreatedAt  time.Time    `json:"created_at"`
	PromotedAt *time.Time   `json:"promoted_at,omitempty"`
	RetiredAt  *time.Time   `json:"retired_at,omitempty"`

	// Training provenance
	TrainingSamples int       `json:"training_samples"`
	TrainingDate    time.Time `json:"training_date"`
	Features        []string  `json:"features"` // feature names used by model

	// Accuracy at different stages
	OfflineAccuracy    float64 `json:"offline_accuracy"`
	ShadowAccuracy     float64 `json:"shadow_accuracy"`     // -1 if not yet shadow-tested
	ProductionAccuracy float64 `json:"production_accuracy"` // -1 if not yet in production
	ProductionRuns     int     `json:"production_runs"`

	// Performance vs baseline
	RegretVsRules float64 `json:"regret_vs_rules"` // negative = better than rules
	DriftDetected bool    `json:"drift_detected"`

	// File reference
	ModelPath string `json:"model_path"`

	// Rollback info
	RolledBackFrom string `json:"rolled_back_from,omitempty"` // version we rolled back from
	RollbackReason string `json:"rollback_reason,omitempty"`
}

// ───────────────────────────────────────────────────────────────
// Policy Registry
// ───────────────────────────────────────────────────────────────

// PolicyLifecycleRegistry manages all policy versions.
type PolicyLifecycleRegistry struct {
	Versions []PolicyVersionRecord `json:"versions"`
}

// LoadPolicyRegistry loads the registry from JSON.
func LoadPolicyRegistry(path string) (*PolicyLifecycleRegistry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &PolicyLifecycleRegistry{}, nil
		}
		return nil, fmt.Errorf("load policy registry: %w", err)
	}
	var reg PolicyLifecycleRegistry
	if err := json.Unmarshal(data, &reg); err != nil {
		return nil, fmt.Errorf("parse policy registry: %w", err)
	}
	return &reg, nil
}

// Save persists the registry to JSON.
func (r *PolicyLifecycleRegistry) Save(path string) error {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal policy registry: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}

// ───────────────────────────────────────────────────────────────
// Query Methods
// ───────────────────────────────────────────────────────────────

// ActiveVersion returns the currently active version for a domain/decision type.
// Returns nil if no active version exists.
func (r *PolicyLifecycleRegistry) ActiveVersion(domain string, decisionType string) *PolicyVersionRecord {
	for i := range r.Versions {
		v := &r.Versions[i]
		if v.Domain == domain && v.DecisionType == decisionType && v.Status == PolicyStatusActive {
			return v
		}
	}
	return nil
}

// VersionHistory returns all versions for a domain/decision type, ordered by creation date.
func (r *PolicyLifecycleRegistry) VersionHistory(domain string, decisionType string) []PolicyVersionRecord {
	var result []PolicyVersionRecord
	for _, v := range r.Versions {
		if v.Domain == domain && v.DecisionType == decisionType {
			result = append(result, v)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.Before(result[j].CreatedAt)
	})
	return result
}

// FindVersion returns a specific version by ID and version string.
func (r *PolicyLifecycleRegistry) FindVersion(id string, version string) *PolicyVersionRecord {
	for i := range r.Versions {
		if r.Versions[i].ID == id && r.Versions[i].Version == version {
			return &r.Versions[i]
		}
	}
	return nil
}

// AllActive returns all active policies across all domains.
func (r *PolicyLifecycleRegistry) AllActive() []PolicyVersionRecord {
	var result []PolicyVersionRecord
	for _, v := range r.Versions {
		if v.Status == PolicyStatusActive {
			result = append(result, v)
		}
	}
	return result
}

// ───────────────────────────────────────────────────────────────
// Lifecycle Operations
// ───────────────────────────────────────────────────────────────

// Register adds a new policy version in training status.
func (r *PolicyLifecycleRegistry) Register(rec PolicyVersionRecord) {
	rec.Status = PolicyStatusTraining
	if rec.ShadowAccuracy == 0 {
		rec.ShadowAccuracy = -1
	}
	if rec.ProductionAccuracy == 0 {
		rec.ProductionAccuracy = -1
	}
	r.Versions = append(r.Versions, rec)
}

// PromoteToShadow moves a version from training to shadow.
func (r *PolicyLifecycleRegistry) PromoteToShadow(id string, version string) error {
	v := r.FindVersion(id, version)
	if v == nil {
		return fmt.Errorf("version %s/%s not found", id, version)
	}
	if v.Status != PolicyStatusTraining {
		return fmt.Errorf("version %s/%s is %s, expected training", id, version, v.Status)
	}
	v.Status = PolicyStatusShadow
	return nil
}

// PromoteToActive moves a version from shadow to active.
// Retires the previously active version for the same domain/decision type.
func (r *PolicyLifecycleRegistry) PromoteToActive(id string, version string) error {
	v := r.FindVersion(id, version)
	if v == nil {
		return fmt.Errorf("version %s/%s not found", id, version)
	}
	if v.Status != PolicyStatusShadow {
		return fmt.Errorf("version %s/%s is %s, expected shadow", id, version, v.Status)
	}

	// Retire current active version.
	current := r.ActiveVersion(v.Domain, v.DecisionType)
	if current != nil {
		now := time.Now()
		current.Status = PolicyStatusRetired
		current.RetiredAt = &now
	}

	now := time.Now()
	v.Status = PolicyStatusActive
	v.PromotedAt = &now
	return nil
}

// Rollback reverts to a previous version. Retires the current active and
// reactivates the specified version.
func (r *PolicyLifecycleRegistry) Rollback(domain string, decisionType string, targetVersion string, reason string) error {
	// Find target.
	var target *PolicyVersionRecord
	for i := range r.Versions {
		v := &r.Versions[i]
		if v.Domain == domain && v.DecisionType == decisionType && v.Version == targetVersion {
			target = v
			break
		}
	}
	if target == nil {
		return fmt.Errorf("target version %s not found for %s/%s", targetVersion, domain, decisionType)
	}

	// Retire current.
	current := r.ActiveVersion(domain, decisionType)
	rolledBackFrom := ""
	if current != nil {
		now := time.Now()
		current.Status = PolicyStatusRetired
		current.RetiredAt = &now
		rolledBackFrom = current.Version
	}

	// Reactivate target.
	now := time.Now()
	target.Status = PolicyStatusActive
	target.PromotedAt = &now
	target.RolledBackFrom = rolledBackFrom
	target.RollbackReason = reason
	return nil
}

// Retire explicitly retires a version.
func (r *PolicyLifecycleRegistry) Retire(id string, version string) error {
	v := r.FindVersion(id, version)
	if v == nil {
		return fmt.Errorf("version %s/%s not found", id, version)
	}
	now := time.Now()
	v.Status = PolicyStatusRetired
	v.RetiredAt = &now
	return nil
}

// UpdateShadowAccuracy sets the shadow accuracy for a version.
func (r *PolicyLifecycleRegistry) UpdateShadowAccuracy(id string, version string, accuracy float64) error {
	v := r.FindVersion(id, version)
	if v == nil {
		return fmt.Errorf("version %s/%s not found", id, version)
	}
	v.ShadowAccuracy = accuracy
	return nil
}

// UpdateProductionMetrics updates production metrics for the active version.
func (r *PolicyLifecycleRegistry) UpdateProductionMetrics(id string, version string, accuracy float64, runs int, regret float64) error {
	v := r.FindVersion(id, version)
	if v == nil {
		return fmt.Errorf("version %s/%s not found", id, version)
	}
	v.ProductionAccuracy = accuracy
	v.ProductionRuns = runs
	v.RegretVsRules = regret
	return nil
}

// ───────────────────────────────────────────────────────────────
// Comparison
// ───────────────────────────────────────────────────────────────

// PolicyComparison holds a side-by-side comparison of two versions.
type PolicyComparison struct {
	VersionA PolicyVersionRecord
	VersionB PolicyVersionRecord

	// Differences
	AccuracyDelta       float64 // B - A (positive = B is better)
	RegretDelta         float64 // B - A (negative = B has less regret)
	TrainingSampleDelta int     // B - A
	Recommendation      string  // "promote_b", "keep_a", "needs_more_data"
}

// Compare produces a side-by-side comparison of two versions.
func (r *PolicyLifecycleRegistry) Compare(idA string, versionA string, idB string, versionB string) (*PolicyComparison, error) {
	a := r.FindVersion(idA, versionA)
	if a == nil {
		return nil, fmt.Errorf("version A (%s/%s) not found", idA, versionA)
	}
	b := r.FindVersion(idB, versionB)
	if b == nil {
		return nil, fmt.Errorf("version B (%s/%s) not found", idB, versionB)
	}

	comp := &PolicyComparison{
		VersionA:            *a,
		VersionB:            *b,
		AccuracyDelta:       bestAccuracy(b) - bestAccuracy(a),
		RegretDelta:         b.RegretVsRules - a.RegretVsRules,
		TrainingSampleDelta: b.TrainingSamples - a.TrainingSamples,
	}

	// Recommendation logic.
	if comp.AccuracyDelta > 0.05 && b.ProductionRuns >= 10 {
		comp.Recommendation = "promote_b"
	} else if comp.AccuracyDelta < -0.05 {
		comp.Recommendation = "keep_a"
	} else {
		comp.Recommendation = "needs_more_data"
	}

	return comp, nil
}

func bestAccuracy(v *PolicyVersionRecord) float64 {
	if v.ProductionAccuracy > 0 {
		return v.ProductionAccuracy
	}
	if v.ShadowAccuracy > 0 {
		return v.ShadowAccuracy
	}
	return v.OfflineAccuracy
}
