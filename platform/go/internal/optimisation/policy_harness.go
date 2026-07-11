package optimisation

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"time"
)

// PolicyHarnessReport is the Step 1 measurement artifact (rules vs hybrid vs learned).
type PolicyHarnessReport struct {
	GeneratedAt   string                    `json:"generatedAt"`
	Step          int                       `json:"step"`
	MLMaturity    float64                   `json:"mlMaturity"`
	TotalRuns     int                       `json:"totalRuns"`
	Prefix        string                    `json:"prefix"`
	ModeSummaries []PolicyModeSummary       `json:"modeSummaries"`
	Comparisons   []PolicyHarnessComparison `json:"comparisons"`
	Gates         PolicyHarnessGates        `json:"gates"`
}

type PolicyModeSummary struct {
	PolicyMode      string  `json:"policyMode"`
	N               int     `json:"n"`
	MeanObjective   float64 `json:"meanObjective"`
	MeanRuntimeMs   float64 `json:"meanRuntimeMs"`
	FeasibilityRate float64 `json:"feasibilityRate"`
}

type PolicyHarnessComparison struct {
	Domain          string  `json:"domain"`
	Instance        string  `json:"instance"`
	Algorithm       string  `json:"algorithm"`
	ModeA           string  `json:"modeA"`
	ModeB           string  `json:"modeB"`
	MeanObjectiveA  float64 `json:"meanObjectiveA"`
	MeanObjectiveB  float64 `json:"meanObjectiveB"`
	ObjectiveDelta  float64 `json:"objectiveDelta"`
	MeanRuntimeA    float64 `json:"meanRuntimeA"`
	MeanRuntimeB    float64 `json:"meanRuntimeB"`
	RuntimeSavedPct float64 `json:"runtimeSavedPct"`
	WelchP          float64 `json:"welchP"`
	CohensD         float64 `json:"cohensD"`
	Verdict         string  `json:"verdict"`
	RuntimeVerdict  string  `json:"runtimeVerdict"`
	ROI             float64 `json:"roi"`
}

type PolicyHarnessGates struct {
	Step1HarnessOK   bool `json:"step1HarnessOk"`
	Step2QualityWins int  `json:"step2QualityWins"`
	Step2RuntimeWins int  `json:"step2RuntimeWins"`
	Step2PromoteOK   bool `json:"step2PromoteOk"`
}

// BuildPolicyHarnessReport compares policy modes on completed val-* runs.
func BuildPolicyHarnessReport(results []ExperimentResult, prefix string, mlMaturity float64) PolicyHarnessReport {
	report := PolicyHarnessReport{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Step:        1,
		MLMaturity:  mlMaturity,
		TotalRuns:   len(results),
		Prefix:      prefix,
		Gates:       PolicyHarnessGates{Step1HarnessOK: len(results) > 0},
	}

	byMode := map[string][]ExperimentResult{}
	for _, r := range results {
		byMode[r.PolicyMode] = append(byMode[r.PolicyMode], r)
	}
	for mode, group := range byMode {
		stats := ComputeStatistics(group)
		feasible := 0
		for _, r := range group {
			if r.Feasible {
				feasible++
			}
		}
		rate := 0.0
		if len(group) > 0 {
			rate = float64(feasible) / float64(len(group))
		}
		report.ModeSummaries = append(report.ModeSummaries, PolicyModeSummary{
			PolicyMode:      mode,
			N:               stats.N,
			MeanObjective:   stats.Mean,
			MeanRuntimeMs:   stats.MeanRuntimeMs,
			FeasibilityRate: rate,
		})
	}
	sort.Slice(report.ModeSummaries, func(i, j int) bool {
		return report.ModeSummaries[i].PolicyMode < report.ModeSummaries[j].PolicyMode
	})

	objComparisons := ComparePolicyModes(results)
	runtimeMap := compareRuntimeByConfig(results)

	for _, c := range objComparisons {
		if c.Verdict == "not_evaluated" {
			continue
		}
		key := fmt.Sprintf("%s|%s|%s|%s|%s", c.Domain, c.Instance, c.Algorithm, c.ModeA, c.ModeB)
		rt := runtimeMap[key]
		roi := computeROI(c, rt)
		runtimeVerdict := "equivalent"
		if rt.RuntimeSavedPct > 2 {
			runtimeVerdict = "faster"
		} else if rt.RuntimeSavedPct < -2 {
			runtimeVerdict = "slower"
		}

		entry := PolicyHarnessComparison{
			Domain:          c.Domain,
			Instance:        c.Instance,
			Algorithm:       c.Algorithm,
			ModeA:           c.ModeA,
			ModeB:           c.ModeB,
			MeanObjectiveA:  c.MeanA,
			MeanObjectiveB:  c.MeanB,
			ObjectiveDelta:  c.MeanB - c.MeanA,
			MeanRuntimeA:    rt.MeanA,
			MeanRuntimeB:    rt.MeanB,
			RuntimeSavedPct: rt.RuntimeSavedPct,
			WelchP:          c.WelchP,
			CohensD:         c.CohensD,
			Verdict:         c.Verdict,
			RuntimeVerdict:  runtimeVerdict,
			ROI:             roi,
		}
		report.Comparisons = append(report.Comparisons, entry)

		if c.ModeB == "hybrid" || c.ModeB == "learned" {
			if c.Verdict == "better" || c.Verdict == "equivalent" {
				report.Gates.Step2QualityWins++
			}
			if runtimeVerdict == "faster" && c.Verdict != "worse" {
				report.Gates.Step2RuntimeWins++
			}
		}
	}
	sort.Slice(report.Comparisons, func(i, j int) bool {
		a, b := report.Comparisons[i], report.Comparisons[j]
		if a.Domain != b.Domain {
			return a.Domain < b.Domain
		}
		return a.Algorithm < b.Algorithm
	})

	report.Gates.Step2PromoteOK = report.Gates.Step2QualityWins >= 2 || report.Gates.Step2RuntimeWins >= 2
	return report
}

type runtimeComparison struct {
	MeanA           float64
	MeanB           float64
	RuntimeSavedPct float64
}

func compareRuntimeByConfig(results []ExperimentResult) map[string]runtimeComparison {
	byConfig := map[string]map[string][]ExperimentResult{}
	for _, r := range results {
		cfgKey := fmt.Sprintf("%s|%s|%s", r.Domain, r.Instance, r.Algorithm)
		if byConfig[cfgKey] == nil {
			byConfig[cfgKey] = map[string][]ExperimentResult{}
		}
		byConfig[cfgKey][r.PolicyMode] = append(byConfig[cfgKey][r.PolicyMode], r)
	}

	out := map[string]runtimeComparison{}
	for _, modes := range byConfig {
		rules, okR := modes["rules"]
		if !okR {
			continue
		}
		for modeB, groupB := range modes {
			if modeB == "rules" {
				continue
			}
			meanA := meanRuntime(rules)
			meanB := meanRuntime(groupB)
			saved := 0.0
			if meanA > 0 {
				saved = (meanA - meanB) / meanA * 100
			}
			if len(rules) == 0 {
				continue
			}
			key := fmt.Sprintf("%s|%s|%s|rules|%s", rules[0].Domain, rules[0].Instance, rules[0].Algorithm, modeB)
			out[key] = runtimeComparison{MeanA: meanA, MeanB: meanB, RuntimeSavedPct: saved}
		}
	}
	return out
}

func meanRuntime(results []ExperimentResult) float64 {
	if len(results) == 0 {
		return 0
	}
	sum := 0.0
	for _, r := range results {
		sum += float64(r.RuntimeMs)
	}
	return sum / float64(len(results))
}

func computeROI(c ComparisonResult, rt runtimeComparison) float64 {
	qualityGainPct := 0.0
	if c.MeanA != 0 {
		qualityGainPct = (c.MeanA - c.MeanB) / math.Abs(c.MeanA) * 100
	}
	if c.Verdict == "worse" {
		qualityGainPct = -math.Abs(qualityGainPct)
	}
	complexityCost := 1.0 // Step 2 default cost unit
	return (qualityGainPct + rt.RuntimeSavedPct) / complexityCost
}

// MarshalPolicyHarness returns indented JSON for reports.
func MarshalPolicyHarness(report PolicyHarnessReport) ([]byte, error) {
	return json.MarshalIndent(report, "", "  ")
}
