package ilp

import (
	"path/filepath"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/jobshop"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/runoutput"
)

// FinalizeOptions configures run artifact output for a JSS ILP benchmark.
type FinalizeOptions struct {
	RunLabel     string
	InstancePath string
	Storage      runoutput.StorageConfig
	TimeLimitSec int
}

// FinalizeBenchmark writes run artifacts and uploads to S3 when RunLabel is set.
func FinalizeBenchmark(ds *jobshop.Dataset, result BenchmarkResult, opts FinalizeOptions) (string, error) {
	if opts.RunLabel == "" {
		return "", nil
	}

	instanceName := opts.InstancePath
	if instanceName == "" {
		instanceName = ds.Name
	}
	instanceName = filepath.Base(instanceName)

	outputDir := runoutput.EnsureDir(opts.RunLabel)
	meta := map[string]interface{}{
		"problemType":  "jss",
		"mode":         "ilp",
		"instance":     instanceName,
		"jobs":         ds.Jobs,
		"machines":     ds.Machines,
		"solver":       "highs",
		"objective":    result.Objective,
		"bestMakespan": result.Objective,
		"bound":        result.LowerBound,
		"gap":          result.GapPercent,
		"status":       result.Status,
		"runtime":      result.RuntimeSeconds,
		"timeLimit":    opts.TimeLimitSec,
		"variables":    result.Variables,
		"constraints":  result.Constraints,
		"runLabel":     opts.RunLabel,
	}
	if err := runoutput.WriteILPBenchmark(outputDir, meta, result, result.SolutionJSON); err != nil {
		return outputDir, err
	}
	return outputDir, runoutput.Upload(opts.Storage, opts.RunLabel, outputDir, "ilp", result.Objective)
}
