package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/cli"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2/siadapter"
)

func runTunePFRSBeam(
	opts TunePFRSOptions,
	disp cli.Options,
	sc inrc2.Scenario,
	weekFiles []string,
	numWeeks int,
	hist inrc2.History,
	grid []inrc2.TuningGridEntry,
	workerSI pfrsWorkerIntelligence,
) {
	entry := grid[0]
	effectiveBeamSeeds := opts.BeamSeeds
	if len(effectiveBeamSeeds) == 0 {
		effectiveBeamSeeds = opts.Seeds
	}

	baseConfig := inrc2.BuildTunePFRSConfig(inrc2.TuneConfigParams{
		Entry:             entry,
		Mode:              opts.WorkerMode,
		Portfolio:         opts.Portfolio,
		MaxConcurrent:     opts.MaxConcurrent,
		CoolingMode:       opts.CoolingMode,
		LAHCBufferLength:  opts.LAHCBufferLength,
		ReheatThreshold:   opts.ReheatThreshold,
		ReheatFactor:      opts.ReheatFactor,
		ReheatMinFraction: opts.ReheatMinFraction,
		NoReheat:          opts.NoReheat,
		Beam:              true,
		DecisionEngine:    workerSI.Engine,
		DecisionRecorder:  workerSI.DecisionRecorder,
		AssistMode:        workerSI.AssistMode,
		AssistRecorder:    workerSI.AssistRecorder,
	})

	if opts.ProgressEnabled {
		baseConfig.ProgressIntervalMs = int64(opts.ProgressIntervalSec) * 1000
		baseConfig.OnProgress = func(p inrc2.PFRSProgress) {
			fmt.Fprintf(os.Stderr, "  %s active %d queued %d total %d candidates %s best penalty %s elapsed %s\n",
				disp.Icon(cli.EmojiRunning),
				p.ActiveWorkers, p.QueueDepth, p.WorkersStarted,
				cli.FormatInt(p.CandidatesEvaluated),
				disp.Green(cli.FormatInt(p.BestPenalty)),
				cli.FormatMs(p.ElapsedMs))
			os.Stderr.Sync()
		}
	}

	beam := inrc2.BeamConfig{
		BeamWidth:         opts.BeamWidth,
		Seeds:             effectiveBeamSeeds,
		FinalWindowWeeks:  opts.FinalWindowWeeks,
		FinalWindowIter:   opts.FinalWindowIter,
		LookaheadWeight:   opts.LookaheadWeight,
		DiversitySlotsPct: opts.DiversitySlotsPct,
		BeamStrategy:      opts.BeamStrategy,
	}

	onProgress := func(week int, path inrc2.BeamPath) {
		fmt.Fprintf(os.Stderr, "    beam week %d: path %d (parent %d) seed %d penalty=%d cumulative=%d\n",
			week, path.ID, path.ParentID, path.Seed, path.WeekPenalty, path.CumulativePenalty)
		os.Stderr.Sync()
	}

	beamResult, err := inrc2.RunBeamSearch(sc, weekFiles[:numWeeks], hist, baseConfig, beam, onProgress)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Beam search failed: %v\n", err)
		os.Exit(1)
	}

	printBeamSearchResults(disp, beamResult, opts.BeamWidth, effectiveBeamSeeds)

	if opts.TreeCSVPath != "" {
		logTelemetryFileWrite(
			inrc2.WriteBeamTreeCSV(opts.TreeCSVPath, beamResult),
			"tree CSV",
			fmt.Sprintf("Tree CSV written: %s (%d paths)", opts.TreeCSVPath, len(beamResult.AllPaths)),
		)
	}

	runCtx := inrc2.RunContext{
		RunID:       fmt.Sprintf("%s-%d", sc.ID, baseConfig.Seed),
		Instance:    sc.ID,
		Seed:        baseConfig.Seed,
		BeamWidth:   opts.BeamWidth,
		Iterations:  baseConfig.IterationsPerWorker,
		Temperature: baseConfig.InitialTemperature,
		CoolingMode: baseConfig.CoolingMode,
		Timestamp:   time.Now().Format(time.RFC3339),
	}

	telemetryDir := filepath.Dir(opts.AuditCSVPath)
	telSummary, err := inrc2.WriteBeamWinningPathTelemetry(inrc2.BeamWinningPathTelemetryParams{
		TelemetryDir: telemetryDir,
		RunCtx:       runCtx,
		Config:       baseConfig,
		WinningPath:  beamResult.WinningPath,
		Scenario:     sc,
		BeamResult:   beamResult,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing beam telemetry CSVs: %v\n", err)
	} else {
		logBeamTelemetrySummary(telemetryDir, telSummary)
	}

	if len(beamResult.WinningPath) > 0 {
		refinedPath := beamResult.WinningPath
		var refSummary inrc2.RefinementSummary
		if opts.RefinementMode != "none" {
			preRefPenalty, preRefViolations := officialValidate(sc, weekFiles[:numWeeks], beamResult.WinningPath, hist)
			fmt.Fprintf(os.Stderr, "\n  Before Refinement: penalty=%s violations=%d\n",
				cli.FormatInt(preRefPenalty), preRefViolations)
			fmt.Fprintf(os.Stderr, "  Refinement: %s (%d iterations/week, temp=%.1f)\n",
				opts.RefinementMode, opts.RefinementIter, opts.RefinementTemp)

			refinedPath, refSummary = inrc2.Refine(sc, weekFiles[:numWeeks], beamResult.WinningPath, inrc2.RefinementConfig{
				Mode:               opts.RefinementMode,
				Iterations:         opts.RefinementIter,
				Seed:               baseConfig.Seed,
				InitialTemperature: opts.RefinementTemp,
			}, hist)

			postRefPenalty, postRefViolations := officialValidate(sc, weekFiles[:numWeeks], refinedPath, hist)
			fmt.Fprintf(os.Stderr, "  After Refinement:  penalty=%s violations=%d\n",
				cli.FormatInt(postRefPenalty), postRefViolations)
			fmt.Fprintf(os.Stderr, "  Refinement result: penalty %d→%d (%+d) | violations %d→%d (%+d) | moves=%d time=%dms\n",
				preRefPenalty, postRefPenalty, postRefPenalty-preRefPenalty,
				preRefViolations, postRefViolations, postRefViolations-preRefViolations,
				refSummary.TotalMoves, refSummary.TotalDurationMs)

			beamResult.WinningPath = refinedPath
			if len(refinedPath) > 0 {
				beamResult.TotalPenalty = refinedPath[len(refinedPath)-1].CumulativePenalty
			}
		}
		_ = refSummary

		updatedPath, finalPenalty, totalViolations := inrc2.OfficialRevalidateBeamPath(sc, weekFiles[:numWeeks], beamResult.WinningPath, hist)
		beamResult.WinningPath = updatedPath
		beamResult.TotalPenalty = finalPenalty
		fmt.Fprintf(os.Stderr, "\n  Final Validation (official scorer):\n")
		for _, wp := range beamResult.WinningPath {
			fmt.Fprintf(os.Stderr, "    Week %d: penalty=%d violations=%d hard=%d\n",
				wp.Week, wp.WeekPenalty, len(wp.ScoreResult.SoftDetails), wp.ScoreResult.HardViolations)
		}
		fmt.Fprintf(os.Stderr, "  ────────────────────────────────\n")
		fmt.Fprintf(os.Stderr, "  Final Official Penalty: %s\n", disp.Green(cli.FormatInt(finalPenalty)))
		fmt.Fprintf(os.Stderr, "  Total Soft Violations:  %d\n", totalViolations)

		rosterPath := filepath.Join(telemetryDir, "roster.json")
		logTelemetryFileWrite(
			inrc2.WriteRosterJSON(rosterPath, sc, beamResult.WinningPath),
			"roster JSON",
			fmt.Sprintf("Roster JSON written: %s", rosterPath),
		)
	}

	outputDir := filepath.Dir(opts.AuditCSVPath)
	if err := inrc2.FinalizeBeamArtifacts(inrc2.BeamArtifactsParams{
		OutputDir:    outputDir,
		AuditCSVPath: opts.AuditCSVPath,
		ScenarioID:   sc.ID,
		Config:       baseConfig,
		WinningPath:  beamResult.WinningPath,
		RunJSON: inrc2.PFRSBeamRunJSONParams{
			InstanceID: sc.ID, Mode: baseConfig.Mode,
			IterationsPerWorker: baseConfig.IterationsPerWorker,
			InitialTemperature:  baseConfig.InitialTemperature, CoolingMode: baseConfig.CoolingMode,
			EffectiveCoolingRate: baseConfig.EffectiveCoolingRate(),
			LateAcceptanceLength: baseConfig.LateAcceptanceLength,
			BeamWidth: opts.BeamWidth, BeamSeeds: effectiveBeamSeeds, Seed: baseConfig.Seed,
			MaxTotalWorkers: baseConfig.MaxTotalWorkers,
			LookaheadWeight: opts.LookaheadWeight, FinalWindowWeeks: opts.FinalWindowWeeks,
			FinalWindowIter: opts.FinalWindowIter, BeamStrategy: opts.BeamStrategy,
			DiversitySlotsPct: opts.DiversitySlotsPct, Portfolio: opts.Portfolio, RunLabel: opts.RunLabel,
		},
		LearningCfg: inrc2.NRPLearningConfig{
			Instance:            sc.ID,
			RunSeed:             baseConfig.Seed,
			Temperature:         baseConfig.InitialTemperature,
			LAHCLength:          baseConfig.LateAcceptanceLength,
			TabuTenure:          0,
			IterationsPerWorker: baseConfig.IterationsPerWorker,
		},
	}); err != nil {
		fmt.Fprintf(os.Stderr, "Error finalizing beam artifacts: %v\n", err)
	}

	uploadRunOutput(opts.Storage, opts.RunLabel, outputDir, baseConfig.Mode, beamResult.TotalPenalty)
	emitPFRSTelemetry(siadapter.PFRSTelemetryInput{
		Instance:           sc.ID,
		WorkerMode:         baseConfig.Mode,
		Portfolio:          opts.Portfolio,
		Seed:               baseConfig.Seed,
		Iterations:         baseConfig.IterationsPerWorker,
		BestPenalty:        beamResult.TotalPenalty,
		DecisionRecorder:   workerSI.DecisionRecorder,
		AssistRecorder:     workerSI.AssistRecorder,
		PolicyMode:         opts.PolicyMode,
		PolicyDir:          opts.PolicyDir,
		WorkerDecisionMode: opts.WorkerDecisionMode,
	})
}

func logBeamTelemetrySummary(dir string, s inrc2.BeamWinningPathTelemetrySummary) {
	if s.PlateauEvents > 0 {
		logTelemetryFileWrite(nil, "", fmt.Sprintf("Plateau CSV written: %s (%d events)", filepath.Join(dir, "plateaus.csv"), s.PlateauEvents))
	}
	if s.WorkerRows > 0 {
		logTelemetryFileWrite(nil, "", fmt.Sprintf("Workers CSV written: %s (%d workers)", filepath.Join(dir, "workers.csv"), s.WorkerRows))
	}
	if s.ImprovementRows > 0 {
		logTelemetryFileWrite(nil, "", fmt.Sprintf("Improvements CSV written: %s (%d events)", filepath.Join(dir, "improvements.csv"), s.ImprovementRows))
	}
	if s.BranchRows > 0 {
		logTelemetryFileWrite(nil, "", fmt.Sprintf("Branches CSV written: %s (%d events)", filepath.Join(dir, "branches.csv"), s.BranchRows))
	}
	if s.DiversityRows > 0 {
		logTelemetryFileWrite(nil, "", fmt.Sprintf("Diversity CSV written: %s (%d rows)", filepath.Join(dir, "diversity.csv"), s.DiversityRows))
	}
	if s.DiscoveryRows > 0 {
		logTelemetryFileWrite(nil, "", fmt.Sprintf("Discoveries CSV written: %s (%d events)", filepath.Join(dir, "discoveries.csv"), s.DiscoveryRows))
	}
}
