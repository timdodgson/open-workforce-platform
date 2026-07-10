package ilp

import (
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/cvrp"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/runoutput"
)

// FinalizeOptions configures run artifact output for a CVRP ILP benchmark.
type FinalizeOptions struct {
	RunLabel     string
	Storage      runoutput.StorageConfig
	TimeLimitSec int
}

// FinalizeBenchmark writes run artifacts and uploads to S3 when RunLabel is set.
// Returns the output directory path (empty when no run label).
func FinalizeBenchmark(ds *cvrp.Dataset, result BenchmarkResult, opts FinalizeOptions) (string, error) {
	if opts.RunLabel == "" {
		return "", nil
	}

	outputDir := runoutput.EnsureDir(opts.RunLabel)
	meta := map[string]interface{}{
		"problemType": "cvrp",
		"mode":        "ilp",
		"instance":    ds.Name,
		"customers":   len(ds.Customers),
		"capacity":    ds.Capacity,
		"solver":      "highs",
		"objective":   result.Objective,
		"bound":       result.LowerBound,
		"gap":         result.GapPercent,
		"status":      result.Status,
		"runtime":     result.RuntimeSeconds,
		"timeLimit":   opts.TimeLimitSec,
		"vehicles":    result.Vehicles,
		"variables":   result.Variables,
		"constraints": result.Constraints,
		"runLabel":    opts.RunLabel,
	}
	if err := runoutput.WriteILPBenchmark(outputDir, meta, result, result.SolutionJSON); err != nil {
		return outputDir, err
	}
	return outputDir, runoutput.Upload(opts.Storage, opts.RunLabel, outputDir, "ilp", result.Objective)
}
