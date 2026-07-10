package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

func runValidateSI2() {
	args := os.Args[2:]
	if len(args) == 0 {
		printValidateSI2Usage()
		os.Exit(1)
	}

	switch args[0] {
	case "plan":
		runValidateSI2Plan(args[1:])
	case "analyze":
		runValidateSI2Analyze(args[1:])
	default:
		printValidateSI2Usage()
		os.Exit(1)
	}
}

func printValidateSI2Usage() {
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  owp validate-si2 plan [--quick]")
	fmt.Fprintln(os.Stderr, "  owp validate-si2 analyze [--runs-dir <path>] [--prefix <label-prefix>]")
}

func runValidateSI2Plan(args []string) {
	quick := false
	for _, a := range args {
		if a == "--quick" {
			quick = true
		}
	}

	cfg := optimisation.DefaultValidationConfig()
	if quick {
		cfg.Seeds = []int64{42}
		for i := range cfg.Domains {
			cfg.Domains[i].Iterations /= 10
		}
	}

	fmt.Println("SI 2.0 Validation Plan")
	fmt.Println()
	fmt.Printf("  Domains:     %d configs\n", len(cfg.Domains))
	fmt.Printf("  Policies:    %s\n", strings.Join(cfg.PolicyModes, ", "))
	fmt.Printf("  Seeds:       %d\n", len(cfg.Seeds))
	fmt.Printf("  Total runs:  %d\n", cfg.TotalExperiments())
	fmt.Println()
	fmt.Println("  Domain matrix:")
	for _, d := range cfg.Domains {
		fmt.Printf("    %-6s %-12s %-10s %d iterations\n", d.Domain, d.Instance, d.Algorithm, d.Iterations)
	}
	fmt.Println()
	fmt.Println("  Execute: .\\scripts\\validate-si2.ps1 (full) or .\\scripts\\validate-si2-quick.ps1")
	fmt.Println("  Analyze: owp validate-si2 analyze --prefix val-")
}

func runValidateSI2Analyze(args []string) {
	runsDir := "../web/pfrs-lab/data/runs"
	prefix := ""
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--runs-dir":
			if i+1 < len(args) {
				i++
				runsDir = args[i]
			}
		case "--prefix":
			if i+1 < len(args) {
				i++
				prefix = args[i]
			}
		}
	}

	results, err := optimisation.LoadExperimentResultsFromRunsDir(runsDir, prefix)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if len(results) == 0 {
		fmt.Fprintf(os.Stderr, "No runs found in %s (prefix %q)\n", runsDir, prefix)
		os.Exit(1)
	}

	fmt.Printf("SI 2.0 Analysis — %d runs from %s\n\n", len(results), runsDir)

	groups := optimisation.GroupExperimentResults(results)
	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sortStrings(keys)

	fmt.Printf("%-6s %-12s %-10s %-8s %3s %8s %8s %8s\n",
		"Domain", "Instance", "Algorithm", "Policy", "N", "Mean", "StdDev", "Min")
	fmt.Println(strings.Repeat("-", 72))
	for _, k := range keys {
		stats := optimisation.ComputeStatistics(groups[k])
		fmt.Printf("%-6s %-12s %-10s %-8s %3d %8.1f %8.1f %8d\n",
			stats.Domain, stats.Instance, stats.Algorithm, stats.PolicyMode,
			stats.N, stats.Mean, stats.StdDev, stats.Min)
	}

	comparisons := optimisation.ComparePolicyModes(results)
	if len(comparisons) > 0 {
		fmt.Println()
		fmt.Println("Policy comparisons (vs rules):")
		fmt.Printf("  %-6s %-12s %-10s %-8s → %-8s %-12s %s\n",
			"Domain", "Instance", "Algorithm", "ModeA", "ModeB", "Verdict", "MeanΔ")
		for _, c := range comparisons {
			if c.Verdict == "not_evaluated" {
				continue
			}
			fmt.Printf("  %-6s %-12s %-10s %-8s → %-8s %-12s %+.1f\n",
				c.Domain, c.Instance, c.Algorithm, c.ModeA, c.ModeB, c.Verdict, c.MeanB-c.MeanA)
		}
	}
}

func sortStrings(ss []string) {
	for i := 0; i < len(ss); i++ {
		for j := i + 1; j < len(ss); j++ {
			if ss[j] < ss[i] {
				ss[i], ss[j] = ss[j], ss[i]
			}
		}
	}
}
