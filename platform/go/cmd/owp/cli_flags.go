package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
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

// parseWeights reads the --weights flag from remaining args.
// Defaults to "default".
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

// parseProfile reads the --profile flag from remaining args.
// Defaults to "default".
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

// parseTimeBudget reads the --time flag from args. Returns 0 if not supplied.
// Exits with error if --time is present but invalid.
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

// applyProfileOverrides parses per-algorithm CLI flags and overrides profile values.
func applyProfileOverrides(args []string, p optimisation.AlgorithmProfile) optimisation.AlgorithmProfile {
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

// displayEffectiveConfig prints the algorithm configuration being used.
func parseIntFlag(args []string, flag string) int {
	for i, arg := range args {
		if arg == flag && i+1 < len(args) {
			return atoiOrFail(args[i+1], flag)
		}
		if strings.HasPrefix(arg, flag+"=") {
			return atoiOrFail(strings.TrimPrefix(arg, flag+"="), flag)
		}
	}
	return 0
}

func parseFloatFlag(args []string, flag string) float64 {
	for i, arg := range args {
		var val string
		if arg == flag && i+1 < len(args) {
			val = args[i+1]
		} else if strings.HasPrefix(arg, flag+"=") {
			val = strings.TrimPrefix(arg, flag+"=")
		}
		if val != "" {
			f := parseFloat(val, flag)
			return f
		}
	}
	return 0
}

func parseBoolFlag(args []string, flag string) string {
	for i, arg := range args {
		if arg == flag && i+1 < len(args) {
			v := strings.TrimSpace(args[i+1])
			if v != "true" && v != "false" {
				fmt.Fprintf(os.Stderr, "Invalid %s value: %s (must be true or false)\n", flag, v)
				os.Exit(1)
			}
			return v
		}
		if strings.HasPrefix(arg, flag+"=") {
			v := strings.TrimSpace(strings.TrimPrefix(arg, flag+"="))
			if v != "true" && v != "false" {
				fmt.Fprintf(os.Stderr, "Invalid %s value: %s (must be true or false)\n", flag, v)
				os.Exit(1)
			}
			return v
		}
	}
	return ""
}

func parseStringFlag(args []string, flag string) string {
	for i, arg := range args {
		if arg == flag && i+1 < len(args) {
			return strings.TrimSpace(args[i+1])
		}
		if strings.HasPrefix(arg, flag+"=") {
			return strings.TrimSpace(strings.TrimPrefix(arg, flag+"="))
		}
	}
	return ""
}

// parsePFRSConfig reads all --pfrs-* flags from args and returns a PFRSConfig.
func parsePFRSConfig(args []string) inrc2.PFRSConfig {
	config := inrc2.DefaultPFRSConfig()
	if v := parseStringFlag(args, "--pfrs-mode"); v != "" {
		if v != "sa" && v != "lahc" && v != "tabu" && v != "portfolio" {
			fmt.Fprintf(os.Stderr, "Invalid --pfrs-mode: %s (must be sa, lahc, tabu, or portfolio)\n", v)
			os.Exit(1)
		}
		config.Mode = v
	}
	if v := parseIntFlag(args, "--pfrs-iterations-per-worker"); v > 0 {
		config.IterationsPerWorker = v
	}
	if v := parseIntFlag(args, "--pfrs-max-concurrent-workers"); v > 0 {
		config.MaxConcurrentWorkers = v
	}
	if v := parseIntFlag(args, "--pfrs-max-total-workers"); v > 0 {
		config.MaxTotalWorkers = v
	}
	if v := parseFloatFlag(args, "--pfrs-initial-temperature"); v > 0 {
		config.InitialTemperature = v
	}
	if v := parseFloatFlag(args, "--pfrs-cooling-rate"); v > 0 {
		config.CoolingRate = v
	}
	if v := parseFloatFlag(args, "--pfrs-min-temperature"); v > 0 {
		config.MinTemperature = v
	}
	if v := parseIntFlag(args, "--pfrs-late-acceptance-length"); v > 0 {
		config.LateAcceptanceLength = v
	}
	if v := parseIntFlag(args, "--pfrs-tabu-tenure"); v > 0 {
		config.TabuTenure = v
	}
	if v := parseStringFlag(args, "--pfrs-portfolio"); v != "" {
		config.Portfolio = strings.Split(v, ",")
		if config.Mode != "portfolio" {
			config.Mode = "portfolio"
		}
	}
	if v := parseIntFlag(args, "--pfrs-branch-cooldown"); v > 0 {
		config.BranchCooldown = v
	}
	if v := parseIntFlag(args, "--pfrs-seed"); v > 0 {
		config.Seed = int64(v)
	}
	if v := parseBoolFlag(args, "--pfrs-deterministic"); v != "" {
		config.Deterministic = v == "true"
	}
	if v := parseStringFlag(args, "--pfrs-scoring-mode"); v != "" {
		if v != "official-penalty" && v != "soft-violation-count" {
			fmt.Fprintf(os.Stderr, "Invalid --pfrs-scoring-mode: %s (must be official-penalty or soft-violation-count)\n", v)
			os.Exit(1)
		}
		config.ScoringMode = v
	}
	if v := parseStringFlag(args, "--pfrs-cooling-mode"); v != "" {
		if v != "adaptive" && v != "fixed-rate" {
			fmt.Fprintf(os.Stderr, "Invalid --pfrs-cooling-mode: %s (must be adaptive or fixed-rate)\n", v)
			os.Exit(1)
		}
		config.CoolingMode = v
	}
	return config
}

func atoiOrFail(s, flag string) int {
	s = strings.TrimSpace(s)
	n := 0
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			fmt.Fprintf(os.Stderr, "Invalid %s value: %s (must be a positive integer)\n", flag, s)
			os.Exit(1)
		}
		n = n*10 + int(ch-'0')
	}
	if n <= 0 {
		fmt.Fprintf(os.Stderr, "Invalid %s value: %s (must be a positive integer)\n", flag, s)
		os.Exit(1)
	}
	return n
}

// parseSeedList parses a comma-separated list of positive integers.
func parseSeedList(s string) []int64 {
	parts := strings.Split(s, ",")
	var seeds []int64
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		n := int64(0)
		for _, ch := range p {
			if ch < '0' || ch > '9' {
				fmt.Fprintf(os.Stderr, "Invalid seed value: %s (must be a positive integer)\n", p)
				os.Exit(1)
			}
			n = n*10 + int64(ch-'0')
		}
		if n <= 0 {
			fmt.Fprintf(os.Stderr, "Invalid seed value: %s (must be a positive integer)\n", p)
			os.Exit(1)
		}
		seeds = append(seeds, n)
	}
	if len(seeds) == 0 {
		fmt.Fprintln(os.Stderr, "No valid seeds provided")
		os.Exit(1)
	}
	return seeds
}

func parseFloat(s, flag string) float64 {
	s = strings.TrimSpace(s)
	// Simple float parser: digits, optional dot, digits.
	var result float64
	var decimal float64
	inDecimal := false
	divisor := 1.0
	for _, ch := range s {
		if ch == '.' {
			if inDecimal {
				fmt.Fprintf(os.Stderr, "Invalid %s value: %s (must be a number)\n", flag, s)
				os.Exit(1)
			}
			inDecimal = true
			continue
		}
		if ch < '0' || ch > '9' {
			fmt.Fprintf(os.Stderr, "Invalid %s value: %s (must be a number)\n", flag, s)
			os.Exit(1)
		}
		if inDecimal {
			divisor *= 10
			decimal += float64(ch-'0') / divisor
		} else {
			result = result*10 + float64(ch-'0')
		}
	}
	result += decimal
	if result <= 0 {
		fmt.Fprintf(os.Stderr, "Invalid %s value: %s (must be a positive number)\n", flag, s)
		os.Exit(1)
	}
	return result
}

// parseShowInvalidFlag returns true when --show-invalid is present.
func parseShowInvalidFlag(args []string) bool {
	for _, arg := range args {
		if arg == "--show-invalid" {
			return true
		}
	}
	return false
}

// requireInstanceFlag exits with usage when --instance is missing.
func requireInstanceFlag(args []string, usage string) string {
	instancePath := parseStringFlag(args, "--instance")
	if instancePath == "" {
		fmt.Fprintln(os.Stderr, "Error: --instance <path> is required")
		if usage != "" {
			fmt.Fprintln(os.Stderr, usage)
		}
		os.Exit(1)
	}
	return instancePath
}

// buildCapacityLookup reads capacity from each resource's details for display.
