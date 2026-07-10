package main

import (
	"fmt"
	"os"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/cli"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2/legacysearch"
)

func parseDisplayOptions(args []string) cli.Options {
	opts := cli.DefaultOptions()
	for _, arg := range args {
		switch arg {
		case "--plain":
			opts.Colour = false
			opts.Emoji = false
		case "--no-colour", "--no-color":
			opts.Colour = false
		case "--no-emoji":
			opts.Emoji = false
		}
	}
	return opts
}

func displayEffectiveConfig(algorithm string, p legacysearch.AlgorithmProfile) {
	fmt.Println("Effective Configuration:")
	switch algorithm {
	case "constructive":
		fmt.Println("  (no tunables)")
	case "hill-climbing":
		fmt.Printf("  HCMaxIterations: %d\n", p.HCMaxIterations)
	case "simulated-annealing":
		fmt.Printf("  SAMaxIterations: %d\n", p.SAMaxIterations)
		fmt.Printf("  SAInitialTemperature: %.1f\n", p.SAInitialTemperature)
		fmt.Printf("  SACoolingRate: %.4f\n", p.SACoolingRate)
		fmt.Printf("  SAMinTemperature: %.2f\n", p.SAMinTemperature)
	case "tabu-search":
		fmt.Printf("  TabuMaxIterations: %d\n", p.TabuMaxIterations)
		fmt.Printf("  TabuListSize: %d\n", p.TabuListSize)
		fmt.Printf("  TabuAspirationEnabled: %v\n", p.TabuAspirationEnabled)
	case "large-neighbourhood-search":
		fmt.Printf("  LNSIterations: %d\n", p.LNSIterations)
		fmt.Printf("  LNSDestroySize: %d\n", p.LNSDestroySize)
		fmt.Printf("  LNSRepairStrategy: %s\n", p.LNSRepairStrategy)
	case "parallel-feasible-roster-search":
		pfrsConfig := parsePFRSConfig(os.Args[1:])
		fmt.Printf("  PFRSMode: %s\n", pfrsConfig.Mode)
		fmt.Printf("  PFRSIterationsPerWorker: %d\n", pfrsConfig.IterationsPerWorker)
		fmt.Printf("  PFRSMaxConcurrentWorkers: %d\n", pfrsConfig.MaxConcurrentWorkers)
		fmt.Printf("  PFRSMaxTotalWorkers: %d\n", pfrsConfig.MaxTotalWorkers)
		fmt.Printf("  PFRSInitialTemperature: %.4f\n", pfrsConfig.InitialTemperature)
		fmt.Printf("  PFRSCoolingRate: %.4f\n", pfrsConfig.CoolingRate)
		fmt.Printf("  PFRSMinTemperature: %.4f\n", pfrsConfig.MinTemperature)
		fmt.Printf("  PFRSLateAcceptanceLength: %d\n", pfrsConfig.LateAcceptanceLength)
		fmt.Printf("  PFRSSeed: %d\n", pfrsConfig.Seed)
		fmt.Printf("  PFRSDeterministic: %v\n", pfrsConfig.Deterministic)
	default:
		fmt.Printf("  HCMaxIterations: %d\n", p.HCMaxIterations)
		fmt.Printf("  SAMaxIterations: %d\n", p.SAMaxIterations)
		fmt.Printf("  SAInitialTemperature: %.1f\n", p.SAInitialTemperature)
		fmt.Printf("  SACoolingRate: %.4f\n", p.SACoolingRate)
		fmt.Printf("  SAMinTemperature: %.2f\n", p.SAMinTemperature)
		fmt.Printf("  TabuMaxIterations: %d\n", p.TabuMaxIterations)
		fmt.Printf("  TabuListSize: %d\n", p.TabuListSize)
		fmt.Printf("  TabuAspirationEnabled: %v\n", p.TabuAspirationEnabled)
		fmt.Printf("  LNSIterations: %d\n", p.LNSIterations)
		fmt.Printf("  LNSDestroySize: %d\n", p.LNSDestroySize)
		fmt.Printf("  LNSRepairStrategy: %s\n", p.LNSRepairStrategy)
	}
	fmt.Println()
}
