package ilp

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/jobshop"
	nrpilp "github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/ilp"
)

// BenchmarkConfig holds parameters for a JSS ILP benchmark.
type BenchmarkConfig struct {
	Instance   string
	TimeLimit  time.Duration
	Parallel   bool
	OutputPath string
}

// BenchmarkResult holds the outcome of a JSS ILP solve.
type BenchmarkResult struct {
	Instance       string  `json:"instance"`
	Jobs           int     `json:"jobs"`
	Machines       int     `json:"machines"`
	Operations     int     `json:"operations"`
	Solver         string  `json:"solver"`
	Status         string  `json:"status"`
	Objective      int     `json:"objective"`
	LowerBound     int     `json:"lowerBound"`
	GapPercent     float64 `json:"gapPercent"`
	RuntimeSeconds float64 `json:"runtimeSeconds"`
	TimeLimit      int     `json:"timeLimit"`
	Variables      int     `json:"variables"`
	Constraints    int     `json:"constraints"`
	Notes          string  `json:"notes,omitempty"`
	SolutionJSON   []byte  `json:"-"`
}

// RunBenchmark executes a complete JSS ILP benchmark.
func RunBenchmark(ds *jobshop.Dataset, config BenchmarkConfig) (BenchmarkResult, error) {
	tmpDir, err := os.MkdirTemp("", "jss-ilp-*")
	if err != nil {
		return BenchmarkResult{}, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	modelPath := filepath.Join(tmpDir, "jss.lp")

	info, err := BuildModel(ds, modelPath)
	if err != nil {
		return BenchmarkResult{}, fmt.Errorf("build model: %w", err)
	}

	solver := &nrpilp.HighsSolver{Parallel: config.Parallel}
	if !solver.Available() {
		return BenchmarkResult{
			Instance: ds.Name,
			Status:   "ERROR",
			Notes:    "HiGHS binary not found on PATH",
		}, fmt.Errorf("HiGHS not available")
	}

	timeLimit := config.TimeLimit
	if timeLimit == 0 {
		timeLimit = 5 * time.Minute
	}

	solverOutput, err := solver.Solve(modelPath, timeLimit)

	result := BenchmarkResult{
		Instance:       ds.Name,
		Jobs:           ds.Jobs,
		Machines:       ds.Machines,
		Operations:     info.Operations,
		Solver:         "highs",
		Status:         solverOutput.Status,
		Objective:      int(solverOutput.Objective),
		LowerBound:     int(solverOutput.LowerBound),
		RuntimeSeconds: solverOutput.RuntimeSeconds,
		TimeLimit:      int(timeLimit.Seconds()),
		Variables:      info.Variables,
		Constraints:    info.Constraints,
	}

	if result.LowerBound > 0 && result.Objective > 0 {
		result.GapPercent = float64(result.Objective-result.LowerBound) / float64(result.LowerBound) * 100
	}

	if result.Status == "OPTIMAL" || result.Status == "FEASIBLE" {
		if solJSON, solErr := ExtractSolution(ds, solverOutput); solErr == nil {
			result.SolutionJSON = solJSON
		} else if result.Notes == "" {
			result.Notes = "solution extract: " + solErr.Error()
		}
	}

	if err != nil && result.Status == "ERROR" {
		result.Notes = err.Error()
	}

	return result, nil
}
