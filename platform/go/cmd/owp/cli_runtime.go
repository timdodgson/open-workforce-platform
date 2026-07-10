package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/ilp"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/runoutput"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

// storageConfig holds S3 upload settings parsed from CLI flags.
type storageConfig struct {
	Mode   string
	Bucket string
	Region string
}

// parseStorageConfig reads storage flags. When pfrsPrefix is true, prefers --pfrs-storage etc.
// Both --storage and --pfrs-storage are accepted on all commands.
func parseStorageConfig(args []string, pfrsPrefix bool) storageConfig {
	primary := ""
	alt := "pfrs-"
	if pfrsPrefix {
		primary = "pfrs-"
		alt = ""
	}
	cfg := storageConfig{
		Mode: firstNonEmpty(
			parseStringFlag(args, "--"+primary+"storage"),
			parseStringFlag(args, "--"+alt+"storage"),
		),
	}
	ro := runoutput.StorageConfig{
		Bucket: firstNonEmpty(
			parseStringFlag(args, "--"+primary+"s3-bucket"),
			parseStringFlag(args, "--"+alt+"s3-bucket"),
		),
		Region: firstNonEmpty(
			parseStringFlag(args, "--"+primary+"s3-region"),
			parseStringFlag(args, "--"+alt+"s3-region"),
		),
	}.WithDefaults()
	cfg.Bucket = ro.Bucket
	cfg.Region = ro.Region
	return cfg
}

func runoutputStorage(cfg storageConfig) runoutput.StorageConfig {
	return runoutput.StorageConfig{
		Mode:   cfg.Mode,
		Bucket: cfg.Bucket,
		Region: cfg.Region,
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// parseRunLabelFlag reads --run-label or --pfrs-run-label.
func parseRunLabelFlag(args []string, pfrsPrefix bool) string {
	if pfrsPrefix {
		return parseStringFlag(args, "--pfrs-run-label")
	}
	return parseStringFlag(args, "--run-label")
}

// ensureRunOutputDir creates platform/web/pfrs-lab/data/runs/<label>/ and returns the path.
func ensureRunOutputDir(runLabel string) string {
	return runoutput.EnsureDir(runLabel)
}

// writeJSONFile marshals v with indentation and writes to path (0644). Ignores write errors (matches prior behaviour).
func writeJSONFile(path string, v interface{}) {
	data, _ := json.MarshalIndent(v, "", "  ")
	os.WriteFile(path, data, 0644)
}

// writeRunMetadata writes run.json under outputDir.
func writeRunMetadata(outputDir string, meta map[string]interface{}) {
	_ = runoutput.WriteRunMetadata(outputDir, meta)
}

// uploadRunOutput uploads a completed run directory to S3 when configured.
func uploadRunOutput(cfg storageConfig, runLabel, outputDir, algorithm string, penalty int) {
	_ = runoutput.Upload(runoutputStorage(cfg), runLabel, outputDir, algorithm, penalty)
}

type portfolioRunParams struct {
	Problem            optimisation.Problem
	Config             optimisation.SearchConfig
	WorkerDecisionMode string
	Domain             string
	Instance           string
	PortfolioModelPath string
}

// runSearchOrPortfolio runs portfolio assist or a single search depending on mode.
// extraPortfolioModes lists additional modes treated as portfolio (e.g. JSS "adaptive").
func runSearchOrPortfolio(mode string, extraPortfolioModes []string, p portfolioRunParams) (optimisation.SearchResult, *optimisation.PortfolioAssistRecorder) {
	usePortfolio := mode == "portfolio"
	for _, m := range extraPortfolioModes {
		if mode == m {
			usePortfolio = true
			break
		}
	}
	if !usePortfolio {
		return optimisation.RunSearch(p.Problem, p.Config), nil
	}
	assistConfig := optimisation.PortfolioAssistConfig{
		Mode:      p.WorkerDecisionMode,
		Domain:    p.Domain,
		Instance:  p.Instance,
		ModelPath: p.PortfolioModelPath,
	}
	pr, recorder := optimisation.RunPortfolioWithAssist(p.Problem, p.Config, assistConfig)
	return pr.BestResult, recorder
}

// searchIntelligenceOpts controls optional stdout from SI flag application.
type searchIntelligenceOpts struct {
	PrintPolicyDir bool
}

// applySearchIntelligenceFlags validates and applies --worker-decision-mode and --policy-mode.
// --worker-decision-mode sets SearchConfig.AssistMode for SearchAssist and PortfolioAssist
// on CVRP/JSS/VRPTW (modes: off, shadow, assist, adaptive).
// --policy-mode sets SearchConfig.PolicyMode for SI 2.0 (rules, hybrid, learned).
// Returns the worker decision mode string for portfolio-mode wiring.
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
	default:
		return "SA"
	}
}
