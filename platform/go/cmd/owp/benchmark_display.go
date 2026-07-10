package main

import (
	"fmt"
	"strings"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/cli"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2"
)

func printINRC2BenchmarkResults(valid, invalid []*inrc2.AlgorithmBenchmarkResult, showInvalid bool) {
	if len(valid) > 0 {
		fmt.Printf("%-6s %-28s %10s %8s %12s %10s %12s\n",
			"Rank", "Algorithm", "Penalty", "Soft", "Assignments", "Runtime", "Candidates")
		fmt.Println(strings.Repeat("-", 92))

		for rank, r := range valid {
			fmt.Printf("%-6d %-28s %10d %8d %12d %8dms %12d\n",
				rank+1, r.Algorithm, r.TotalPenalty, r.TotalSoft,
				r.TotalAssign, r.TotalMs, r.TotalCands)
		}
	} else {
		fmt.Println("No valid solutions (Hard = 0) found.")
	}

	if len(invalid) > 0 && showInvalid {
		fmt.Println()
		fmt.Println("Rejected (Invalid Solutions):")
		fmt.Printf("       %-28s %10s %8s %8s %12s %10s %12s\n",
			"Algorithm", "Penalty", "Hard", "Soft", "Assignments", "Runtime", "Candidates")
		fmt.Println("       " + strings.Repeat("-", 92))

		for _, r := range invalid {
			fmt.Printf("       %-28s %10d %8d %8d %12d %8dms %12d\n",
				r.Algorithm, r.TotalPenalty, r.TotalHard, r.TotalSoft,
				r.TotalAssign, r.TotalMs, r.TotalCands)
		}
	}
}

func printINRC2BenchmarkSummary(valid, invalid []*inrc2.AlgorithmBenchmarkResult, showInvalid bool) {
	fmt.Println()
	fmt.Println("Summary:")

	if len(valid) == 0 {
		fmt.Println("  No valid solution found.")
		if showInvalid && len(invalid) > 0 {
			fmt.Printf("  Least invalid:     %s (hard: %d, penalty: %d)\n",
				invalid[0].Algorithm, invalid[0].TotalHard, invalid[0].TotalPenalty)
		}
		fmt.Println()
		fmt.Println("Done.")
		return
	}

	fmt.Printf("  Best algorithm:    %s (penalty: %d)\n", valid[0].Algorithm, valid[0].TotalPenalty)

	fastest := valid[0]
	for _, r := range valid {
		if r.TotalMs < fastest.TotalMs {
			fastest = r
		}
	}
	fmt.Printf("  Fastest valid:     %s (%dms)\n", fastest.Algorithm, fastest.TotalMs)

	totalPenalty, totalMs, totalSoft := 0, int64(0), 0
	for _, r := range valid {
		totalPenalty += r.TotalPenalty
		totalMs += r.TotalMs
		totalSoft += r.TotalSoft
	}
	n := len(valid)
	fmt.Printf("  Average penalty:   %d\n", totalPenalty/n)
	fmt.Printf("  Average runtime:   %dms\n", totalMs/int64(n))
	fmt.Printf("  Average soft:      %d\n", totalSoft/n)
	fmt.Println()
	fmt.Println("Done.")
}

func printRoutingILPResult(disp cli.Options, status string, objective, lowerBound int, gap, runtime float64, variables, constraints, vehicles int, notes string) {
	fmt.Println(disp.Heading(cli.EmojiValid, "Result"))
	fmt.Println()
	fmt.Printf("  Status:      %s\n", status)
	fmt.Printf("  Objective:   %d\n", objective)
	if lowerBound > 0 {
		fmt.Printf("  Lower Bound: %d\n", lowerBound)
		fmt.Printf("  Gap:         %.2f%%\n", gap)
	}
	fmt.Printf("  Runtime:     %.1fs\n", runtime)
	fmt.Printf("  Variables:   %d\n", variables)
	fmt.Printf("  Constraints: %d\n", constraints)
	fmt.Printf("  Vehicles:    %d\n", vehicles)
	if notes != "" {
		fmt.Printf("  Notes:       %s\n", notes)
	}
	fmt.Println()
}

func printJSSILPResult(disp cli.Options, status string, makespan, lowerBound int, gap, runtime float64, variables, constraints, operations int, notes string) {
	fmt.Println(disp.Heading(cli.EmojiValid, "Result"))
	fmt.Println()
	fmt.Printf("  Status:      %s\n", status)
	fmt.Printf("  Makespan:    %d\n", makespan)
	if lowerBound > 0 {
		fmt.Printf("  Lower Bound: %d\n", lowerBound)
		fmt.Printf("  Gap:         %.2f%%\n", gap)
	}
	fmt.Printf("  Runtime:     %.1fs\n", runtime)
	fmt.Printf("  Variables:   %d\n", variables)
	fmt.Printf("  Constraints: %d\n", constraints)
	fmt.Printf("  Operations:  %d\n", operations)
	if notes != "" {
		fmt.Printf("  Notes:       %s\n", notes)
	}
	fmt.Println()
}
