package policy

import (
	"fmt"
	"os"
)

// PolicySearchDecision records one policy-aware search checkpoint decision.
type PolicySearchDecision struct {
	Checkpoint     int
	Candidates     int
	PolicyMode     string
	PolicyUsed     string
	Action         string
	Confidence     float64
	FallbackReason string
	SafetyOverride bool
}

// PolicySearchConfig configures policy-based search hooks.
type PolicySearchConfig struct {
	PolicyMode          string
	PolicyDir           string
	Domain              string
	Instance            string
	ConfidenceThreshold float64
}

// WritePolicyDecisionsCSV writes the policy decisions to a CSV file.
func WritePolicyDecisionsCSV(path string, decisions []PolicySearchDecision) error {
	if len(decisions) == 0 {
		return nil
	}

	header := "checkpoint,candidates,policy_mode,policy_used,action,confidence,fallback_reason,safety_override\n"
	var rows string
	for _, d := range decisions {
		safety := "0"
		if d.SafetyOverride {
			safety = "1"
		}
		rows += fmt.Sprintf("%d,%d,%s,%s,%s,%.4f,%s,%s\n",
			d.Checkpoint, d.Candidates, d.PolicyMode, d.PolicyUsed,
			d.Action, d.Confidence, d.FallbackReason, safety)
	}

	return os.WriteFile(path, []byte(header+rows), 0644)
}
