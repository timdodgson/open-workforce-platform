package ilp

import (
	"path/filepath"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/runoutput"
)

// BenchmarkOptions extends BenchmarkConfig with run output and upload settings.
type BenchmarkOptions struct {
	Config       BenchmarkConfig
	RunLabel     string
	Storage      runoutput.StorageConfig
	TimeLimitSec int
	Parallel     bool
	SolverName   string
}

// BenchmarkOutput is the result of a finalized NRP ILP benchmark run.
type BenchmarkOutput struct {
	Result    BenchmarkResult
	OutputDir string
}

// RunAndFinalize executes the NRP ILP benchmark and writes run artifacts when RunLabel is set.
func RunAndFinalize(sc inrc2.Scenario, weekDataFiles []string, initialHist inrc2.History, opts BenchmarkOptions) (BenchmarkOutput, error) {
	cfg := opts.Config
	if opts.RunLabel != "" {
		outputDir := runoutput.EnsureDir(opts.RunLabel)
		cfg.OutputPath = filepath.Join(outputDir, "ilp-benchmark.json")
	}

	result, err := RunBenchmark(sc, weekDataFiles, initialHist, cfg)
	out := BenchmarkOutput{Result: result}
	if opts.RunLabel == "" {
		return out, err
	}

	out.OutputDir = filepath.Dir(result.SolutionPath)
	if out.OutputDir == "" {
		out.OutputDir = runoutput.EnsureDir(opts.RunLabel)
	}

	meta := map[string]interface{}{
		"problemType":    "nrp",
		"mode":           "ilp",
		"instance":       cfg.Instance,
		"weeks":          cfg.Weeks,
		"solver":         opts.SolverName,
		"timeLimit":      opts.TimeLimitSec,
		"parallel":       opts.Parallel,
		"objective":      result.Objective,
		"bound":          result.LowerBound,
		"gap":            result.GapPercent,
		"status":         result.Status,
		"runtime":        result.RuntimeSeconds,
		"hardViolations": result.HardViolations,
		"runLabel":       opts.RunLabel,
	}
	if writeErr := runoutput.WriteRunMetadata(out.OutputDir, meta); writeErr != nil && err == nil {
		err = writeErr
	}
	if uploadErr := runoutput.Upload(opts.Storage, opts.RunLabel, out.OutputDir, "ilp", result.Objective); uploadErr != nil && err == nil {
		err = uploadErr
	}
	return out, err
}
