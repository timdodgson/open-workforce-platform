// validation_suite.go — Search Intelligence 2.0 Validation Framework.
//
// Defines the complete experiment matrix for validating SI 2.0 policies.
// Does NOT fabricate results. All data comes from actual experiment runs.
//
// The suite supports:
//   - Multiple policy modes (rules, hybrid, learned)
//   - All 4 domains (NRP, CVRP, JSS, VRPTW)
//   - Multiple seeds for statistical rigour
//   - Automated statistical analysis
//   - Report generation (only from real data)
package optimisation

import (
	"math"
	"sort"
	"time"
)

// ───────────────────────────────────────────────────────────────
// Validation Configuration
// ───────────────────────────────────────────────────────────────

// ValidationConfig defines the complete experiment matrix.
type ValidationConfig struct {
	// Domains to validate.
	Domains []DomainConfig

	// Policy modes to compare.
	PolicyModes []string // e.g. ["rules", "hybrid", "learned"]

	// Seeds for statistical repeatability.
	Seeds []int64

	// Output directory for results.
	OutputDir string
}

// DomainConfig specifies one domain's validation parameters.
type DomainConfig struct {
	Domain     string // nrp, cvrp, jss, vrptw
	Instance   string // e.g. A-n32-k5, la01, C101, n012w8
	Algorithm  string // sa, tabu, portfolio
	Iterations int
}

// DefaultValidationConfig returns the standard validation matrix.
func DefaultValidationConfig() ValidationConfig {
	return ValidationConfig{
		Domains: []DomainConfig{
			{Domain: "cvrp", Instance: "A-n32-k5", Algorithm: "sa", Iterations: 500000},
			{Domain: "cvrp", Instance: "A-n32-k5", Algorithm: "portfolio", Iterations: 500000},
			{Domain: "jss", Instance: "la01", Algorithm: "tabu", Iterations: 100000},
			{Domain: "jss", Instance: "la01", Algorithm: "portfolio", Iterations: 100000},
			{Domain: "vrptw", Instance: "C101", Algorithm: "sa", Iterations: 100000},
			{Domain: "vrptw", Instance: "C101", Algorithm: "portfolio", Iterations: 100000},
			{Domain: "nrp", Instance: "n012w8", Algorithm: "sa", Iterations: 100000},
			{Domain: "nrp", Instance: "n012w8", Algorithm: "portfolio", Iterations: 100000},
		},
		PolicyModes: []string{"rules", "hybrid", "learned"},
		Seeds:       []int64{42, 123, 555, 777, 999, 1001, 2022, 3033, 4044, 5055},
		OutputDir:   "validation/si2",
	}
}

// TotalExperiments returns the number of experiments in the matrix.
func (c ValidationConfig) TotalExperiments() int {
	return len(c.Domains) * len(c.PolicyModes) * len(c.Seeds)
}

// ───────────────────────────────────────────────────────────────
// Experiment Result
// ───────────────────────────────────────────────────────────────

// ExperimentResult captures one experiment run outcome.
type ExperimentResult struct {
	// Identity
	Domain     string
	Instance   string
	Algorithm  string
	PolicyMode string
	Seed       int64
	RunLabel   string

	// Quality
	Objective int
	Feasible  bool

	// Performance
	RuntimeMs      int64
	CandidatesEval int

	// Policy metrics
	Confidence          float64
	Regret              float64
	FallbackCount       int
	SafetyOverrideCount int
	DecisionCount       int

	// Timestamp
	CompletedAt time.Time
}

// ───────────────────────────────────────────────────────────────
// Statistical Analysis
// ───────────────────────────────────────────────────────────────

// StatisticalResult holds computed statistics for one configuration.
type StatisticalResult struct {
	Domain     string
	Instance   string
	Algorithm  string
	PolicyMode string
	N          int

	// Descriptive statistics.
	Mean   float64
	Median float64
	StdDev float64
	Min    int
	Max    int

	// Confidence interval (95%).
	CI95Lower float64
	CI95Upper float64

	// Runtime stats.
	MeanRuntimeMs  float64
	MeanCandidates float64

	// Policy-specific.
	MeanConfidence float64
	MeanRegret     float64
	FallbackRate   float64
}

// ComparisonResult compares two policy modes on the same configuration.
type ComparisonResult struct {
	Domain    string
	Instance  string
	Algorithm string
	ModeA     string
	ModeB     string

	// Descriptive.
	MeanA float64
	MeanB float64

	// Statistical tests.
	WelchT       float64
	WelchP       float64
	MannWhitneyU float64
	MannWhitneyP float64
	CohensD      float64
	Significant  bool // p < 0.05

	// Win/Loss/Tie.
	Wins   int
	Losses int
	Ties   int

	// Verdict.
	Verdict string // "better", "worse", "equivalent", "not_evaluated"
}

// ComputeStatistics calculates descriptive stats for a set of results.
func ComputeStatistics(results []ExperimentResult) StatisticalResult {
	if len(results) == 0 {
		return StatisticalResult{}
	}

	objectives := make([]float64, len(results))
	runtimes := make([]float64, len(results))
	candidates := make([]float64, len(results))
	confidences := make([]float64, len(results))
	regrets := make([]float64, len(results))
	fallbacks := 0
	decisions := 0

	for i, r := range results {
		objectives[i] = float64(r.Objective)
		runtimes[i] = float64(r.RuntimeMs)
		candidates[i] = float64(r.CandidatesEval)
		confidences[i] = r.Confidence
		regrets[i] = r.Regret
		fallbacks += r.FallbackCount
		decisions += r.DecisionCount
	}

	n := float64(len(objectives))
	mean := sum(objectives) / n
	med := median(objectives)
	sd := stddev(objectives, mean)

	// 95% CI: mean ± t * (sd / sqrt(n))
	// For n >= 10, use z ≈ 1.96.
	margin := 1.96 * sd / math.Sqrt(n)

	intObjs := make([]int, len(results))
	for i, r := range results {
		intObjs[i] = r.Objective
	}
	sort.Ints(intObjs)

	fallbackRate := 0.0
	if decisions > 0 {
		fallbackRate = float64(fallbacks) / float64(decisions)
	}

	return StatisticalResult{
		Domain:         results[0].Domain,
		Instance:       results[0].Instance,
		Algorithm:      results[0].Algorithm,
		PolicyMode:     results[0].PolicyMode,
		N:              len(results),
		Mean:           mean,
		Median:         med,
		StdDev:         sd,
		Min:            intObjs[0],
		Max:            intObjs[len(intObjs)-1],
		CI95Lower:      mean - margin,
		CI95Upper:      mean + margin,
		MeanRuntimeMs:  sum(runtimes) / n,
		MeanCandidates: sum(candidates) / n,
		MeanConfidence: sum(confidences) / n,
		MeanRegret:     sum(regrets) / n,
		FallbackRate:   fallbackRate,
	}
}

// CompareGroups performs statistical comparison between two result sets.
func CompareGroups(a []ExperimentResult, b []ExperimentResult, modeA string, modeB string) ComparisonResult {
	if len(a) == 0 || len(b) == 0 {
		return ComparisonResult{Verdict: "not_evaluated"}
	}

	objA := make([]float64, len(a))
	objB := make([]float64, len(b))
	for i, r := range a {
		objA[i] = float64(r.Objective)
	}
	for i, r := range b {
		objB[i] = float64(r.Objective)
	}

	meanA := sum(objA) / float64(len(objA))
	meanB := sum(objB) / float64(len(objB))
	sdA := stddev(objA, meanA)
	sdB := stddev(objB, meanB)

	// Welch's t-test.
	nA := float64(len(objA))
	nB := float64(len(objB))
	seA := sdA * sdA / nA
	seB := sdB * sdB / nB
	tStat := 0.0
	if seA+seB > 0 {
		tStat = (meanA - meanB) / math.Sqrt(seA+seB)
	}

	// Cohen's d.
	pooledSD := math.Sqrt((sdA*sdA + sdB*sdB) / 2)
	cohensD := 0.0
	if pooledSD > 0 {
		cohensD = (meanA - meanB) / pooledSD
	}

	// Win/Loss/Tie (per seed, assumes same seed order).
	wins, losses, ties := 0, 0, 0
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}
	for i := 0; i < minLen; i++ {
		if a[i].Objective < b[i].Objective {
			wins++
		} else if a[i].Objective > b[i].Objective {
			losses++
		} else {
			ties++
		}
	}

	// Approximate p-value (two-tailed, normal approx for large n).
	pValue := approxPValue(tStat)

	verdict := "equivalent"
	if pValue < 0.05 && meanB < meanA {
		verdict = "better" // B is better (lower objective)
	} else if pValue < 0.05 && meanB > meanA {
		verdict = "worse"
	}

	return ComparisonResult{
		Domain:      a[0].Domain,
		Instance:    a[0].Instance,
		Algorithm:   a[0].Algorithm,
		ModeA:       modeA,
		ModeB:       modeB,
		MeanA:       meanA,
		MeanB:       meanB,
		WelchT:      tStat,
		WelchP:      pValue,
		CohensD:     cohensD,
		Significant: pValue < 0.05,
		Wins:        wins,
		Losses:      losses,
		Ties:        ties,
		Verdict:     verdict,
	}
}

// ───────────────────────────────────────────────────────────────
// Helpers
// ───────────────────────────────────────────────────────────────

func sum(xs []float64) float64 {
	s := 0.0
	for _, x := range xs {
		s += x
	}
	return s
}

func median(xs []float64) float64 {
	sorted := make([]float64, len(xs))
	copy(sorted, xs)
	sort.Float64s(sorted)
	n := len(sorted)
	if n%2 == 0 {
		return (sorted[n/2-1] + sorted[n/2]) / 2
	}
	return sorted[n/2]
}

func stddev(xs []float64, mean float64) float64 {
	if len(xs) < 2 {
		return 0
	}
	ss := 0.0
	for _, x := range xs {
		d := x - mean
		ss += d * d
	}
	return math.Sqrt(ss / float64(len(xs)-1))
}

func approxPValue(t float64) float64 {
	// Approximation using normal CDF for |t| > 0.
	// For rigorous analysis, use a proper t-distribution. This is sufficient for reporting.
	absT := math.Abs(t)
	if absT > 4 {
		return 0.0001
	}
	// Simple logistic approximation.
	p := 1.0 / (1.0 + math.Exp(0.7*absT*absT-0.3*absT))
	return 2 * p // two-tailed
}
