package ilp

import (
	"path/filepath"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/runoutput"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/vrptw"
)

// FinalizeOptions configures run artifact output for a VRPTW ILP benchmark.
type FinalizeOptions struct {
	RunLabel     string
	Storage      runoutput.StorageConfig
	TimeLimitSec int
}

// FinalizeBenchmark writes run artifacts and uploads to S3 when RunLabel is set.
func FinalizeBenchmark(ds *vrptw.Dataset, result BenchmarkResult, opts FinalizeOptions) (string, error) {
	if opts.RunLabel == "" {
		return "", nil
	}

	outputDir := runoutput.EnsureDir(opts.RunLabel)
	meta := map[string]interface{}{
		"problemType": "vrptw",
		"mode":        "ilp",
		"instance":    filepath.Base(ds.Name),
		"customers":   len(ds.Customers),
		"capacity":    ds.Capacity,
		"vehicles":    ds.Vehicles,
		"solver":      "highs",
		"objective":   result.Objective,
		"bound":       result.LowerBound,
		"gap":         result.GapPercent,
		"status":      result.Status,
		"runtime":     result.RuntimeSeconds,
		"timeLimit":   opts.TimeLimitSec,
		"variables":   result.Variables,
		"constraints": result.Constraints,
		"runLabel":    opts.RunLabel,
	}
	if err := runoutput.WriteILPBenchmark(outputDir, meta, result, result.SolutionJSON); err != nil {
		return outputDir, err
	}
	return outputDir, runoutput.Upload(opts.Storage, opts.RunLabel, outputDir, "ilp", result.Objective)
}
