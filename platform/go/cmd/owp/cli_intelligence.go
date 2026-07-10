package main

import (
	"fmt"
	"os"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

// searchIntelligenceOpts controls optional stdout from SI flag application.
type searchIntelligenceOpts struct {
	PrintPolicyDir bool
}

// applySearchIntelligenceFlags validates and applies --worker-decision-mode and --policy-mode.
func applySearchIntelligenceFlags(args []string, config *optimisation.SearchConfig, opts searchIntelligenceOpts) string {
	workerDecisionMode := parseStringFlag(args, "--worker-decision-mode")
	if workerDecisionMode != "" && workerDecisionMode != "off" && workerDecisionMode != "shadow" && workerDecisionMode != "assist" && workerDecisionMode != "adaptive" {
		fmt.Fprintf(os.Stderr, "Error: --worker-decision-mode must be off, shadow, assist, or adaptive (got %q)\n", workerDecisionMode)
		os.Exit(1)
	}
	if workerDecisionMode == "shadow" || workerDecisionMode == "assist" || workerDecisionMode == "adaptive" {
		config.AssistMode = workerDecisionMode
		config.AssistConfig = optimisation.DefaultSearchAssistConfig()
		fmt.Printf("  Decision Mode: %s\n", workerDecisionMode)
	}

	policyMode := parseStringFlag(args, "--policy-mode")
	policyDir := parseStringFlag(args, "--policy-dir")
	if policyMode != "" && policyMode != "rules" && policyMode != "hybrid" && policyMode != "learned" {
		fmt.Fprintf(os.Stderr, "Error: --policy-mode must be rules, hybrid, or learned (got %q)\n", policyMode)
		os.Exit(1)
	}
	if policyMode != "" {
		config.PolicyMode = policyMode
		config.PolicyDir = optimisation.ResolvePolicyDir(policyMode, policyDir)
		fmt.Printf("  Policy Mode: %s\n", policyMode)
		if opts.PrintPolicyDir {
			fmt.Printf("  Policy Dir: %s\n", config.PolicyDir)
		}
		if config.AssistMode == "" {
			config.AssistMode = "shadow"
			config.AssistConfig = optimisation.DefaultSearchAssistConfig()
		}
	}
	return workerDecisionMode
}

// applyPFRSIntelligenceFlags validates --worker-decision-mode and --policy-mode for tune-pfrs.
func applyPFRSIntelligenceFlags(args []string) (workerDecisionMode, policyMode, policyDir string) {
	workerDecisionMode = parseStringFlag(args, "--worker-decision-mode")
	if workerDecisionMode != "" && workerDecisionMode != "off" && workerDecisionMode != "shadow" && workerDecisionMode != "assist" && workerDecisionMode != "adaptive" {
		fmt.Fprintf(os.Stderr, "Error: --worker-decision-mode must be off, shadow, assist, or adaptive (got %q)\n", workerDecisionMode)
		os.Exit(1)
	}
	policyMode = parseStringFlag(args, "--policy-mode")
	policyDir = parseStringFlag(args, "--policy-dir")
	if policyMode != "" && policyMode != "rules" && policyMode != "hybrid" && policyMode != "learned" {
		fmt.Fprintf(os.Stderr, "Error: --policy-mode must be rules, hybrid, or learned (got %q)\n", policyMode)
		os.Exit(1)
	}
	return workerDecisionMode, policyMode, policyDir
}

// pfrsWorkerIntelligence is an alias for the inrc2 wire result used by tune-pfrs.
type pfrsWorkerIntelligence = inrc2.WorkerIntelligenceWire

// wirePFRSWorkerIntelligence configures PFRS worker-level SI for tune-pfrs.
func wirePFRSWorkerIntelligence(mode, policyMode, policyDir string) pfrsWorkerIntelligence {
	wire, lines := inrc2.WireWorkerIntelligence(mode, policyMode, policyDir)
	for _, line := range lines {
		fmt.Println(line)
	}
	if len(lines) > 0 {
		fmt.Println()
		os.Stdout.Sync()
	}
	return wire
}
