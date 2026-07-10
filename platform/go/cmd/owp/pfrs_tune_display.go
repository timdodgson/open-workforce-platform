package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/cli"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2"
)

func printPFRSHeader(disp cli.Options, sc inrc2.Scenario, opts inrc2.TuneOptions, grid []inrc2.TuningGridEntry, numWeeks int) {
	fmt.Println(disp.Heading(cli.EmojiConfig, "PFRS Tuning Sweep"))
	fmt.Println()
	fmt.Printf("  Instance: %s\n", disp.Bold(sc.ID))
	fmt.Printf("  Weeks:    %d\n", numWeeks)
	fmt.Printf("  Grid:     %d combinations\n", len(grid))
	fmt.Printf("  Seeds:    %d (%v)\n", len(opts.Seeds), opts.Seeds)
	fmt.Printf("  CPUs:     %d\n", opts.MaxConcurrent)
	fmt.Printf("  Cooling:  %s\n", opts.CoolingMode)
	if opts.SingleConfig() && opts.CoolingMode == "adaptive" {
		sampleConfig := inrc2.PFRSConfig{
			InitialTemperature:  grid[0].InitialTemperature,
			MinTemperature:      0.0001,
			IterationsPerWorker: grid[0].IterationsPerWorker,
			CoolingMode:         opts.CoolingMode,
		}
		fmt.Printf("  Effective Cooling Rate: %.10f\n", sampleConfig.EffectiveCoolingRate())
	}
	fmt.Println()
	os.Stdout.Sync()

	if opts.UseBeamSearch() {
		fmt.Printf("  Beam Width: %d\n", opts.BeamWidth)
		if len(opts.BeamSeeds) > 0 {
			fmt.Printf("  Beam Seeds: %v\n", opts.BeamSeeds)
		}
		fmt.Println()
		os.Stdout.Sync()
	}
}

func printBeamSearchResults(disp cli.Options, beamResult inrc2.BeamResult, beamWidth int, beamSeeds []int64) {
	fmt.Println()
	fmt.Println(disp.Heading(cli.EmojiValid, "Beam Search Results"))
	fmt.Println()
	fmt.Printf("  Beam Width:      %d\n", beamWidth)
	fmt.Printf("  Seeds per path:  %d\n", len(beamSeeds))
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
}

func printPFRSValidResults(disp cli.Options, valid []inrc2.MultiSeedResult, seeds []int64) {
	if len(valid) == 0 {
		fmt.Println(disp.Warning("No valid solutions (Hard = 0) found."))
		return
	}

	fmt.Println()
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
		return
	}

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

func printPFRSInvalidResults(disp cli.Options, invalid []inrc2.MultiSeedResult) {
	if len(invalid) == 0 {
		return
	}

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

func printPFRSBestConfig(disp cli.Options, valid []inrc2.MultiSeedResult, seeds []int64) {
	if len(valid) == 0 {
		return
	}

	best := valid[0]
	fmt.Println()
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
