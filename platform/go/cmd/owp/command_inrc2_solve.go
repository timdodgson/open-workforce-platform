package main

import (
	"fmt"
	"os"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

func runSolveINRC2() {
	if len(os.Args) < 6 {
		fmt.Fprintln(os.Stderr, "Usage: owp solve-inrc2 <scenario-file> <week-file> <history-file> <solution-output-file> [--algorithm tabu-search] [--profile default]")
		os.Exit(1)
	}

	scenarioPath := os.Args[2]
	weekPath := os.Args[3]
	historyPath := os.Args[4]
	outputPath := os.Args[5]
	flagArgs := os.Args[5:]

	algorithm := parseAlgorithm(flagArgs)
	profileName := parseProfile(flagArgs)

	algProfile, ok := optimisation.GetProfile(profileName)
	if !ok {
		fmt.Fprintf(os.Stderr, "Unknown profile: %s\n", profileName)
		os.Exit(1)
	}
	algProfile = applyProfileOverrides(flagArgs, algProfile)

	sc, wd, hist, _, err := loadINRC2WeekInputs(scenarioPath, weekPath, historyPath, "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	pfrsConfig := parsePFRSConfig(flagArgs)
	out, err := inrc2.SolveSingleWeek(sc, wd, hist, inrc2.WeekSolveParams{
		Algorithm:  algorithm,
		AlgProfile: algProfile,
		PFRSConfig: pfrsConfig,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error solving: %v\n", err)
		os.Exit(1)
	}

	if err := inrc2.WriteSolution(out.Solution, outputPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing solution: %v\n", err)
		os.Exit(1)
	}

	printINRC2SolveResult(sc, algorithm, out, outputPath)
}
