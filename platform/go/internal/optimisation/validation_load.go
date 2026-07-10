package optimisation

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// LoadExperimentResultsFromRunsDir reads completed run directories and builds experiment results.
// Only directories containing run.json are included. Optional prefix filters run labels.
func LoadExperimentResultsFromRunsDir(runsDir string, labelPrefix string) ([]ExperimentResult, error) {
	entries, err := os.ReadDir(runsDir)
	if err != nil {
		return nil, fmt.Errorf("read runs dir: %w", err)
	}

	var results []ExperimentResult
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		label := e.Name()
		if labelPrefix != "" && !strings.HasPrefix(label, labelPrefix) {
			continue
		}
		runPath := filepath.Join(runsDir, label, "run.json")
		data, err := os.ReadFile(runPath)
		if err != nil {
			continue
		}
		var meta map[string]interface{}
		if err := json.Unmarshal(data, &meta); err != nil {
			continue
		}
		if r, ok := experimentResultFromRunMeta(label, meta); ok {
			results = append(results, r)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Domain != results[j].Domain {
			return results[i].Domain < results[j].Domain
		}
		if results[i].PolicyMode != results[j].PolicyMode {
			return results[i].PolicyMode < results[j].PolicyMode
		}
		return results[i].Seed < results[j].Seed
	})

	return results, nil
}

func experimentResultFromRunMeta(label string, meta map[string]interface{}) (ExperimentResult, bool) {
	domain := stringField(meta, "problemType", "domain")
	if domain == "" {
		domain = inferDomainFromLabel(label)
	}
	policyMode := stringField(meta, "policyMode", "policy_mode")
	if policyMode == "" {
		policyMode = inferPolicyFromLabel(label)
	}
	algorithm := stringField(meta, "mode", "algorithm", "winnerStrategy")
	if algorithm == "" {
		algorithm = inferAlgorithmFromLabel(label)
	}
	instance := stringField(meta, "instance")
	objective := intField(meta, "bestObjective", "bestDistance", "bestPenalty", "objective")
	if objective == 0 {
		objective = intField(meta, "finalPenalty")
	}
	seed := int64Field(meta, "seed")
	if seed == 0 {
		seed = inferSeedFromLabel(label)
	}

	if domain == "" || policyMode == "" {
		return ExperimentResult{}, false
	}

	feasible := true
	if v, ok := meta["feasible"].(bool); ok {
		feasible = v
	}

	return ExperimentResult{
		Domain:         domain,
		Instance:       instance,
		Algorithm:      algorithm,
		PolicyMode:     policyMode,
		Seed:           seed,
		RunLabel:       label,
		Objective:      objective,
		Feasible:       feasible,
		RuntimeMs:      int64Field(meta, "runtimeMs", "runtime_ms"),
		CandidatesEval: int(int64Field(meta, "candidates", "candidatesEval")),
		CompletedAt:    time.Now(),
	}, true
}

func stringField(meta map[string]interface{}, keys ...string) string {
	for _, k := range keys {
		if v, ok := meta[k]; ok {
			switch s := v.(type) {
			case string:
				if s != "" {
					return s
				}
			}
		}
	}
	return ""
}

func intField(meta map[string]interface{}, keys ...string) int {
	for _, k := range keys {
		if v, ok := meta[k]; ok {
			switch n := v.(type) {
			case float64:
				return int(n)
			case int:
				return n
			case json.Number:
				i, _ := n.Int64()
				return int(i)
			}
		}
	}
	return 0
}

func int64Field(meta map[string]interface{}, keys ...string) int64 {
	for _, k := range keys {
		if v, ok := meta[k]; ok {
			switch n := v.(type) {
			case float64:
				return int64(n)
			case int:
				return int64(n)
			case int64:
				return n
			case json.Number:
				i, _ := n.Int64()
				return i
			}
		}
	}
	return 0
}

func inferDomainFromLabel(label string) string {
	parts := strings.Split(label, "-")
	if len(parts) < 2 {
		return ""
	}
	switch parts[1] {
	case "cvrp", "jss", "vrptw", "nrp":
		return parts[1]
	case "deep":
		if len(parts) >= 3 {
			return parts[2]
		}
	}
	return ""
}

func inferPolicyFromLabel(label string) string {
	for _, p := range []string{"rules", "hybrid", "learned"} {
		if strings.Contains(label, "-"+p+"-") || strings.HasSuffix(label, "-"+p) {
			return p
		}
	}
	return ""
}

func inferAlgorithmFromLabel(label string) string {
	for _, a := range []string{"portfolio", "tabu", "lahc", "sa"} {
		if strings.Contains(label, "-"+a+"-") || strings.HasSuffix(label, "-"+a) {
			return a
		}
	}
	return ""
}

func inferSeedFromLabel(label string) int64 {
	parts := strings.Split(label, "-")
	for _, p := range parts {
		if strings.HasPrefix(p, "s") && len(p) > 1 {
			if n, err := strconv.ParseInt(p[1:], 10, 64); err == nil {
				return n
			}
		}
	}
	return 0
}

// GroupExperimentResults groups results by domain, instance, algorithm, and policy mode.
func GroupExperimentResults(results []ExperimentResult) map[string][]ExperimentResult {
	groups := make(map[string][]ExperimentResult)
	for _, r := range results {
		key := fmt.Sprintf("%s|%s|%s|%s", r.Domain, r.Instance, r.Algorithm, r.PolicyMode)
		groups[key] = append(groups[key], r)
	}
	return groups
}

// ComparePolicyModes compares rules vs hybrid and rules vs learned for each domain config.
func ComparePolicyModes(results []ExperimentResult) []ComparisonResult {
	byConfig := make(map[string]map[string][]ExperimentResult)
	for _, r := range results {
		cfgKey := fmt.Sprintf("%s|%s|%s", r.Domain, r.Instance, r.Algorithm)
		if byConfig[cfgKey] == nil {
			byConfig[cfgKey] = make(map[string][]ExperimentResult)
		}
		byConfig[cfgKey][r.PolicyMode] = append(byConfig[cfgKey][r.PolicyMode], r)
	}

	var comparisons []ComparisonResult
	for _, modes := range byConfig {
		if rules, ok := modes["rules"]; ok {
			if hybrid, ok := modes["hybrid"]; ok {
				comparisons = append(comparisons, CompareGroups(rules, hybrid, "rules", "hybrid"))
			}
			if learned, ok := modes["learned"]; ok {
				comparisons = append(comparisons, CompareGroups(rules, learned, "rules", "learned"))
			}
		}
	}
	return comparisons
}
