package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/ilp"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/s3upload"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

const (
	defaultS3Bucket = "pfrs-research-lab-data"
	defaultS3Region = "eu-west-1"
	runsDataRoot    = "../web/pfrs-lab/data/runs"
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
	cfg.Bucket = firstNonEmpty(
		parseStringFlag(args, "--"+primary+"s3-bucket"),
		parseStringFlag(args, "--"+alt+"s3-bucket"),
		defaultS3Bucket,
	)
	cfg.Region = firstNonEmpty(
		parseStringFlag(args, "--"+primary+"s3-region"),
		parseStringFlag(args, "--"+alt+"s3-region"),
		defaultS3Region,
	)
	return cfg
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
	outputDir := filepath.Join(runsDataRoot, runLabel)
	os.MkdirAll(outputDir, 0755)
	return outputDir
}

// writeJSONFile marshals v with indentation and writes to path (0644). Ignores write errors (matches prior behaviour).
func writeJSONFile(path string, v interface{}) {
	data, _ := json.MarshalIndent(v, "", "  ")
	os.WriteFile(path, data, 0644)
}

// writeRunMetadata writes run.json under outputDir.
func writeRunMetadata(outputDir string, meta map[string]interface{}) {
	writeJSONFile(filepath.Join(outputDir, "run.json"), meta)
}

// uploadRunOutput uploads a completed run directory to S3 when configured.
func uploadRunOutput(cfg storageConfig, runLabel, outputDir, algorithm string, penalty int) {
	s3upload.UploadRun(cfg.Mode, s3upload.UploadRunConfig{
		RunLabel:  runLabel,
		RunDir:    outputDir,
		Algorithm: algorithm,
		Penalty:   penalty,
		Bucket:    cfg.Bucket,
		Region:    cfg.Region,
	})
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

// pfrsWorkerIntelligence holds PFRS beam-search worker decision wiring.
type pfrsWorkerIntelligence struct {
	Engine           inrc2.WorkerDecisionEngine
	DecisionRecorder *inrc2.ShadowRecorder
	AssistRecorder   *inrc2.AssistRecorder
	AssistMode       bool
}

// wirePFRSWorkerIntelligence configures PFRS worker-level SI for tune-pfrs.
// Maps --worker-decision-mode to inrc2 engines and recorders:
//
//	shadow → DecisionRecorder (worker_decisions.csv)
//	assist/adaptive → AssistMode + AssistRecorder (worker_assist.csv)
//
// When --policy-mode is hybrid or learned, loads worker_policy.json from --policy-dir.
func wirePFRSWorkerIntelligence(mode, policyMode, policyDir string) pfrsWorkerIntelligence {
	var out pfrsWorkerIntelligence
	if mode == "shadow" || mode == "assist" || mode == "adaptive" {
		resolvedDir := optimisation.ResolvePolicyDir(policyMode, policyDir)
		if policyMode != "" && policyMode != "rules" {
			out.Engine = inrc2.NewHybridWorkerDecisionEngine(policyMode, resolvedDir)
			fmt.Printf("  Worker Policy: %s (dir: %s)\n", policyMode, resolvedDir)
		} else {
			out.Engine = inrc2.NewRuleBasedEngine()
		}
		out.DecisionRecorder = inrc2.NewShadowRecorder()
		if mode == "shadow" {
			fmt.Println("  Decision Mode: shadow (recording predictions, no behaviour change)")
		}
	}
	if mode == "assist" || mode == "adaptive" {
		out.AssistMode = true
		out.AssistRecorder = inrc2.NewAssistRecorder()
		if mode == "adaptive" {
			fmt.Println("  Decision Mode: adaptive (live-updating decisions, safety overrides active)")
		} else {
			fmt.Println("  Decision Mode: assist (AI advises optimiser, safety overrides active)")
		}
	}
	if mode == "shadow" || mode == "assist" || mode == "adaptive" {
		fmt.Println()
		os.Stdout.Sync()
	}
	return out
}

// resolveINRC2InstanceDir locates an INRC-II instance directory by path or short name.
func resolveINRC2InstanceDir(instanceName string) string {
	if _, err := os.Stat(instanceName); err == nil {
		return instanceName
	}
	for _, base := range []string{
		"../../examples/inrc2/testdatasets_json",
		"../../examples/inrc2/datasets_json",
	} {
		candidate := filepath.Join(base, instanceName)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	fmt.Fprintf(os.Stderr, "Instance not found: %s\n", instanceName)
	os.Exit(1)
	return ""
}

// scanINRC2Dir lists Sc-, WD-, and H0- files in an instance directory.
func scanINRC2Dir(dir string) (scenarioFile string, weekFiles, histFiles []string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading directory: %v\n", err)
		os.Exit(1)
	}
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, "Sc-") && strings.HasSuffix(name, ".json") {
			scenarioFile = filepath.Join(dir, name)
		} else if strings.HasPrefix(name, "WD-") && strings.HasSuffix(name, ".json") {
			weekFiles = append(weekFiles, filepath.Join(dir, name))
		} else if strings.HasPrefix(name, "H0-") && strings.HasSuffix(name, ".json") {
			histFiles = append(histFiles, filepath.Join(dir, name))
		}
	}
	sort.Strings(weekFiles)
	sort.Strings(histFiles)
	return scenarioFile, weekFiles, histFiles
}

// inrc2InstanceBundle holds loaded INRC-II instance data from a directory.
type inrc2InstanceBundle struct {
	Dir          string
	ScenarioFile string
	WeekFiles    []string
	HistFiles    []string
	Scenario     inrc2.Scenario
	History      inrc2.History
}

// loadINRC2Instance resolves and loads a complete INRC-II instance by name or path.
func loadINRC2Instance(instanceName string) inrc2InstanceBundle {
	dir := resolveINRC2InstanceDir(instanceName)
	scenarioFile, weekFiles, histFiles := scanINRC2Dir(dir)
	if scenarioFile == "" || len(weekFiles) == 0 || len(histFiles) == 0 {
		fmt.Fprintln(os.Stderr, "Incomplete instance data (need Sc-, WD-, H0- files)")
		os.Exit(1)
	}
	sc, err := inrc2.LoadScenario(scenarioFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	hist, err := inrc2.LoadHistory(histFiles[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	return inrc2InstanceBundle{
		Dir:          dir,
		ScenarioFile: scenarioFile,
		WeekFiles:    weekFiles,
		HistFiles:    histFiles,
		Scenario:     sc,
		History:      hist,
	}
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
