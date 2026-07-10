package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2/legacysearch"
)

func parseAlgorithm(args []string) string {
	for i, arg := range args {
		if arg == "--algorithm" && i+1 < len(args) {
			return strings.TrimSpace(args[i+1])
		}
		if strings.HasPrefix(arg, "--algorithm=") {
			return strings.TrimSpace(strings.TrimPrefix(arg, "--algorithm="))
		}
	}
	return "constructive"
}

func parseWeights(args []string) string {
	for i, arg := range args {
		if arg == "--weights" && i+1 < len(args) {
			return strings.TrimSpace(args[i+1])
		}
		if strings.HasPrefix(arg, "--weights=") {
			return strings.TrimSpace(strings.TrimPrefix(arg, "--weights="))
		}
	}
	return "default"
}

func parseProfile(args []string) string {
	for i, arg := range args {
		if arg == "--profile" && i+1 < len(args) {
			return strings.TrimSpace(args[i+1])
		}
		if strings.HasPrefix(arg, "--profile=") {
			return strings.TrimSpace(strings.TrimPrefix(arg, "--profile="))
		}
	}
	return "default"
}

func parseTimeBudget(args []string) int {
	for i, arg := range args {
		var val string
		if arg == "--time" && i+1 < len(args) {
			val = strings.TrimSpace(args[i+1])
		} else if strings.HasPrefix(arg, "--time=") {
			val = strings.TrimSpace(strings.TrimPrefix(arg, "--time="))
		}
		if val != "" {
			n := 0
			for _, ch := range val {
				if ch < '0' || ch > '9' {
					fmt.Fprintf(os.Stderr, "Invalid --time value: %s (must be a positive integer)\n", val)
					os.Exit(1)
				}
				n = n*10 + int(ch-'0')
			}
			if n <= 0 {
				fmt.Fprintf(os.Stderr, "Invalid --time value: %s (must be a positive integer)\n", val)
				os.Exit(1)
			}
			return n
		}
	}
	return 0
}

func applyProfileOverrides(args []string, p legacysearch.AlgorithmProfile) legacysearch.AlgorithmProfile {
	if v := parseIntFlag(args, "--hc-max-iterations"); v > 0 {
		p.HCMaxIterations = v
	}
	if v := parseIntFlag(args, "--sa-max-iterations"); v > 0 {
		p.SAMaxIterations = v
	}
	if v := parseFloatFlag(args, "--sa-initial-temperature"); v > 0 {
		p.SAInitialTemperature = v
	}
	if v := parseFloatFlag(args, "--sa-cooling-rate"); v > 0 {
		p.SACoolingRate = v
	}
	if v := parseFloatFlag(args, "--sa-min-temperature"); v > 0 {
		p.SAMinTemperature = v
	}
	if v := parseIntFlag(args, "--tabu-max-iterations"); v > 0 {
		p.TabuMaxIterations = v
	}
	if v := parseIntFlag(args, "--tabu-list-size"); v > 0 {
		p.TabuListSize = v
	}
	if v := parseBoolFlag(args, "--tabu-aspiration"); v != "" {
		p.TabuAspirationEnabled = v == "true"
	}
	if v := parseIntFlag(args, "--lns-iterations"); v > 0 {
		p.LNSIterations = v
	}
	if v := parseIntFlag(args, "--lns-destroy-size"); v > 0 {
		p.LNSDestroySize = v
	}
	if v := parseStringFlag(args, "--lns-repair-strategy"); v != "" {
		if v != "greedy" && v != "priority" && v != "best-fit" {
			fmt.Fprintf(os.Stderr, "Invalid --lns-repair-strategy: %s (must be greedy, priority, or best-fit)\n", v)
			os.Exit(1)
		}
		p.LNSRepairStrategy = v
	}
	return p
}
