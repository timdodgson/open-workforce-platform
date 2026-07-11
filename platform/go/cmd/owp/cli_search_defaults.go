package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/ilp"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2"
)

// loadINRC2Instance resolves and loads a complete INRC-II instance by name or path.
func loadINRC2Instance(instanceName string) inrc2.InstanceBundle {
	bundle, err := inrc2.LoadInstanceBundle(instanceName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	return bundle
}

// requireHiGHS exits with install instructions if the HiGHS binary is not on PATH.
func requireHiGHS(extraHint string) {
	solver := &ilp.HighsSolver{}
	if solver.Available() {
		fmt.Println("  Solver found: ✓")
		fmt.Println()
		return
	}
	fmt.Fprintf(os.Stderr, "ERROR: HiGHS binary not found on PATH.\n")
	fmt.Fprintf(os.Stderr, "Install from: https://github.com/ERGO-Code/HiGHS/releases\n")
	if extraHint != "" {
		fmt.Fprintln(os.Stderr, extraHint)
	}
	os.Exit(1)
}

// parseSearchMode returns --mode or defaultVal when unset.
func parseSearchMode(args []string, defaultVal string) string {
	mode := parseStringFlag(args, "--mode")
	if mode == "" {
		return defaultVal
	}
	return mode
}

// parseSearchSeed returns --seed or defaultVal when unset or zero.
func parseSearchSeed(args []string, defaultVal int64) int64 {
	seed := int64(parseIntFlag(args, "--seed"))
	if seed == 0 {
		return defaultVal
	}
	return seed
}

// parseSearchIterations returns --iterations or defaultVal when unset or non-positive.
func parseSearchIterations(args []string, defaultVal int) int {
	iterations := parseIntFlag(args, "--iterations")
	if iterations <= 0 {
		return defaultVal
	}
	return iterations
}

// parseSearchTemperature returns --temperature or defaultVal when unset or non-positive.
func parseSearchTemperature(args []string, defaultVal float64) float64 {
	temperature := parseFloatFlag(args, "--temperature")
	if temperature <= 0 {
		return defaultVal
	}
	return temperature
}

// searchModeLabel returns the uppercase display label for a search mode.
func searchModeLabel(mode string) string {
	switch mode {
	case "lahc":
		return "LAHC"
	case "tabu":
		return "TABU"
	case "portfolio":
		return "PORTFOLIO"
	case "adaptive":
		return "ADAPTIVE"
	case "sa":
		return "SA"
	default:
		return strings.ToUpper(mode)
	}
}
