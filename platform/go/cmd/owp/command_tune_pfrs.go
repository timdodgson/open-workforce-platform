package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/cli"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/s3upload"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

func runTunePFRS() {
	args := os.Args[2:]

	// Parse flags.
	instanceName := parseStringFlag(args, "--instance")
	if instanceName == "" {
		instanceName = "n012w8"
	}
	maxConcurrent := runtime.NumCPU()
	if v := parseIntFlag(args, "--pfrs-max-concurrent"); v > 0 {
		maxConcurrent = v
	}
	showInvalid := false
	for _, arg := range args {
		if arg == "--show-invalid" {
			showInvalid = true
		}
	}

	// Parse progress flags.
	progressEnabled := true
	if v := parseBoolFlag(args, "--progress"); v == "false" {
		progressEnabled = false
	}
	progressIntervalSec := 10
	if v := parseStringFlag(args, "--progress-interval"); v != "" {
		// Parse "10s" or just "10".
		v = strings.TrimSuffix(v, "s")
		n := 0
		for _, ch := range v {
			if ch < '0' || ch > '9' {
				fmt.Fprintf(os.Stderr, "Invalid --progress-interval: %s\n", v)
				os.Exit(1)
			}
			n = n*10 + int(ch-'0')
		}
		if n > 0 {
			progressIntervalSec = n
		}
	}

	// Parse seeds.
	seeds := []int64{42}
	if seedStr := parseStringFlag(args, "--seeds"); seedStr != "" {
		seeds = parseSeedList(seedStr)
	}

	// Parse audit CSV output path.
	auditCSVPath := parseStringFlag(args, "--audit-csv")
	if auditCSVPath == "" {
		auditCSVPath = "../web/pfrs-lab/data/results.csv"
	}

	// Parse tree CSV output path.
	treeCSVPath := parseStringFlag(args, "--tree-csv")
	if treeCSVPath == "" {
		treeCSVPath = "../web/pfrs-lab/data/tree.csv"
	}

	// Parse beam search flags.
	beamWidth := parseIntFlag(args, "--pfrs-beam-width")
	if beamWidth <= 0 {
		beamWidth = 1
	}
	var beamSeeds []int64
	if beamSeedStr := parseStringFlag(args, "--pfrs-beam-seeds"); beamSeedStr != "" {
		beamSeeds = parseSeedList(beamSeedStr)
	}

	// Parse PFRS override flags for single-config mode.
	// Support both long-form (--pfrs-iterations-per-worker) and short-form (--iterations).
	overrideIter := parseIntFlag(args, "--pfrs-iterations-per-worker")
	if overrideIter == 0 {
		overrideIter = parseIntFlag(args, "--iterations")
	}
	overrideWorkers := parseIntFlag(args, "--pfrs-max-total-workers")
	overrideTemp := parseFloatFlag(args, "--pfrs-initial-temperature")
	if overrideTemp == 0 {
		overrideTemp = parseFloatFlag(args, "--temperature")
	}
	overrideRate := parseFloatFlag(args, "--pfrs-cooling-rate")
	coolingMode := parseStringFlag(args, "--pfrs-cooling-mode")
	if coolingMode == "" {
		coolingMode = parseStringFlag(args, "--cooling")
	}
	if coolingMode == "" {
		// If user explicitly provided a cooling rate, imply fixed-rate mode.
		if overrideRate > 0 {
			coolingMode = "fixed-rate"
		} else {
			coolingMode = "adaptive"
		}
	}
	if coolingMode != "adaptive" && coolingMode != "fixed-rate" {
		fmt.Fprintf(os.Stderr, "Invalid cooling mode: %s (must be adaptive or fixed-rate)\n", coolingMode)
		os.Exit(1)
	}

	// Parse reheat flags.
	reheatThreshold := parseIntFlag(args, "--pfrs-reheat-threshold")
	reheatFactor := parseFloatFlag(args, "--pfrs-reheat-factor")
	reheatMinFraction := parseFloatFlag(args, "--pfrs-reheat-min-fraction")
	noReheat := false
	for _, arg := range args {
		if arg == "--pfrs-no-reheat" {
			noReheat = true
		}
	}

	// Parse final window flags.
	finalWindowWeeks := parseIntFlag(args, "--pfrs-final-window-weeks")
	if finalWindowWeeks <= 0 {
		finalWindowWeeks = 1 // default: no coupling
	}
	finalWindowIter := parseIntFlag(args, "--pfrs-final-window-iterations")

	// Parse look-ahead weight flag.
	lookaheadWeight := parseFloatFlag(args, "--pfrs-lookahead-weight")

	// Parse diversity slots percentage.
	diversitySlotsPct := parseIntFlag(args, "--pfrs-diversity-slots")

	// Parse beam strategy.
	beamStrategy := parseStringFlag(args, "--pfrs-beam-strategy")
	if beamStrategy == "" {
		// Auto-detect: if lookahead weight is set, default to lookahead. Otherwise none.
		if lookaheadWeight > 0 {
			beamStrategy = "lookahead"
		} else {
			beamStrategy = "none"
		}
	}
	if beamStrategy != "none" && beamStrategy != "lookahead" && beamStrategy != "budget" {
		fmt.Fprintf(os.Stderr, "Invalid --pfrs-beam-strategy: %s (must be none, lookahead, or budget)\n", beamStrategy)
		os.Exit(1)
	}

	// Parse refinement flags.
	refinementMode := parseStringFlag(args, "--pfrs-refinement")
	if refinementMode == "" {
		refinementMode = "none"
	}
	refinementIter := parseIntFlag(args, "--pfrs-refinement-iterations")
	if refinementIter <= 0 {
		refinementIter = 100000 // default 100K per week
	}
	refinementTemp := parseFloatFlag(args, "--pfrs-refinement-temperature")
	if refinementTemp <= 0 {
		refinementTemp = 10.0
	}

	// Parse worker mode flag (sa or lahc).
	workerMode := parseStringFlag(args, "--pfrs-mode")
	if workerMode == "" {
		workerMode = "sa" // default
	}
	if workerMode != "sa" && workerMode != "lahc" && workerMode != "tabu" && workerMode != "portfolio" {
		fmt.Fprintf(os.Stderr, "Invalid --pfrs-mode: %s (must be sa, lahc, tabu, or portfolio)\n", workerMode)
		os.Exit(1)
	}

	// Parse portfolio strategies.
	portfolioStrategies := parseStringFlag(args, "--pfrs-portfolio")
	var portfolio []string
	if portfolioStrategies != "" {
		portfolio = strings.Split(portfolioStrategies, ",")
		workerMode = "portfolio"
	}

	// Parse LAHC buffer length override.
	lahcBufferLength := parseIntFlag(args, "--pfrs-late-acceptance-length")

	// Parse Search Intelligence mode for PFRS worker decisions.
	// Modes: off, shadow (record only), assist, adaptive (live-updating + safety).
	workerDecisionMode := parseStringFlag(args, "--worker-decision-mode")

	// Parse run label for saving results to named directory.
	runLabel := parseRunLabelFlag(args, true)
	storage := parseStorageConfig(args, true)

	// If run label is set, redirect output to data/runs/<label>/.
	if runLabel != "" {
		labelDir := ensureRunOutputDir(runLabel)
		auditCSVPath = filepath.Join(labelDir, "results.csv")
		treeCSVPath = filepath.Join(labelDir, "tree.csv")
	}

	// Determine if running single config (any PFRS param or beam flag supplied).
	singleConfig := overrideIter > 0 || overrideWorkers > 0 || overrideTemp > 0 || overrideRate > 0 ||
		beamWidth > 1 || len(beamSeeds) > 0

	inst := loadINRC2Instance(instanceName)
	sc := inst.Scenario
	hist := inst.History
	weekFiles := inst.WeekFiles

	numWeeks := sc.NumberOfWeeks
	if numWeeks > len(weekFiles) {
		numWeeks = len(weekFiles)
	}

	// Build grid.
	var grid []inrc2.TuningGridEntry
	if singleConfig {
		// Apply defaults from DefaultPFRSConfig for any unspecified params.
		defaults := inrc2.DefaultPFRSConfig()
		iter := overrideIter
		if iter <= 0 {
			iter = defaults.IterationsPerWorker
		}
		workers := overrideWorkers
		if workers <= 0 {
			workers = defaults.MaxTotalWorkers
		}
		temp := overrideTemp
		if temp <= 0 {
			temp = defaults.InitialTemperature
		}
		rate := overrideRate
		if rate <= 0 {
			rate = defaults.CoolingRate
		}
		grid = []inrc2.TuningGridEntry{{
			IterationsPerWorker: iter,
			MaxTotalWorkers:     workers,
			InitialTemperature:  temp,
			CoolingRate:         rate,
		}}
	} else {
		iterations := []int{30000, 60000, 100000}
		workers := []int{16, 32}
		temps := []float64{1.0, 2.0, 5.0}
		rates := []float64{0.0009, 0.0005, 0.0001}
		grid = inrc2.GenerateGrid(iterations, workers, temps, rates)
	}

	// Header.
	disp := parseDisplayOptions(args)
	fmt.Println(disp.Heading(cli.EmojiConfig, "PFRS Tuning Sweep"))
	fmt.Println()
	fmt.Printf("  Instance: %s\n", disp.Bold(sc.ID))
	fmt.Printf("  Weeks:    %d\n", numWeeks)
	fmt.Printf("  Grid:     %d combinations\n", len(grid))
	fmt.Printf("  Seeds:    %d (%v)\n", len(seeds), seeds)
	fmt.Printf("  CPUs:     %d\n", maxConcurrent)
	fmt.Printf("  Cooling:  %s\n", coolingMode)
	if singleConfig && coolingMode == "adaptive" {
		// Show effective rate for the single config.
		sampleConfig := inrc2.PFRSConfig{
			InitialTemperature:  grid[0].InitialTemperature,
			MinTemperature:      0.0001,
			IterationsPerWorker: grid[0].IterationsPerWorker,
			CoolingMode:         coolingMode,
		}
		fmt.Printf("  Effective Cooling Rate: %.10f\n", sampleConfig.EffectiveCoolingRate())
	}
	fmt.Println()
	os.Stdout.Sync()

	// If beam search is active, show beam config.
	useBeamSearch := beamWidth > 1 || len(beamSeeds) > 0
	if useBeamSearch {
		fmt.Printf("  Beam Width: %d\n", beamWidth)
		if len(beamSeeds) > 0 {
			fmt.Printf("  Beam Seeds: %v\n", beamSeeds)
		}
		fmt.Println()
		os.Stdout.Sync()
	}

	algProfile, _ := optimisation.GetProfile("research")

	// PFRS worker-level Search Intelligence (--worker-decision-mode).
	workerSI := wirePFRSWorkerIntelligence(workerDecisionMode)
	decisionEngine := workerSI.Engine
	decisionRecorder := workerSI.DecisionRecorder
	assistRecorder := workerSI.AssistRecorder
	assistMode := workerSI.AssistMode

	// Audit row collection for CSV export.
	var auditRows []inrc2.WeekAuditRow

	// --- Beam Search Path ---
	if useBeamSearch && singleConfig {
		entry := grid[0]
		effectiveBeamSeeds := beamSeeds
		if len(effectiveBeamSeeds) == 0 {
			effectiveBeamSeeds = seeds // fall back to --seeds
		}

		baseConfig := inrc2.PFRSConfig{
			Mode:                       workerMode,
			Portfolio:                  portfolio,
			IterationsPerWorker:        entry.IterationsPerWorker,
			MaxConcurrentWorkers:       maxConcurrent,
			MaxTotalWorkers:            entry.MaxTotalWorkers,
			BranchOnGlobalBest:         true,
			InitialTemperature:         entry.InitialTemperature,
			CoolingRate:                entry.CoolingRate,
			CoolingMode:                coolingMode,
			MinTemperature:             0.0001,
			LateAcceptanceLength:       1000,
			TabuTenure:                 7,
			BranchCooldown:             25000,
			Deterministic:              true,
			ScoringMode:                "official-penalty",
			ReheatEnabled:              !noReheat,
			ReheatThreshold:            50000,
			ReheatFactor:               1.0,
			ReheatMinCandidateFraction: 0.20,
			DecisionEngine:             decisionEngine,
			DecisionRecorder:           decisionRecorder,
			AssistMode:                 assistMode,
			AssistRecorder:             assistRecorder,
		}

		// Auto-scale LAHC buffer: 3% of iterations (unless manually overridden).
		if workerMode == "lahc" {
			baseConfig.LateAcceptanceLength = int(float64(baseConfig.IterationsPerWorker) * 0.03)
			if baseConfig.LateAcceptanceLength < 1000 {
				baseConfig.LateAcceptanceLength = 1000
			}
		}
		if lahcBufferLength > 0 {
			baseConfig.LateAcceptanceLength = lahcBufferLength
		}
		if len(portfolio) > 0 {
			baseConfig.Portfolio = portfolio
		}

		// Apply reheat overrides from CLI flags.
		if reheatThreshold > 0 {
			baseConfig.ReheatThreshold = reheatThreshold
		}
		if reheatFactor > 0 {
			baseConfig.ReheatFactor = reheatFactor
		}
		if reheatMinFraction > 0 {
			baseConfig.ReheatMinCandidateFraction = reheatMinFraction
		}

		// Progress callback for beam runs.
		if progressEnabled {
			baseConfig.ProgressIntervalMs = int64(progressIntervalSec) * 1000
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
			BeamWidth:         beamWidth,
			Seeds:             effectiveBeamSeeds,
			FinalWindowWeeks:  finalWindowWeeks,
			FinalWindowIter:   finalWindowIter,
			LookaheadWeight:   lookaheadWeight,
			DiversitySlotsPct: diversitySlotsPct,
			BeamStrategy:      beamStrategy,
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

		// Display beam results.
		fmt.Println()
		fmt.Println(disp.Heading(cli.EmojiValid, "Beam Search Results"))
		fmt.Println()
		fmt.Printf("  Beam Width:      %d\n", beamWidth)
		fmt.Printf("  Seeds per path:  %d\n", len(effectiveBeamSeeds))
		fmt.Printf("  Total Penalty:   %s\n", disp.Green(cli.FormatInt(beamResult.TotalPenalty)))
		fmt.Printf("  All Valid:       %v\n", beamResult.AllValid)
		fmt.Println()

		fmt.Println("  Per-Week:")
		fmt.Printf("    %-5s %12s %10s %16s\n", "Week", "Candidates", "Retained", "Best Cumulative")
		for _, ws := range beamResult.WeekSummaries {
			fmt.Printf("    %-5d %12d %10d %16d\n", ws.Week, ws.Candidates, ws.Retained, ws.BestCumulative)
		}
		fmt.Println()

		fmt.Println(disp.Grey("Done."))

		// Write beam tree CSV.
		if treeCSVPath != "" {
			if err := inrc2.WriteBeamTreeCSV(treeCSVPath, beamResult); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing tree CSV: %v\n", err)
			} else {
				fmt.Fprintf(os.Stderr, "Tree CSV written: %s (%d paths)\n", treeCSVPath, len(beamResult.AllPaths))
			}
		}

		// Run context for all CSV exports.
		runCtx := inrc2.RunContext{
			RunID:       fmt.Sprintf("%s-%d", sc.ID, baseConfig.Seed),
			Instance:    sc.ID,
			Seed:        baseConfig.Seed,
			BeamWidth:   beamWidth,
			Iterations:  baseConfig.IterationsPerWorker,
			Temperature: baseConfig.InitialTemperature,
			CoolingMode: baseConfig.CoolingMode,
			Timestamp:   time.Now().Format(time.RFC3339),
		}

		// Write plateau CSV — aggregate from all winning path audits.
		var allPlateaus []inrc2.PlateauEvent
		for weekIdx, wp := range beamResult.WinningPath {
			for i := range wp.Audit.Plateaus {
				wp.Audit.Plateaus[i].Week = weekIdx + 1
			}
			allPlateaus = append(allPlateaus, wp.Audit.Plateaus...)
		}
		if len(allPlateaus) > 0 {
			plateauPath := filepath.Join(filepath.Dir(auditCSVPath), "plateaus.csv")
			if err := inrc2.WritePlateauCSV(plateauPath, runCtx, allPlateaus, baseConfig.IterationsPerWorker, beamResult.WinningPath[0].Stats.DurationMs); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing plateau CSV: %v\n", err)
			} else {
				fmt.Fprintf(os.Stderr, "Plateau CSV written: %s (%d events)\n", plateauPath, len(allPlateaus))
			}
		}

		// Write branches CSV — best-update events that triggered branches.
		var allBranchRows []inrc2.BranchRow

		// Write workers.csv — per-worker lifecycle data.
		var allWorkerRows []inrc2.WorkerLifecycleRow
		var allImprovementRows []inrc2.ImprovementRow
		var allDiscoveryRows []inrc2.DiscoveryRow
		for weekIdx, wp := range beamResult.WinningPath {
			// Build branch counts per worker from BestUpdates.
			branchCounts := make(map[int]int)
			for _, bu := range wp.Audit.BestUpdates {
				branchCounts[bu.WorkerID]++
			}
			// Build depth map from worker parent chain.
			depthMap := make(map[int]int)
			for _, w := range wp.Audit.Workers {
				depth := 0
				pid := w.ParentWorkerID
				for pid >= 0 {
					depth++
					found := false
					for _, w2 := range wp.Audit.Workers {
						if w2.WorkerID == pid {
							pid = w2.ParentWorkerID
							found = true
							break
						}
					}
					if !found {
						break
					}
				}
				depthMap[w.WorkerID] = depth
			}
			rows := inrc2.BuildWorkerLifecycleRows(runCtx, wp.Audit.Workers, weekIdx+1, wp.Seed,
				baseConfig.InitialTemperature, branchCounts, depthMap)
			allWorkerRows = append(allWorkerRows, rows...)

			// Improvements for this week.
			impRows := inrc2.BuildImprovementRows(runCtx, weekIdx+1, wp.Audit.BestUpdates, baseConfig.EffectiveCoolingRate())
			allImprovementRows = append(allImprovementRows, impRows...)

			// Branches for this week.
			parentMap := make(map[int]int)
			for _, w := range wp.Audit.Workers {
				parentMap[w.WorkerID] = w.ParentWorkerID
			}
			branchRows := inrc2.BuildBranchRows(runCtx, weekIdx+1, wp.Audit.BestUpdates,
				baseConfig.EffectiveCoolingRate(), depthMap, parentMap)
			allBranchRows = append(allBranchRows, branchRows...)

			// Discoveries for this week.
			discRows := inrc2.BuildDiscoveryRows(runCtx, weekIdx+1, wp.ID, wp.Seed,
				wp.Audit.Discoveries, depthMap, wp.Audit.WinningWorkerID)
			allDiscoveryRows = append(allDiscoveryRows, discRows...)
		}
		if len(allWorkerRows) > 0 {
			workersPath := filepath.Join(filepath.Dir(auditCSVPath), "workers.csv")
			if err := inrc2.WriteWorkerLifecycleCSV(workersPath, allWorkerRows); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing workers CSV: %v\n", err)
			} else {
				fmt.Fprintf(os.Stderr, "Workers CSV written: %s (%d workers)\n", workersPath, len(allWorkerRows))
			}
		}
		if len(allImprovementRows) > 0 {
			impPath := filepath.Join(filepath.Dir(auditCSVPath), "improvements.csv")
			if err := inrc2.WriteImprovementsCSV(impPath, allImprovementRows); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing improvements CSV: %v\n", err)
			} else {
				fmt.Fprintf(os.Stderr, "Improvements CSV written: %s (%d events)\n", impPath, len(allImprovementRows))
			}
		}
		if len(allBranchRows) > 0 {
			branchPath := filepath.Join(filepath.Dir(auditCSVPath), "branches.csv")
			if err := inrc2.WriteBranchCSV(branchPath, allBranchRows); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing branches CSV: %v\n", err)
			} else {
				fmt.Fprintf(os.Stderr, "Branches CSV written: %s (%d events)\n", branchPath, len(allBranchRows))
			}
		}

		// Write diversity CSV — beam path diversity metrics.
		diversityRows := inrc2.BuildDiversityRows(runCtx, beamResult, sc)
		if len(diversityRows) > 0 {
			diversityPath := filepath.Join(filepath.Dir(auditCSVPath), "diversity.csv")
			if err := inrc2.WriteDiversityCSV(diversityPath, diversityRows); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing diversity CSV: %v\n", err)
			} else {
				fmt.Fprintf(os.Stderr, "Diversity CSV written: %s (%d rows)\n", diversityPath, len(diversityRows))
			}
		}

		// Write discoveries CSV — every local/global best discovery event.
		if len(allDiscoveryRows) > 0 {
			discoveriesPath := filepath.Join(filepath.Dir(auditCSVPath), "discoveries.csv")
			if err := inrc2.WriteDiscoveriesCSV(discoveriesPath, allDiscoveryRows); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing discoveries CSV: %v\n", err)
			} else {
				fmt.Fprintf(os.Stderr, "Discoveries CSV written: %s (%d events)\n", discoveriesPath, len(allDiscoveryRows))
			}
		}

		// Write roster JSON — final winning schedule for dashboard visualisation.
		if len(beamResult.WinningPath) > 0 {
			// Run refinement phase if enabled.
			refinedPath := beamResult.WinningPath
			var refSummary inrc2.RefinementSummary
			if refinementMode != "none" {
				// Score before refinement (official).
				preRefPenalty, preRefViolations := officialValidate(sc, weekFiles[:numWeeks], beamResult.WinningPath, hist)
				fmt.Fprintf(os.Stderr, "\n  Before Refinement: penalty=%s violations=%d\n",
					cli.FormatInt(preRefPenalty), preRefViolations)

				fmt.Fprintf(os.Stderr, "  Refinement: %s (%d iterations/week, temp=%.1f)\n", refinementMode, refinementIter, refinementTemp)
				refinedPath, refSummary = inrc2.Refine(sc, weekFiles[:numWeeks], beamResult.WinningPath, inrc2.RefinementConfig{
					Mode:               refinementMode,
					Iterations:         refinementIter,
					Seed:               baseConfig.Seed,
					InitialTemperature: refinementTemp,
				}, hist)

				// Score after refinement (official).
				postRefPenalty, postRefViolations := officialValidate(sc, weekFiles[:numWeeks], refinedPath, hist)
				fmt.Fprintf(os.Stderr, "  After Refinement:  penalty=%s violations=%d\n",
					cli.FormatInt(postRefPenalty), postRefViolations)
				fmt.Fprintf(os.Stderr, "  Refinement result: penalty %d→%d (%+d) | violations %d→%d (%+d) | moves=%d time=%dms\n",
					preRefPenalty, postRefPenalty, postRefPenalty-preRefPenalty,
					preRefViolations, postRefViolations, postRefViolations-preRefViolations,
					refSummary.TotalMoves, refSummary.TotalDurationMs)

				// Update beam result's winning path with refined version.
				beamResult.WinningPath = refinedPath
				if len(refinedPath) > 0 {
					beamResult.TotalPenalty = refinedPath[len(refinedPath)-1].CumulativePenalty
				}
			}
			_ = refSummary

			// Final validation: re-score entire solution with official scorer and proper rolling history.
			// This is the single authoritative penalty — same as competition validator.
			{
				validationHist := hist
				var finalPenalty int
				var totalViolations int
				fmt.Fprintf(os.Stderr, "\n  Final Validation (official scorer):\n")
				for i, wp := range beamResult.WinningPath {
					weekIdx := wp.Week - 1
					if weekIdx < 0 || weekIdx >= len(weekFiles) {
						continue
					}
					wd, err := inrc2.LoadWeekData(weekFiles[weekIdx])
					if err != nil {
						continue
					}
					result := inrc2.Score(sc, wd, validationHist, wp.Solution)
					violations := len(result.SoftDetails)
					fmt.Fprintf(os.Stderr, "    Week %d: penalty=%d violations=%d hard=%d\n", wp.Week, result.SoftPenalty, violations, result.HardViolations)
					finalPenalty += result.SoftPenalty
					totalViolations += violations
					validationHist = inrc2.UpdateHistory(sc, validationHist, wp.Solution)
					// Update the path's score to match official validation.
					beamResult.WinningPath[i].WeekPenalty = result.SoftPenalty
					beamResult.WinningPath[i].ScoreResult = result
				}
				beamResult.TotalPenalty = finalPenalty
				fmt.Fprintf(os.Stderr, "  ────────────────────────────────\n")
				fmt.Fprintf(os.Stderr, "  Final Official Penalty: %s\n", disp.Green(cli.FormatInt(finalPenalty)))
				fmt.Fprintf(os.Stderr, "  Total Soft Violations:  %d\n", totalViolations)
			}

			rosterPath := filepath.Join(filepath.Dir(auditCSVPath), "roster.json")
			if err := inrc2.WriteRosterJSON(rosterPath, sc, beamResult.WinningPath); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing roster JSON: %v\n", err)
			} else {
				fmt.Fprintf(os.Stderr, "Roster JSON written: %s\n", rosterPath)
			}
		}

		// Write run.json metadata for the dashboard.
		writePFRSBeamRunJSON(filepath.Dir(auditCSVPath), pfrsBeamRunJSONParams{
			InstanceID: sc.ID, Mode: baseConfig.Mode,
			IterationsPerWorker: baseConfig.IterationsPerWorker,
			InitialTemperature:  baseConfig.InitialTemperature, CoolingMode: baseConfig.CoolingMode,
			EffectiveCoolingRate: baseConfig.EffectiveCoolingRate(),
			LateAcceptanceLength: baseConfig.LateAcceptanceLength,
			BeamWidth:            beamWidth, BeamSeeds: effectiveBeamSeeds, Seed: baseConfig.Seed,
			MaxTotalWorkers: baseConfig.MaxTotalWorkers,
			LookaheadWeight: lookaheadWeight, FinalWindowWeeks: finalWindowWeeks,
			FinalWindowIter: finalWindowIter, BeamStrategy: beamStrategy,
			DiversitySlotsPct: diversitySlotsPct, Portfolio: portfolio, RunLabel: runLabel,
		})

		// Build audit rows from the winning lineage.
		if auditCSVPath != "" && len(beamResult.WinningPath) > 0 {
			for _, wp := range beamResult.WinningPath {
				// Determine start penalty from first worker in audit.
				startPenalty := 0
				if len(wp.Audit.Workers) > 0 {
					for _, wa := range wp.Audit.Workers {
						if wa.WorkerID == 0 {
							startPenalty = wa.StartPenalty
							break
						}
					}
				}

				row := inrc2.BuildWeekAuditRow(sc.ID, baseConfig, wp.Week, startPenalty, wp.Stats, wp.ScoreResult, wp.Audit)
				row.Seed = wp.Seed
				auditRows = append(auditRows, row)
			}

			writePFRSAuditCSV(auditCSVPath, auditRows)
		}

		// --- S3 Upload ---
		s3upload.UploadRun(storage.Mode, s3upload.UploadRunConfig{
			RunLabel: runLabel, RunDir: filepath.Dir(auditCSVPath), Algorithm: baseConfig.Mode,
			Penalty: beamResult.TotalPenalty, Bucket: storage.Bucket, Region: storage.Region,
		})

		emitPFRSWorkerIntelligenceCSVs(filepath.Dir(auditCSVPath), decisionRecorder, assistRecorder)

		return
	}

	// --- Standard single-path execution ---

	// Run each grid entry with each seed.
	var multiResults []inrc2.MultiSeedResult
	var allWeekAuditBundles []inrc2.WeekAuditBundle

	for _, entry := range grid {
		var seedResults []inrc2.TuningResult

		for _, seed := range seeds {
			fmt.Fprintf(os.Stdout, "  %s%s iter=%s workers=%d temp=%.1f rate=%.4f\n",
				disp.Icon(cli.EmojiSeed),
				disp.Grey(fmt.Sprintf("[seed %d]", seed)),
				cli.FormatInt(entry.IterationsPerWorker), entry.MaxTotalWorkers,
				entry.InitialTemperature, entry.CoolingRate)
			os.Stdout.Sync()

			// Progress callback for this seed run.
			currentWeek := 0
			var progressCb inrc2.ProgressFunc
			var progressMs int64
			if progressEnabled {
				progressMs = int64(progressIntervalSec) * 1000
				progressCb = func(p inrc2.PFRSProgress) {
					fmt.Fprintf(os.Stderr, "  %s%s week %d/%d active %d queued %d total %d candidates %s best penalty %s elapsed %s\n",
						disp.Icon(cli.EmojiRunning),
						disp.Grey(fmt.Sprintf("[seed %d]", seed)),
						currentWeek+1, numWeeks, p.ActiveWorkers, p.QueueDepth, p.WorkersStarted,
						cli.FormatInt(p.CandidatesEvaluated),
						disp.Green(cli.FormatInt(p.BestPenalty)),
						cli.FormatMs(p.ElapsedMs))
					os.Stderr.Sync()
				}
			}

			// Audit callback: captures per-worker data for each week.
			var weekAudit inrc2.PFRSAudit
			var weekAuditBundles []inrc2.WeekAuditBundle
			auditCb := func(a inrc2.PFRSAudit) {
				weekAudit = a
			}

			config := inrc2.PFRSConfig{
				Mode:                 "sa",
				IterationsPerWorker:  entry.IterationsPerWorker,
				MaxConcurrentWorkers: maxConcurrent,
				MaxTotalWorkers:      entry.MaxTotalWorkers,
				BranchOnGlobalBest:   true,
				InitialTemperature:   entry.InitialTemperature,
				CoolingRate:          entry.CoolingRate,
				CoolingMode:          coolingMode,
				MinTemperature:       0.0001,
				LateAcceptanceLength: 1000,
				Seed:                 seed,
				Deterministic:        true,
				ScoringMode:          "official-penalty",
				OnProgress:           progressCb,
				ProgressIntervalMs:   progressMs,
				OnAudit:              auditCb,
				DecisionEngine:       decisionEngine,
				DecisionRecorder:     decisionRecorder,
				AssistMode:           assistMode,
				AssistRecorder:       assistRecorder,
			}

			result := inrc2.TuningResult{Entry: entry, Seed: seed}
			currentHist := hist

			for w := 0; w < numWeeks; w++ {
				currentWeek = w
				wd, err := inrc2.LoadWeekData(weekFiles[w])
				if err != nil {
					continue
				}

				weekAudit = inrc2.PFRSAudit{} // reset for this week
				sol, stats, scoreResult, err := inrc2.SolveWeekPFRS(sc, wd, currentHist, config)
				if err != nil {
					result.TotalHard += 1
					cSol, _, _ := inrc2.SolveWeek(sc, wd, currentHist, "constructive", algProfile)
					currentHist = inrc2.UpdateHistory(sc, currentHist, cSol)
					continue
				}

				result.TotalPenalty += scoreResult.SoftPenalty
				result.TotalHard += scoreResult.HardViolations
				result.TotalSoft += len(scoreResult.SoftDetails)
				result.TotalAssign += len(sol.Assignments)
				result.TotalMs += stats.DurationMs
				result.TotalCands += stats.CandidatesEvaluated

				// Concise per-week summary to stderr.
				fmt.Fprintf(os.Stderr, "    week %d: penalty=%d workers=%d branches=%d candidates=%s time=%s\n",
					w+1, scoreResult.SoftPenalty,
					stats.WorkersStarted, stats.BranchesCreated,
					cli.FormatInt(stats.CandidatesEvaluated),
					cli.FormatMs(stats.DurationMs))
				os.Stderr.Sync()

				// Build audit row for CSV.
				startPenalty := 0
				if len(weekAudit.Workers) > 0 {
					for _, wa := range weekAudit.Workers {
						if wa.WorkerID == 0 {
							startPenalty = wa.StartPenalty
							break
						}
					}
				}
				row := inrc2.BuildWeekAuditRow(sc.ID, config, w+1, startPenalty, stats, scoreResult, weekAudit)
				auditRows = append(auditRows, row)

				// Collect worker learning data for this week.
				if len(weekAudit.Workers) > 0 {
					weekAuditBundles = append(weekAuditBundles, inrc2.WeekAuditBundle{
						Week:                w + 1,
						Depth:               0,
						GlobalBestAtSpawn:   startPenalty,
						TotalWorkersStarted: stats.WorkersStarted,
						ActiveFamilies:      1,
						Workers:             weekAudit.Workers,
					})
				}

				currentHist = inrc2.UpdateHistory(sc, currentHist, sol)
			}

			result.Valid = result.TotalHard == 0
			seedResults = append(seedResults, result)

			// Collect worker learning bundles from this seed.
			allWeekAuditBundles = append(allWeekAuditBundles, weekAuditBundles...)

			// Clear progress line after seed completes.
			if progressEnabled {
				fmt.Printf("\r  %s%s penalty %s, %s                                                    \n",
					disp.Icon(cli.EmojiValid),
					disp.Grey(fmt.Sprintf("[seed %d]", seed)),
					disp.Green(cli.FormatInt(result.TotalPenalty)),
					cli.FormatMs(result.TotalMs))
			}
		}

		ms := inrc2.AggregateSeeds(entry, seedResults)
		multiResults = append(multiResults, ms)
	}
	fmt.Println()

	// Rank by average penalty.
	valid, invalid := inrc2.RankMultiSeedResults(multiResults)

	// Display results.
	fmt.Println()
	if len(valid) > 0 {
		fmt.Println(disp.Heading(cli.EmojiValid, "Valid Results (Hard = 0)"))
		fmt.Println()

		if len(seeds) > 1 {
			tbl := cli.NewTable([]cli.Column{
				{Name: "Rank", Width: 5},
				{Name: "Iterations", Width: 11, Right: true},
				{Name: "Workers", Width: 8, Right: true},
				{Name: "Temp", Width: 6, Right: true},
				{Name: "Rate", Width: 8, Right: true},
				{Name: "Avg Pen", Width: 9, Right: true},
				{Name: "Best Pen", Width: 9, Right: true},
				{Name: "Worst Pen", Width: 10, Right: true},
				{Name: "Best Seed", Width: 10, Right: true},
				{Name: "Avg Soft", Width: 9, Right: true},
				{Name: "Avg Time", Width: 10, Right: true},
				{Name: "Candidates", Width: 14, Right: true},
			}, disp)

			fmt.Println(tbl.Header())
			fmt.Println(tbl.Separator())

			for rank, r := range valid {
				row := []string{
					fmt.Sprintf("%d", rank+1),
					cli.FormatInt(r.Entry.IterationsPerWorker),
					fmt.Sprintf("%d", r.Entry.MaxTotalWorkers),
					fmt.Sprintf("%.1f", r.Entry.InitialTemperature),
					fmt.Sprintf("%.4f", r.Entry.CoolingRate),
					cli.FormatInt(r.AvgPen),
					cli.FormatInt(r.BestPen),
					cli.FormatInt(r.WorstPen),
					fmt.Sprintf("%d", r.BestSeed),
					fmt.Sprintf("%d", r.AvgSoft),
					cli.FormatMs(r.AvgMs),
					cli.FormatInt(r.TotalCands),
				}
				if rank == 0 {
					fmt.Println(tbl.HighlightRow(row))
				} else {
					fmt.Println(tbl.Row(row))
				}
			}
		} else {
			tbl := cli.NewTable([]cli.Column{
				{Name: "Rank", Width: 5},
				{Name: "Iterations", Width: 11, Right: true},
				{Name: "Workers", Width: 8, Right: true},
				{Name: "Temp", Width: 6, Right: true},
				{Name: "Rate", Width: 8, Right: true},
				{Name: "Penalty", Width: 9, Right: true},
				{Name: "Soft", Width: 6, Right: true},
				{Name: "Candidates", Width: 14, Right: true},
				{Name: "Runtime", Width: 10, Right: true},
			}, disp)

			fmt.Println(tbl.Header())
			fmt.Println(tbl.Separator())

			for rank, r := range valid {
				row := []string{
					fmt.Sprintf("%d", rank+1),
					cli.FormatInt(r.Entry.IterationsPerWorker),
					fmt.Sprintf("%d", r.Entry.MaxTotalWorkers),
					fmt.Sprintf("%.1f", r.Entry.InitialTemperature),
					fmt.Sprintf("%.4f", r.Entry.CoolingRate),
					cli.FormatInt(r.BestPen),
					fmt.Sprintf("%d", r.AvgSoft),
					cli.FormatInt(r.TotalCands),
					cli.FormatMs(r.AvgMs),
				}
				if rank == 0 {
					fmt.Println(tbl.HighlightRow(row))
				} else {
					fmt.Println(tbl.Row(row))
				}
			}
		}
	} else {
		fmt.Println(disp.Warning("No valid solutions (Hard = 0) found."))
	}

	if len(invalid) > 0 && showInvalid {
		fmt.Println()
		fmt.Println(disp.Heading(cli.EmojiInvalid, "Invalid (not all seeds Hard = 0)"))
		fmt.Println()

		tbl := cli.NewTable([]cli.Column{
			{Name: "Iterations", Width: 11, Right: true},
			{Name: "Workers", Width: 8, Right: true},
			{Name: "Temp", Width: 6, Right: true},
			{Name: "Rate", Width: 8, Right: true},
			{Name: "Avg Pen", Width: 9, Right: true},
			{Name: "Valid", Width: 7},
			{Name: "Avg Time", Width: 10, Right: true},
		}, disp)

		fmt.Println(tbl.Header())
		fmt.Println(tbl.Separator())

		for _, r := range invalid {
			row := []string{
				cli.FormatInt(r.Entry.IterationsPerWorker),
				fmt.Sprintf("%d", r.Entry.MaxTotalWorkers),
				fmt.Sprintf("%.1f", r.Entry.InitialTemperature),
				fmt.Sprintf("%.4f", r.Entry.CoolingRate),
				cli.FormatInt(r.AvgPen),
				fmt.Sprintf("%d/%d", r.ValidCount, r.Seeds),
				cli.FormatMs(r.AvgMs),
			}
			fmt.Println(tbl.ErrorRow(row))
		}
	}

	// Summary.
	fmt.Println()
	if len(valid) > 0 {
		best := valid[0]
		fmt.Println(disp.Heading(cli.EmojiBest, "Best Configuration"))
		fmt.Println()
		fmt.Printf("  %s\n", disp.Grey("--pfrs-iterations-per-worker "+cli.FormatInt(best.Entry.IterationsPerWorker)))
		fmt.Printf("  %s\n", disp.Grey("--pfrs-max-total-workers "+fmt.Sprintf("%d", best.Entry.MaxTotalWorkers)))
		fmt.Printf("  %s\n", disp.Grey("--pfrs-initial-temperature "+fmt.Sprintf("%.1f", best.Entry.InitialTemperature)))
		fmt.Printf("  %s\n", disp.Grey("--pfrs-cooling-rate "+fmt.Sprintf("%.4f", best.Entry.CoolingRate)))
		fmt.Println()
		if len(seeds) > 1 {
			fmt.Printf("  Average Penalty: %s\n", disp.Green(cli.FormatInt(best.AvgPen)))
			fmt.Printf("  Best Penalty:    %s (seed %d)\n", disp.Green(cli.FormatInt(best.BestPen)), best.BestSeed)
			fmt.Printf("  Worst Penalty:   %s\n", cli.FormatInt(best.WorstPen))
			fmt.Printf("  Average Runtime: %s\n", cli.FormatMs(best.AvgMs))
		} else {
			fmt.Printf("  Penalty: %s\n", disp.Green(cli.FormatInt(best.BestPen)))
			fmt.Printf("  Runtime: %s\n", cli.FormatMs(best.AvgMs))
		}
	}

	fmt.Println()
	fmt.Println(disp.Grey("Done."))

	// Write audit CSV if requested.
	if auditCSVPath != "" && len(auditRows) > 0 {
		// Write run.json for the standard tuning path.
		bestPenForMeta := 0
		if len(valid) > 0 {
			bestPenForMeta = valid[0].BestPen
		}
		writePFRSStandardRunJSON(filepath.Dir(auditCSVPath), pfrsStandardRunJSONParams{
			InstanceName: instanceName,
			WorkerMode:   workerMode,
			BestPenalty:  bestPenForMeta,
			RunLabel:     runLabel,
		})

		writePFRSAuditCSV(auditCSVPath, auditRows)

		// Emit worker learning telemetry.
		if len(allWeekAuditBundles) > 0 {
			learningCfg := inrc2.NRPLearningConfig{
				Instance:            instanceName,
				RunSeed:             seeds[0],
				Temperature:         overrideTemp,
				LAHCLength:          lahcBufferLength,
				IterationsPerWorker: overrideIter,
			}
			if err := inrc2.EmitNRPWorkerLearning(filepath.Dir(auditCSVPath), learningCfg, allWeekAuditBundles); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing worker_learning.csv: %v\n", err)
			} else {
				totalWorkerRecords := 0
				for _, b := range allWeekAuditBundles {
					totalWorkerRecords += len(b.Workers)
				}
				fmt.Fprintf(os.Stderr, "Worker learning CSV written: %d records\n", totalWorkerRecords)
			}
		}

		emitPFRSWorkerIntelligenceCSVs(filepath.Dir(auditCSVPath), decisionRecorder, assistRecorder)

		// Print audit summary to terminal.
		fmt.Println()
		fmt.Println(disp.Heading(cli.EmojiConfig, "Audit Summary"))
		fmt.Println()

		// Aggregate across all weeks in this run.
		totalCands := 0
		totalAccepted := 0
		totalRejected := 0
		totalRejNoop := 0
		totalRejSkill := 0
		totalRejSucc := 0
		totalRejHist := 0
		totalSABetter := 0
		totalSAWorse := 0
		totalSARejProb := 0
		totalLAHCCurrent := 0
		totalLAHCLate := 0
		totalLAHCRejLate := 0
		totalBranches := 0
		totalDropped := 0
		totalWorkers := 0
		totalImproved := 0
		totalProducedBest := 0
		maxDepth := 0
		var totalDurationMs int64

		for _, r := range auditRows {
			totalCands += r.Candidates
			totalAccepted += r.Accepted
			totalRejected += r.Rejected
			totalRejNoop += r.RejectedNoop
			totalRejSkill += r.RejectedSkill
			totalRejSucc += r.RejectedSuccession
			totalRejHist += r.RejectedHistory
			totalSABetter += r.SAAcceptedBetter
			totalSAWorse += r.SAAcceptedWorse
			totalSARejProb += r.SARejectedByProb
			totalLAHCCurrent += r.LAHCAcceptedByCurrent
			totalLAHCLate += r.LAHCAcceptedByLate
			totalLAHCRejLate += r.LAHCRejectedByLate
			totalBranches += r.BranchesCreated
			totalDropped += r.BranchesDropped
			totalWorkers += r.WorkersStarted
			totalImproved += r.WorkersImproved
			totalProducedBest += r.WorkersProducedBest
			totalDurationMs += r.DurationMs
			if r.WinningBranchDepth > maxDepth {
				maxDepth = r.WinningBranchDepth
			}
		}

		acceptRate := 0.0
		if totalCands > 0 {
			acceptRate = float64(totalAccepted) / float64(totalCands) * 100
		}

		fmt.Printf("  Candidates:  %s\n", cli.FormatInt(totalCands))
		fmt.Printf("  Accepted:    %s (%.1f%%)\n", cli.FormatInt(totalAccepted), acceptRate)
		fmt.Printf("  Rejected:    %s\n", cli.FormatInt(totalRejected))
		fmt.Println()
		fmt.Println("  Rejection Breakdown:")
		fmt.Printf("    No-op (same assignment): %s\n", cli.FormatInt(totalRejNoop))
		fmt.Printf("    Skill mismatch:          %s\n", cli.FormatInt(totalRejSkill))
		fmt.Printf("    Forbidden succession:    %s\n", cli.FormatInt(totalRejSucc))
		fmt.Printf("    History succession:       %s\n", cli.FormatInt(totalRejHist))
		fmt.Println()

		if auditRows[0].Mode == "sa" {
			fmt.Println("  SA Acceptance:")
			fmt.Printf("    Accepted (improving):    %s\n", cli.FormatInt(totalSABetter))
			fmt.Printf("    Accepted (worse, prob):  %s\n", cli.FormatInt(totalSAWorse))
			fmt.Printf("    Rejected (by prob):      %s\n", cli.FormatInt(totalSARejProb))
			fmt.Println()
		} else {
			fmt.Println("  LAHC Acceptance:")
			fmt.Printf("    Accepted (current):      %s\n", cli.FormatInt(totalLAHCCurrent))
			fmt.Printf("    Accepted (late score):   %s\n", cli.FormatInt(totalLAHCLate))
			fmt.Printf("    Rejected (late score):   %s\n", cli.FormatInt(totalLAHCRejLate))
			fmt.Println()
		}

		fmt.Println("  Branching:")
		fmt.Printf("    Total workers:           %s\n", cli.FormatInt(totalWorkers))
		fmt.Printf("    Branches created:        %s\n", cli.FormatInt(totalBranches))
		fmt.Printf("    Branches dropped:        %s\n", cli.FormatInt(totalDropped))
		fmt.Printf("    Max winning depth:       %d\n", maxDepth)
		fmt.Printf("    Workers improved parent: %s\n", cli.FormatInt(totalImproved))
		fmt.Printf("    Workers produced best:   %s\n", cli.FormatInt(totalProducedBest))
		fmt.Println()

		fmt.Println("  Per-Week:")
		fmt.Printf("    %-5s %8s %8s %8s %10s %8s %14s %10s\n",
			"Week", "Start", "Final", "Δ", "Workers", "Branches", "Candidates", "Time")
		for _, r := range auditRows {
			fmt.Printf("    %-5d %8d %8d %8d %10d %8d %14s %10s\n",
				r.Week, r.StartPenalty, r.FinalPenalty, r.Improvement,
				r.WorkersStarted, r.BranchesCreated,
				cli.FormatInt(r.Candidates), cli.FormatMs(r.DurationMs))
		}
		fmt.Println()
	}

	// --- S3 Upload (standard tuning path) ---
	bestPenaltyForUpload := 0
	if len(valid) > 0 {
		bestPenaltyForUpload = valid[0].BestPen
	}
	s3upload.UploadRun(storage.Mode, s3upload.UploadRunConfig{
		RunLabel: runLabel, RunDir: filepath.Dir(auditCSVPath), Algorithm: workerMode,
		Penalty: bestPenaltyForUpload, Bucket: storage.Bucket, Region: storage.Region,
	})
}

// officialValidate scores a complete solution path with the official scorer and rolling history.
// Returns total penalty and total soft violation count.
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
