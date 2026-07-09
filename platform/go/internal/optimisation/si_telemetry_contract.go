// SI telemetry contract — every labelled run emits the same CSV set (header at minimum).
package optimisation

import (
	"os"
	"path/filepath"
	"strings"
)

// SITelemetryCSV lists Search Intelligence telemetry files expected per run.
var SITelemetryCSV = []string{
	"worker_learning.csv",
	"worker_decisions.csv",
	"worker_assist.csv",
	"generic_search_assist.csv",
	"portfolio_assist.csv",
	"policy_decisions.csv",
	"policy_evaluation.csv",
	"counterfactual_learning.csv",
}

// SITelemetryHeaders maps contract filenames to CSV header rows.
var SITelemetryHeaders = map[string]string{
	"worker_learning.csv":         workerLearningHeader(),
	"worker_decisions.csv":        workerDecisionsHeader(),
	"worker_assist.csv":           workerAssistHeader(),
	"generic_search_assist.csv":   searchAssistHeader(),
	"portfolio_assist.csv":        portfolioAssistHeader(),
	"policy_decisions.csv":        policyDecisionsHeader(),
	"policy_evaluation.csv":       policyEvaluationHeader(),
	"counterfactual_learning.csv": counterfactualHeader(),
}

func workerLearningHeader() string {
	// Mirror inrc2.WorkerLearningCSVHeader without importing inrc2.
	return strings.Join([]string{
		"problem_type", "instance", "algorithm", "run_seed",
		"week", "phase", "depth", "parent_worker_id", "family_id",
		"beam_rank", "beam_score", "entropy", "diversity", "beam_health",
		"temperature", "lahc_length", "tabu_tenure", "iterations_alloc",
		"global_best", "parent_objective", "distance_from_best",
		"plateau_length", "recent_improv_rate", "worker_count", "active_families",
		"improved", "produced_global_best", "improvement_amount",
		"final_objective", "runtime_ms", "candidates_eval",
		"accepted", "rejected", "plateau_count", "branches_spawned",
		"roi", "improv_per_cpu", "improv_per_100k",
	}, ",")
}

func workerDecisionsHeader() string {
	return strings.Join([]string{
		"worker_id", "week", "depth", "algorithm",
		"parent_objective", "global_best", "distance_from_best",
		"beam_rank", "entropy", "beam_health", "recent_improv_rate", "allocated_iters",
		"recommendation", "confidence", "reason_codes",
		"suggested_algorithm", "suggested_budget",
		"improved", "produced_global_best", "improvement_amount",
		"final_objective", "runtime_ms", "roi",
	}, ",")
}

func workerAssistHeader() string {
	return strings.Join([]string{
		"worker_id", "week", "depth", "algorithm",
		"parent_objective", "global_best", "distance_from_best",
		"recommendation", "confidence", "reason_codes",
		"suggested_algorithm", "suggested_budget",
		"safety_triggered", "safety_rule",
		"outcome", "final_action", "final_budget", "final_algorithm",
		"improved", "produced_global_best", "improvement_amount",
		"final_objective", "runtime_ms",
	}, ",")
}

func searchAssistHeader() string {
	return strings.Join([]string{
		"algorithm", "checkpoint", "candidates", "iterations_total",
		"current_penalty", "best_penalty", "initial_penalty",
		"temperature", "plateau_length", "improvement_rate",
		"recommended_action", "confidence", "reasons",
		"safety_triggered", "safety_rule",
		"accepted", "final_action",
		"final_best_penalty", "total_candidates", "runtime_ms",
	}, ",")
}

func portfolioAssistHeader() string {
	return strings.Join([]string{
		"domain", "instance", "strategy", "seed",
		"original_budget", "recommended_budget", "final_budget",
		"recommendation", "confidence", "reason_codes",
		"accepted", "safety_rejected", "safety_rule",
		"result_objective", "strategy_won", "runtime_ms",
	}, ",")
}

func policyDecisionsHeader() string {
	return "checkpoint,candidates,policy_mode,policy_used,action,confidence,fallback_reason,safety_override\n"
}

func policyEvaluationHeader() string {
	return strings.Join([]string{
		"timestamp", "run_id", "decision_type", "domain", "instance", "algorithm",
		"policy_id", "policy_version", "policy_type", "action", "confidence",
		"expected_improvement", "actual_improvement", "prediction_error",
		"correct", "absolute_error", "squared_error", "regret",
	}, ",")
}

func counterfactualHeader() string {
	return strings.Join([]string{
		"timestamp", "run_id", "decision_type",
		"domain", "instance", "algorithm",
		"actual_action", "actual_confidence", "policy_id", "policy_version", "policy_type",
		"expected_value",
		"cf1_action", "cf1_source", "cf1_expected",
		"cf2_action", "cf2_source", "cf2_expected",
		"cf3_action", "cf3_source", "cf3_expected",
		"actual_outcome", "outcome_metric",
		"regret", "best_alternative",
	}, ",")
}

// WriteTelemetryCSVHeaderOnly writes a header row when the file is absent.
func WriteTelemetryCSVHeaderOnly(path, header string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	h := header
	if !strings.HasSuffix(h, "\n") {
		h += "\n"
	}
	return os.WriteFile(path, []byte(h), 0644)
}

// EnsureSITelemetryContract writes header-only stubs for any missing contract CSVs.
// skip maps filename → true to skip files already written with data.
func EnsureSITelemetryContract(outputDir string, skip map[string]bool) {
	if outputDir == "" {
		return
	}
	if skip == nil {
		skip = map[string]bool{}
	}
	for _, name := range SITelemetryCSV {
		if skip[name] {
			continue
		}
		header, ok := SITelemetryHeaders[name]
		if !ok {
			continue
		}
		_ = WriteTelemetryCSVHeaderOnly(filepath.Join(outputDir, name), header)
	}
}

// MinLearnedPolicyAgreement is the minimum offline agreement required to promote learned policies.
const MinLearnedPolicyAgreement = 0.80

// CanPromoteLearnedPolicy returns whether a learned policy meets the promotion gate.
func CanPromoteLearnedPolicy(agreement float64) bool {
	return agreement >= MinLearnedPolicyAgreement
}
