package optimisation

import "time"

// PolicyEvaluationInput configures post-run policy evaluation CSV emission.
type PolicyEvaluationInput struct {
	RunID          string
	Domain         string
	Instance       string
	Algorithm      string
	InitialPenalty int
	BestPenalty    int
	Decisions      []PolicySearchDecision
}

// BuildPolicyEvaluationRecords converts SI 2.0 checkpoint decisions into evaluation rows.
func BuildPolicyEvaluationRecords(in PolicyEvaluationInput) []PolicyEvaluationRecord {
	if len(in.Decisions) == 0 {
		return nil
	}

	totalImprovement := float64(in.InitialPenalty - in.BestPenalty)
	if totalImprovement < 0 {
		totalImprovement = 0
	}

	records := make([]PolicyEvaluationRecord, 0, len(in.Decisions))
	now := time.Now()

	for _, d := range in.Decisions {
		if d.PolicyUsed == "" || (d.PolicyUsed == "rule" && d.FallbackReason == "no_learned_assessment") {
			continue
		}

		decisionType := "stagnation"
		if d.Action == "restart" {
			decisionType = "restart"
		}

		policyType := d.PolicyUsed
		if policyType == "hybrid_learned" || policyType == "restart" {
			policyType = "hybrid"
		}

		expected := d.Confidence * totalImprovement
		actual := 0.0
		if d.Action == "early_stop" {
			actual = 0
		} else if d.Action == "restart" {
			actual = totalImprovement * 0.1
		}

		correct := d.Action == "early_stop" && totalImprovement > 0 && d.Confidence >= 0.6
		if d.Action == "restart" {
			correct = d.Confidence >= 0.6
		}

		records = append(records, PolicyEvaluationRecord{
			Timestamp:           now,
			RunID:               in.RunID,
			DecisionType:        decisionType,
			Domain:              in.Domain,
			Instance:            in.Instance,
			Algorithm:           in.Algorithm,
			PolicyID:            decisionType + "_" + in.Domain,
			PolicyVersion:       "1.0.0",
			PolicyType:          policyType,
			Action:              d.Action,
			Confidence:          d.Confidence,
			ExpectedImprovement: expected,
			ActualImprovement:   actual,
			Correct:             correct,
		})
	}

	return records
}

// WritePolicyEvaluationCSV writes evaluation records for a completed run.
func WritePolicyEvaluationCSV(dir string, in PolicyEvaluationInput) error {
	records := BuildPolicyEvaluationRecords(in)
	if len(records) == 0 {
		return nil
	}
	evaluator := NewPolicyEvaluator()
	for _, rec := range records {
		evaluator.Record(rec)
	}
	return evaluator.WriteCSV(dir)
}
