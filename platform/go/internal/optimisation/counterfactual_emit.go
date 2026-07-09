package optimisation

import "time"

// CounterfactualEmitInput configures counterfactual CSV emission from policy decisions.
type CounterfactualEmitInput struct {
	RunID          string
	Domain         string
	Instance       string
	Algorithm      string
	InitialPenalty int
	BestPenalty    int
	Decisions      []PolicySearchDecision
}

// BuildCounterfactualRecords converts SI 2.0 policy decisions into counterfactual rows.
func BuildCounterfactualRecords(in CounterfactualEmitInput) []CounterfactualRecord {
	if len(in.Decisions) == 0 {
		return nil
	}

	totalImprovement := float64(in.InitialPenalty - in.BestPenalty)
	if totalImprovement < 0 {
		totalImprovement = 0
	}

	records := make([]CounterfactualRecord, 0, len(in.Decisions))
	now := time.Now()

	for _, d := range in.Decisions {
		if d.PolicyUsed == "" {
			continue
		}

		decisionType := "search"
		altAction := "continue"
		if d.Action == "early_stop" {
			altAction = "restart"
		} else if d.Action == "restart" {
			decisionType = "restart"
			altAction = "continue"
		}

		actualOutcome := totalImprovement
		if d.Action == "early_stop" {
			actualOutcome = 0
		}

		records = append(records, CounterfactualRecord{
			Timestamp:        now,
			RunID:            in.RunID,
			DecisionType:     decisionType,
			Domain:           in.Domain,
			Instance:         in.Instance,
			Algorithm:        in.Algorithm,
			ActualAction:     d.Action,
			ActualConfidence: d.Confidence,
			PolicyID:         decisionType + "_" + in.Domain,
			PolicyVersion:    "1.0.0",
			PolicyType:       d.PolicyUsed,
			ExpectedValue:    d.Confidence * totalImprovement,
			CounterfactualActions: []CounterfactualAction{
				{Action: altAction, Source: "rule", ExpectedValue: totalImprovement * 0.5},
				{Action: "continue", Source: "opposite", ExpectedValue: 0},
			},
			ActualOutcome:   actualOutcome,
			OutcomeMetric:   "objective_improvement",
			BestAlternative: altAction,
		})
	}

	return records
}

// WriteCounterfactualFromPolicyDecisions writes counterfactual_learning.csv for a run.
func WriteCounterfactualFromPolicyDecisions(dir string, in CounterfactualEmitInput) error {
	records := BuildCounterfactualRecords(in)
	if len(records) == 0 {
		return nil
	}
	cr := NewCounterfactualRecorder(dir)
	defer cr.Close()
	for _, rec := range records {
		if err := cr.Record(rec); err != nil {
			return err
		}
	}
	return nil
}
