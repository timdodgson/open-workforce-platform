package ilp

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/cvrp"
	nrpilp "github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/ilp"
)

// BenchmarkConfig holds parameters for a CVRP ILP benchmark.
type BenchmarkConfig struct {
	Instance    string
	TimeLimit   time.Duration
	Parallel    bool
	MaxVehicles int    // 0 = auto (ceil(demand/capacity))
	OutputPath  string // where to write result JSON
}

// BenchmarkResult holds the outcome of a CVRP ILP solve.
type BenchmarkResult struct {
	Instance       string  `json:"instance"`
	Customers      int     `json:"customers"`
	Vehicles       int     `json:"vehicles"`
	Capacity       int     `json:"capacity"`
	Solver         string  `json:"solver"`
	Status         string  `json:"status"` // OPTIMAL, FEASIBLE, INFEASIBLE, TIMEOUT, ERROR
	Objective      int     `json:"objective"`
	LowerBound     int     `json:"lowerBound"`
	GapPercent     float64 `json:"gapPercent"`
	RuntimeSeconds float64 `json:"runtimeSeconds"`
	TimeLimit      int     `json:"timeLimit"`
	Variables      int     `json:"variables"`
	Constraints    int     `json:"constraints"`
	Notes          string  `json:"notes,omitempty"`
}

// RunBenchmark executes a complete CVRP ILP benchmark.
func RunBenchmark(ds *cvrp.Dataset, config BenchmarkConfig) (BenchmarkResult, error) {
	// Create temp directory.
	tmpDir, err := os.MkdirTemp("", "cvrp-ilp-*")
	if err != nil {
		return BenchmarkResult{}, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	modelPath := filepath.Join(tmpDir, "cvrp.lp")

	// Build model.
	info, err := BuildModel(ds, modelPath, config.MaxVehicles)
	if err != nil {
		return BenchmarkResult{}, fmt.Errorf("build model: %w", err)
	}

	// Check solver availability.
	solver := &nrpilp.HighsSolver{Parallel: config.Parallel}
	if !solver.Available() {
		return BenchmarkResult{
			Instance: ds.Name,
			Status:   "ERROR",
			Notes:    "HiGHS binary not found on PATH",
		}, fmt.Errorf("HiGHS not available")
	}

	// Solve.
	timeLimit := config.TimeLimit
	if timeLimit == 0 {
		timeLimit = 5 * time.Minute
	}

	solverOutput, err := solver.Solve(modelPath, timeLimit)

	result := BenchmarkResult{
		Instance:       ds.Name,
		Customers:      len(ds.Customers),
		Vehicles:       info.Vehicles,
		Capacity:       ds.Capacity,
		Solver:         "highs",
		Status:         solverOutput.Status,
		Objective:      int(solverOutput.Objective),
		LowerBound:     int(solverOutput.LowerBound),
		RuntimeSeconds: solverOutput.RuntimeSeconds,
		TimeLimit:      int(timeLimit.Seconds()),
		Variables:      info.Variables,
		Constraints:    info.Constraints,
	}

	// Calculate gap.
	if result.LowerBound > 0 && result.Objective > 0 {
		result.GapPercent = float64(result.Objective-result.LowerBound) / float64(result.LowerBound) * 100
	}

	if err != nil && result.Status == "ERROR" {
		result.Notes = err.Error()
	}

	return result, nil
}
