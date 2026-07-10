package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2"
)

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
