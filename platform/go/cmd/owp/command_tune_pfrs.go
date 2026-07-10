package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/cli"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
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
	printPFRSHeader(disp, sc, opts, grid, numWeeks)

	workerSI := wirePFRSWorkerIntelligence(opts.WorkerDecisionMode, opts.PolicyMode, opts.PolicyDir)

	if opts.UseBeamSearch() && opts.SingleConfig() {
		runTunePFRSBeam(opts, disp, sc, weekFiles, numWeeks, hist, grid, workerSI)
		return
	}

	algProfile, _ := optimisation.GetProfile("research")
	progressMs := int64(0)
	if opts.ProgressEnabled {
		progressMs = int64(opts.ProgressIntervalSec) * 1000
	}

	sweep := inrc2.RunTuningSweep(inrc2.TuningSweepParams{
		Scenario:   sc,
		WeekFiles:  weekFiles,
		NumWeeks:   numWeeks,
		History:    hist,
		Grid:       grid,
		Seeds:      opts.Seeds,
		AlgProfile: algProfile,
		BuildConfig: func(entry inrc2.TuningGridEntry, seed int64, currentWeek *int) inrc2.PFRSConfig {
			var progressCb inrc2.ProgressFunc
			if opts.ProgressEnabled {
				progressCb = func(p inrc2.PFRSProgress) {
					fmt.Fprintf(os.Stderr, "  %s%s week %d/%d active %d queued %d total %d candidates %s best penalty %s elapsed %s\n",
						disp.Icon(cli.EmojiRunning),
						disp.Grey(fmt.Sprintf("[seed %d]", seed)),
						*currentWeek+1, numWeeks, p.ActiveWorkers, p.QueueDepth, p.WorkersStarted,
						cli.FormatInt(p.CandidatesEvaluated),
						disp.Green(cli.FormatInt(p.BestPenalty)),
						cli.FormatMs(p.ElapsedMs))
					os.Stderr.Sync()
				}
			}
			return inrc2.BuildTunePFRSConfig(inrc2.TuneConfigParams{
				Entry:              entry,
				Mode:               opts.WorkerMode,
				Portfolio:          opts.Portfolio,
				MaxConcurrent:      opts.MaxConcurrent,
				CoolingMode:        opts.CoolingMode,
				Seed:               seed,
				LAHCBufferLength:   opts.LAHCBufferLength,
				DecisionEngine:     workerSI.Engine,
				DecisionRecorder:   workerSI.DecisionRecorder,
				AssistMode:         workerSI.AssistMode,
				AssistRecorder:     workerSI.AssistRecorder,
				OnProgress:         progressCb,
				ProgressIntervalMs: progressMs,
			})
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

	if opts.AuditCSVPath != "" && len(sweep.AuditRows) > 0 {
		bestPenForMeta := 0
		if len(valid) > 0 {
			bestPenForMeta = valid[0].BestPen
		}
		outputDir := filepath.Dir(opts.AuditCSVPath)
		totalRecords, err := inrc2.FinalizeStandardArtifacts(inrc2.StandardArtifactsParams{
			OutputDir:    outputDir,
			AuditCSVPath: opts.AuditCSVPath,
			AuditRows:    sweep.AuditRows,
			RunJSON: inrc2.PFRSStandardRunJSONParams{
				InstanceName: opts.InstanceName,
				WorkerMode:   opts.WorkerMode,
				BestPenalty:  bestPenForMeta,
				RunLabel:     opts.RunLabel,
			},
			LearningCfg: inrc2.NRPLearningConfig{
				Instance:            opts.InstanceName,
				RunSeed:             opts.Seeds[0],
				Temperature:         opts.OverrideTemp,
				LAHCLength:          opts.LAHCBufferLength,
				IterationsPerWorker: opts.OverrideIter,
			},
			Bundles: sweep.Bundles,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error finalizing standard artifacts: %v\n", err)
		} else {
			logTelemetryFileWrite(nil, "", fmt.Sprintf("Audit CSV written: %s (%d rows)", opts.AuditCSVPath, len(sweep.AuditRows)))
			if totalRecords > 0 {
				fmt.Fprintf(os.Stderr, "Worker learning CSV written: %d records\n", totalRecords)
			}
		}

		emitPFRSTelemetry(siadapter.PFRSTelemetryInput{
			OutputDir:          outputDir,
			Instance:           opts.InstanceName,
			WorkerMode:         opts.WorkerMode,
			Portfolio:          opts.Portfolio,
			Seed:               opts.Seeds[0],
			Iterations:         opts.OverrideIter,
			BestPenalty:        bestPenForMeta,
			DecisionRecorder:   workerSI.DecisionRecorder,
			AssistRecorder:     workerSI.AssistRecorder,
			PolicyMode:         opts.PolicyMode,
			PolicyDir:          opts.PolicyDir,
			WorkerDecisionMode: opts.WorkerDecisionMode,
		})

		printPFRSAuditSummary(disp, sweep.AuditRows)
	}

	bestPenaltyForUpload := 0
	if len(valid) > 0 {
		bestPenaltyForUpload = valid[0].BestPen
	}
	uploadRunOutput(opts.Storage, opts.RunLabel, filepath.Dir(opts.AuditCSVPath), opts.WorkerMode, bestPenaltyForUpload)
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
