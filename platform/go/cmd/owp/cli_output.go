package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/cli"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/event"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/resource"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/loader"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
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

func displayEffectiveConfig(algorithm string, p optimisation.AlgorithmProfile) {
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
		// Show all for benchmark/unknown.
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

func buildCapacityLookup(resources []resource.Resource) map[string]int {
	lookup := make(map[string]int, len(resources))
	for _, res := range resources {
		var details struct {
			Capacity int `json:"capacity"`
		}
		if err := json.Unmarshal(res.Details(), &details); err == nil {
			lookup[res.ID()] = details.Capacity
		}
	}
	return lookup
}

// buildDurationLookup reads duration from each event's details for display.
// Work item IDs are "WI-" + event ID.
func buildDurationLookup(events []event.BusinessEvent) map[string]int {
	lookup := make(map[string]int, len(events))
	for _, evt := range events {
		var details struct {
			Duration int `json:"duration"`
		}
		if err := json.Unmarshal(evt.Details(), &details); err == nil {
			dur := details.Duration
			if dur <= 0 {
				dur = 1
			}
			lookup["WI-"+evt.ID()] = dur
		}
	}
	return lookup
}

// convertTravel converts loader travel entries to optimisation travel entries.
func convertTravel(entries []loader.TravelEntry) []optimisation.TravelEntry {
	result := make([]optimisation.TravelEntry, len(entries))
	for i, e := range entries {
		result[i] = optimisation.TravelEntry{From: e.From, To: e.To, Minutes: e.Minutes}
	}
	return result
}

// buildTravelDisplayLookup creates a map for travel time display.
func buildTravelDisplayLookup(entries []loader.TravelEntry) map[string]int {
	lookup := make(map[string]int, len(entries))
	for _, e := range entries {
		lookup[e.From+"|"+e.To] = e.Minutes
	}
	return lookup
}

// buildResourceLocationLookup reads starting location from each resource's details.
func buildResourceLocationLookup(resources []resource.Resource) map[string]string {
	lookup := make(map[string]string, len(resources))
	for _, res := range resources {
		var details struct {
			Location string `json:"location"`
		}
		if err := json.Unmarshal(res.Details(), &details); err == nil {
			lookup[res.ID()] = details.Location
		}
	}
	return lookup
}

// buildItemLocationLookup reads location from each event's details.
// Work item IDs are "WI-" + event ID.
func buildItemLocationLookup(events []event.BusinessEvent) map[string]string {
	lookup := make(map[string]string, len(events))
	for _, evt := range events {
		var details struct {
			Location string `json:"location"`
		}
		if err := json.Unmarshal(evt.Details(), &details); err == nil {
			lookup["WI-"+evt.ID()] = details.Location
		}
	}
	return lookup
}

func officialValidate(sc inrc2.Scenario, weekFiles []string, path []inrc2.BeamPath, initialHist inrc2.History) (int, int) {
	totalPenalty := 0
	totalViolations := 0
	valHist := initialHist
	for _, wp := range path {
		weekIdx := wp.Week - 1
		if weekIdx < 0 || weekIdx >= len(weekFiles) {
			continue
		}
		wd, err := inrc2.LoadWeekData(weekFiles[weekIdx])
		if err != nil {
			continue
		}
		result := inrc2.Score(sc, wd, valHist, wp.Solution)
		totalPenalty += result.SoftPenalty
		totalViolations += len(result.SoftDetails)
		valHist = inrc2.UpdateHistory(sc, valHist, wp.Solution)
	}
	return totalPenalty, totalViolations
}
