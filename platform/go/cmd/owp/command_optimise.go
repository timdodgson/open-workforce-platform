package main

import (
	"fmt"
	"os"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/application"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/loader"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

func runOptimise() {
	if len(os.Args) < 3 {
		printUsage()
		os.Exit(1)
	}

	path := os.Args[2]
	algorithm := parseAlgorithm(os.Args[3:])
	weightsProfile := parseWeights(os.Args[3:])
	profileName := parseProfile(os.Args[3:])

	// Validate weights profile.
	if _, ok := optimisation.GetWeightProfile(weightsProfile); !ok {
		fmt.Fprintf(os.Stderr, "Unknown weights profile: %s\n", weightsProfile)
		os.Exit(1)
	}

	// Validate algorithm profile.
	algProfile, ok := optimisation.GetProfile(profileName)
	if !ok {
		fmt.Fprintf(os.Stderr, "Unknown algorithm profile: %s\n", profileName)
		os.Exit(1)
	}

	// Apply explicit CLI overrides.
	algProfile = applyProfileOverrides(os.Args[3:], algProfile)

	dataset, err := loader.LoadDataset(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading dataset: %v\n", err)
		os.Exit(1)
	}

	result, err := application.OptimiseWithNRP(dataset.Events, dataset.Resources, convertTravel(dataset.TravelMatrix), dataset.NRPContext, algorithm, algProfile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error during optimisation: %v\n", err)
		os.Exit(1)
	}

	// Build the application-layer response (all business logic computed here).
	resp := application.BuildResponse(
		result,
		algorithm,
		buildCapacityLookup(dataset.Resources),
		buildDurationLookup(dataset.Events),
		buildItemLocationLookup(dataset.Events),
		buildResourceLocationLookup(dataset.Resources),
		buildTravelDisplayLookup(dataset.TravelMatrix),
	)

	// --- CLI presentation only below this line ---

	fmt.Println("=== Optimised Plan ===")
	fmt.Println()
	fmt.Printf("Algorithm: %s\n", resp.Algorithm)
	fmt.Printf("Profile: %s\n", profileName)
	displayEffectiveConfig(algorithm, algProfile)
	fmt.Printf("Assignment Score: %d\n", resp.AssignmentScore)
	fmt.Printf("Objective Score:  %d\n", resp.ObjectiveScore)
	fmt.Println()
	fmt.Println("Objective Breakdown:")
	for _, entry := range resp.ObjectiveBreakdown {
		fmt.Printf("  %s: %d\n", entry.Name, entry.Score)
	}
	fmt.Println()

	// Constraint Match Reporting.
	fmt.Println("Constraints:")
	fmt.Printf("  Hard: %d\n", resp.Constraints.HardCount)
	fmt.Printf("  Soft: %d\n", resp.Constraints.SoftCount)
	fmt.Printf("  Penalty: %d\n", resp.Constraints.TotalPenalty)
	if len(resp.Constraints.Summary) > 0 {
		fmt.Println("  Breakdown:")
		for _, s := range resp.Constraints.Summary {
			fmt.Printf("    %s: %d\n", s.Constraint, s.Count)
		}
	}
	fmt.Println()

	fmt.Printf("Resources: %d\n", len(resp.Resources))
	fmt.Printf("Capacity:  %d\n", resp.TotalCapacity)
	fmt.Println()
	fmt.Println("Assignments:")
	fmt.Println()

	for _, res := range resp.Resources {
		fmt.Printf("  %s\n", res.ResourceID)
		fmt.Printf("    Used: %d / %d mins\n", res.UsedMins, res.CapacityMins)
		fmt.Println("    Work Items:")
		for _, itemID := range res.WorkItems {
			fmt.Printf("      - %s\n", itemID)
		}
		fmt.Println()
	}

	if len(resp.Unassigned) > 0 {
		fmt.Println("Unassigned:")
		fmt.Println()
		for _, item := range resp.Unassigned {
			fmt.Printf("    %s\n", item.WorkItemID)
			if len(item.Reasons) > 0 {
				fmt.Println("      Reasons:")
				for _, reason := range item.Reasons {
					fmt.Printf("        - %s\n", reason)
				}
			}
		}
	} else {
		fmt.Println("Unassigned: None")
	}

	fmt.Println()

	// Hard violations.
	if resp.Constraints.HardCount > 0 {
		fmt.Println("Hard Violations:")
		fmt.Println()
		for _, m := range resp.Constraints.Matches {
			if m.Severity == "hard" {
				fmt.Printf("  [%s] %s\n", m.Constraint, m.Description)
			}
		}
		fmt.Println()
	}

	// Travel breakdown.
	fmt.Println("Travel:")
	fmt.Println()
	for _, rt := range resp.Travel {
		fmt.Printf("  %s\n", rt.ResourceID)
		for _, leg := range rt.Legs {
			fmt.Printf("    %s -> %s: %d mins\n", leg.From, leg.To, leg.Minutes)
		}
		fmt.Printf("    Total: %d mins\n", rt.TotalMins)
		fmt.Println()
	}

	// Statistics.
	fmt.Println("Optimisation Statistics:")
	fmt.Printf("  Algorithm: %s\n", resp.Statistics.Algorithm)
	fmt.Printf("  Duration: %dms\n", resp.Statistics.DurationMs)
	fmt.Printf("  Iterations: %d\n", resp.Statistics.Iterations)
	fmt.Printf("  Candidates Evaluated: %d\n", resp.Statistics.CandidatesEvaluated)
	fmt.Printf("  Improvements Accepted: %d\n", resp.Statistics.ImprovementsAccepted)
	fmt.Printf("  Final Objective Score: %d\n", resp.Statistics.FinalObjectiveScore)
	fmt.Println()

	fmt.Println("Done.")
}
