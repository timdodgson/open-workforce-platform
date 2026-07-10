package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/vrptw"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

func runSolveVRPTW() {
	warnDeprecatedSolveAlias("solve-vrptw", "vrptw")
	runSolveDomain("vrptw", os.Args[2:])
}

func finalizeVRPTWRun(opts SearchSolveOptions, config optimisation.SearchConfig, workerDecisionMode string, instancePath string, ds *vrptw.Dataset, problem *vrptw.VRPTWProblem, outcome searchRunOutcome, bestDistance, baselineDistance, bestVehicles int, bestFeasible bool) {
	outputDir := ensureRunOutputDir(opts.RunLabel)
	instanceName := strings.TrimSuffix(filepath.Base(instancePath), filepath.Ext(instancePath))
	solJSON, _ := problem.SerializeSolution(outcome.Result.BestSolution)

	extra := map[string][]byte{}
	if disc := vrptw.BuildDiscoveriesCSV(outcome.Result.Discoveries); disc != nil {
		extra["discoveries.csv"] = disc
	}

	finalizeGenericSolverRun(genericSolverRunOutput{
		OutputDir: outputDir,
		RunMeta: map[string]interface{}{
			"problemType":     "vrptw",
			"mode":            opts.Mode,
			"instance":        instanceName,
			"customers":       len(ds.Customers),
			"capacity":        ds.Capacity,
			"vehicles":        ds.Vehicles,
			"iterations":      opts.Iterations,
			"seed":            opts.Seed,
			"runLabel":        opts.RunLabel,
			"bestObjective":   bestDistance,
			"bestDistance":    bestDistance,
			"initialDistance": baselineDistance,
			"bestVehicles":    bestVehicles,
			"feasible":        bestFeasible,
			"runtimeMs":       outcome.Result.DurationMs,
			"policyMode":      config.PolicyMode,
			"policyDir":       config.PolicyDir,
		},
		SolutionJSON: solJSON,
		ExtraFiles:   extra,
		Telemetry: solverTelemetryInput{
			OutputDir: outputDir, ProblemType: "vrptw", Instance: instanceName, Algorithm: opts.Mode,
			Seed: opts.Seed, Temperature: opts.Temperature, Iterations: opts.Iterations,
			AssistMode: workerDecisionMode,
			Result: outcome.Result, PortfolioRecorder: outcome.Recorder,
			PolicyMode: config.PolicyMode, PolicyDir: config.PolicyDir,
		},
		Storage: opts.Storage, RunLabel: opts.RunLabel, Algorithm: opts.Mode, Penalty: bestDistance,
	})
	fmt.Printf("  Output: %s/\n", outputDir)
}
