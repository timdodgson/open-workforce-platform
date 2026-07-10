package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/cli"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2/siadapter"
)

func runTunePFRS() {
	args := os.Args[2:]
	opts := parseTunePFRSOptions(args)
	disp := parseDisplayOptions(args)

	inst := loadINRC2Instance(opts.InstanceName)
	sc := inst.Scenario
	hist := inst.History
	weekFiles := inst.WeekFiles

	numWeeks := sc.NumberOfWeeks
	if numWeeks > len(weekFiles) {
		numWeeks = len(weekFiles)
	}

	grid := opts.BuildGrid()
	printPFRSHeader(disp, sc, opts.TuneOptions, grid, numWeeks)

	workerSI := wirePFRSWorkerIntelligence(opts.WorkerDecisionMode, opts.PolicyMode, opts.PolicyDir)

	if opts.UseBeamSearch() && opts.SingleConfig() {
		runTunePFRSBeam(opts, disp, sc, weekFiles, numWeeks, hist, grid, workerSI)
		return
	}

	sweep := inrc2.RunTuneSweep(inrc2.TuneSweepRunParams{
		Scenario:  sc,
		WeekFiles: weekFiles,
		NumWeeks:  numWeeks,
		History:   hist,
		Options:   opts.TuneOptions,
		Grid:      grid,
		WorkerSI:  workerSI,
		OnPFRSProgress: func(p inrc2.PFRSProgress, seed int64, week, totalWeeks int) {
			fmt.Fprintf(os.Stderr, "  %s%s week %d/%d active %d queued %d total %d candidates %s best penalty %s elapsed %s\n",
				disp.Icon(cli.EmojiRunning),
				disp.Grey(fmt.Sprintf("[seed %d]", seed)),
				week, totalWeeks, p.ActiveWorkers, p.QueueDepth, p.WorkersStarted,
				cli.FormatInt(p.CandidatesEvaluated),
				disp.Green(cli.FormatInt(p.BestPenalty)),
				cli.FormatMs(p.ElapsedMs))
			os.Stderr.Sync()
		},
		Hooks: inrc2.TuningSweepHooks{
			OnSeedStart: func(entry inrc2.TuningGridEntry, seed int64) {
				fmt.Fprintf(os.Stdout, "  %s%s iter=%s workers=%d temp=%.1f rate=%.4f\n",
					disp.Icon(cli.EmojiSeed),
					disp.Grey(fmt.Sprintf("[seed %d]", seed)),
					cli.FormatInt(entry.IterationsPerWorker), entry.MaxTotalWorkers,
					entry.InitialTemperature, entry.CoolingRate)
				os.Stdout.Sync()
			},
			OnWeekLine: func(week, penalty, workers, branches, candidates int, durationMs int64) {
				fmt.Fprintf(os.Stderr, "    week %d: penalty=%d workers=%d branches=%d candidates=%s time=%s\n",
					week, penalty, workers, branches,
					cli.FormatInt(candidates), cli.FormatMs(durationMs))
				os.Stderr.Sync()
			},
			OnSeedDone: func(result inrc2.TuningResult, seed int64) {
				if !opts.ProgressEnabled {
					return
				}
				fmt.Printf("\r  %s%s penalty %s, %s                                                    \n",
					disp.Icon(cli.EmojiValid),
					disp.Grey(fmt.Sprintf("[seed %d]", seed)),
					disp.Green(cli.FormatInt(result.TotalPenalty)),
					cli.FormatMs(result.TotalMs))
			},
		},
	})
	fmt.Println()

	valid, invalid := inrc2.RankMultiSeedResults(sweep.MultiResults)
	printPFRSValidResults(disp, valid, opts.Seeds)
	if opts.ShowInvalid {
		printPFRSInvalidResults(disp, invalid)
	}
	printPFRSBestConfig(disp, valid, opts.Seeds)
	fmt.Println()
	fmt.Println(disp.Grey("Done."))

	bestPenalty := 0
	if len(valid) > 0 {
		bestPenalty = valid[0].BestPen
	}

	if fin, err := inrc2.FinalizeTuneSweep(opts.TuneOptions, sweep, bestPenalty); err != nil {
		fmt.Fprintf(os.Stderr, "Error finalizing standard artifacts: %v\n", err)
	} else if fin.AuditRowCount > 0 {
		logTelemetryFileWrite(nil, "", fmt.Sprintf("Audit CSV written: %s (%d rows)", opts.AuditCSVPath, fin.AuditRowCount))
		if fin.LearningRecordCount > 0 {
			fmt.Fprintf(os.Stderr, "Worker learning CSV written: %d records\n", fin.LearningRecordCount)
		}

		emitPFRSTelemetry(siadapter.PFRSTelemetryInput{
			OutputDir:          filepath.Dir(opts.AuditCSVPath),
			Instance:           opts.InstanceName,
			WorkerMode:         opts.WorkerMode,
			Portfolio:          opts.Portfolio,
			Seed:               opts.Seeds[0],
			Iterations:         opts.OverrideIter,
			BestPenalty:        bestPenalty,
			DecisionRecorder:   workerSI.DecisionRecorder,
			AssistRecorder:     workerSI.AssistRecorder,
			PolicyMode:         opts.PolicyMode,
			PolicyDir:          opts.PolicyDir,
			WorkerDecisionMode: opts.WorkerDecisionMode,
		})

		printPFRSAuditSummary(disp, sweep.AuditRows)
	}

	uploadRunOutput(opts.Storage, opts.RunLabel, filepath.Dir(opts.AuditCSVPath), opts.WorkerMode, bestPenalty)
}

func runTunePFRSBeam(
	opts TunePFRSOptions,
	disp cli.Options,
	sc inrc2.Scenario,
	weekFiles []string,
	numWeeks int,
	hist inrc2.History,
	grid []inrc2.TuningGridEntry,
	workerSI inrc2.WorkerIntelligenceWire,
) {
	result, err := inrc2.RunTuneBeam(inrc2.TuneBeamParams{
		Options:   opts.TuneOptions,
		Scenario:  sc,
		WeekFiles: weekFiles,
		NumWeeks:  numWeeks,
		History:   hist,
		Grid:      grid,
		WorkerSI:  workerSI,
		Hooks: inrc2.TuneBeamHooks{
			OnPFRSProgress: func(p inrc2.PFRSProgress) {
				fmt.Fprintf(os.Stderr, "  %s active %d queued %d total %d candidates %s best penalty %s elapsed %s\n",
					disp.Icon(cli.EmojiRunning),
					p.ActiveWorkers, p.QueueDepth, p.WorkersStarted,
					cli.FormatInt(p.CandidatesEvaluated),
					disp.Green(cli.FormatInt(p.BestPenalty)),
					cli.FormatMs(p.ElapsedMs))
				os.Stderr.Sync()
			},
			OnBeamWeek: func(week int, path inrc2.BeamPath) {
				fmt.Fprintf(os.Stderr, "    beam week %d: path %d (parent %d) seed %d penalty=%d cumulative=%d\n",
					week, path.ID, path.ParentID, path.Seed, path.WeekPenalty, path.CumulativePenalty)
				os.Stderr.Sync()
			},
			OnRefinementBefore: func(penalty, violations int) {
				fmt.Fprintf(os.Stderr, "\n  Before Refinement: penalty=%s violations=%d\n",
					cli.FormatInt(penalty), violations)
			},
			OnRefinementConfig: func(mode string, iter int, temp float64) {
				fmt.Fprintf(os.Stderr, "  Refinement: %s (%d iterations/week, temp=%.1f)\n", mode, iter, temp)
			},
			OnRefinementAfter: func(prePenalty, postPenalty, preViolations, postViolations int, summary inrc2.RefinementSummary) {
				fmt.Fprintf(os.Stderr, "  After Refinement:  penalty=%s violations=%d\n",
					cli.FormatInt(postPenalty), postViolations)
				fmt.Fprintf(os.Stderr, "  Refinement result: penalty %d→%d (%+d) | violations %d→%d (%+d) | moves=%d time=%dms\n",
					prePenalty, postPenalty, postPenalty-prePenalty,
					preViolations, postViolations, postViolations-preViolations,
					summary.TotalMoves, summary.TotalDurationMs)
			},
			OnFinalValidationStart: func() {
				fmt.Fprintf(os.Stderr, "\n  Final Validation (official scorer):\n")
			},
			OnFinalWeekLine: func(week, penalty, softCount, hardViolations int) {
				fmt.Fprintf(os.Stderr, "    Week %d: penalty=%d violations=%d hard=%d\n",
					week, penalty, softCount, hardViolations)
			},
			OnFinalPenalty: func(penalty, totalViolations int) {
				fmt.Fprintf(os.Stderr, "  ────────────────────────────────\n")
				fmt.Fprintf(os.Stderr, "  Final Official Penalty: %s\n", disp.Green(cli.FormatInt(penalty)))
				fmt.Fprintf(os.Stderr, "  Total Soft Violations:  %d\n", totalViolations)
			},
			OnTelemetrySummary: func(dir string, s inrc2.BeamWinningPathTelemetrySummary) {
				logBeamTelemetrySummary(dir, s)
			},
			OnArtifactMessage: func(msg string) {
				logTelemetryFileWrite(nil, "", msg)
			},
			OnError: func(msg string) {
				fmt.Fprintf(os.Stderr, "%s\n", msg)
			},
		},
	})
	if err != nil {
		os.Exit(1)
	}

	printBeamSearchResults(disp, result.BeamResult, opts.BeamWidth, result.EffectiveBeamSeeds)

	uploadRunOutput(opts.Storage, opts.RunLabel, result.OutputDir, result.BaseConfig.Mode, result.BeamResult.TotalPenalty)
	emitPFRSTelemetry(siadapter.PFRSTelemetryInput{
		Instance:           sc.ID,
		WorkerMode:         result.BaseConfig.Mode,
		Portfolio:          opts.Portfolio,
		Seed:               result.BaseConfig.Seed,
		Iterations:         result.BaseConfig.IterationsPerWorker,
		BestPenalty:        result.BeamResult.TotalPenalty,
		DecisionRecorder:   workerSI.DecisionRecorder,
		AssistRecorder:     workerSI.AssistRecorder,
		PolicyMode:         opts.PolicyMode,
		PolicyDir:          opts.PolicyDir,
		WorkerDecisionMode: opts.WorkerDecisionMode,
	})
}

func runVisualisePFRS() {
	args := os.Args[2:]

	auditCSV := parseStringFlag(args, "--audit-csv")
	if auditCSV == "" {
		fmt.Fprintln(os.Stderr, "Usage: owp visualise-pfrs --audit-csv <path> --output-dir <path>")
		os.Exit(1)
	}

	outputDir := parseStringFlag(args, "--output-dir")
	if outputDir == "" {
		outputDir = "./pfrs-report"
	}

	records, err := inrc2.ReadAuditCSV(auditCSV)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading audit CSV: %v\n", err)
		os.Exit(1)
	}

	if len(records) == 0 {
		fmt.Fprintln(os.Stderr, "No records found in audit CSV")
		os.Exit(1)
	}

	if err := inrc2.GenerateReport(records, outputDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating report: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("PFRS dashboard generated: %s/summary.html\n", outputDir)
}
