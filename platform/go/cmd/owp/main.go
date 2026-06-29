package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/application"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/event"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/resource"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/loader"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/nrp"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "optimise":
		runOptimise()
	case "benchmark":
		runBenchmark()
	case "convert-nrp":
		runConvertNRP()
	case "validate-inrc2":
		runValidateINRC2()
	case "solve-inrc2":
		runSolveINRC2()
	case "benchmark-inrc2":
		runBenchmarkINRC2()
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  owp optimise <dataset-path> [--algorithm constructive|hill-climbing|simulated-annealing|tabu-search|large-neighbourhood-search] [--weights default]")
	fmt.Fprintln(os.Stderr, "  owp benchmark <datasets-directory>")
	fmt.Fprintln(os.Stderr, "  owp convert-nrp <nrp-input> <output-dataset>")
	fmt.Fprintln(os.Stderr, "  owp validate-inrc2 <scenario-file> <week-file> <history-file> <solution-file>")
	fmt.Fprintln(os.Stderr, "  owp solve-inrc2 <scenario-file> <week-file> <history-file> <solution-output-file> [--algorithm tabu-search] [--profile default]")
	fmt.Fprintln(os.Stderr, "  owp benchmark-inrc2 [instance-name] [--profile research]")
}

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

func runBenchmark() {
	if len(os.Args) < 3 {
		printUsage()
		os.Exit(1)
	}

	dir := os.Args[2]
	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading directory: %v\n", err)
		os.Exit(1)
	}

	// Discover dataset files.
	var datasetFiles []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
			datasetFiles = append(datasetFiles, e.Name())
		}
	}

	if len(datasetFiles) == 0 {
		fmt.Fprintln(os.Stderr, "No dataset files found in directory")
		os.Exit(1)
	}

	// Get available algorithms.
	algs := optimisation.Available()
	sort.Strings(algs)

	// Print header.
	fmt.Printf("%-28s %-26s %7s %11s %8s %9s %10s %6s %6s %10s %12s\n",
		"Dataset", "Algorithm", "Score", "Objective", "Delta", "Delta %", "Assigned", "Hard", "Soft", "Duration", "Candidates")
	fmt.Println(strings.Repeat("-", 145))

	// Run each combination.
	type benchResult struct {
		alg        string
		score      int
		objective  int
		assigned   int
		hard       int
		soft       int
		duration   int64
		candidates int
	}

	// Aggregate stats per algorithm for summary.
	type algStats struct {
		count      int
		totalObj   int
		totalDelta int
		totalCands int
	}
	summary := make(map[string]*algStats)

	for _, file := range datasetFiles {
		path := filepath.Join(dir, file)
		dataset, err := loader.LoadDataset(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  ERROR loading %s: %v\n", file, err)
			continue
		}

		travel := convertTravel(dataset.TravelMatrix)
		name := strings.TrimSuffix(file, ".json")

		// Run all algorithms and collect results.
		var results []benchResult
		baseline := 0

		for _, alg := range algs {
			result, err := application.OptimiseWithNRP(dataset.Events, dataset.Resources, travel, dataset.NRPContext, alg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  ERROR %s/%s: %v\n", name, alg, err)
				continue
			}

			stats := result.Statistics()
			br := benchResult{
				alg:        alg,
				score:      result.Score(),
				objective:  result.ObjectiveScore(),
				assigned:   result.Size(),
				hard:       result.HardConstraintCount(),
				soft:       result.SoftConstraintCount(),
				duration:   stats.DurationMs,
				candidates: stats.CandidatesEvaluated,
			}
			results = append(results, br)

			if alg == "constructive" {
				baseline = br.objective
			}
		}

		// Print results with delta.
		for _, br := range results {
			delta := br.objective - baseline
			deltaStr := "0"
			if delta > 0 {
				deltaStr = fmt.Sprintf("+%d", delta)
			} else if delta < 0 {
				deltaStr = fmt.Sprintf("%d", delta)
			}

			pctStr := "0.0%"
			if baseline == 0 {
				pctStr = "n/a"
			} else if delta != 0 {
				pct := float64(delta) / float64(baseline) * 100
				if pct > 0 {
					pctStr = fmt.Sprintf("+%.1f%%", pct)
				} else {
					pctStr = fmt.Sprintf("%.1f%%", pct)
				}
			}

			fmt.Printf("%-28s %-26s %7d %11d %8s %9s %10d %6d %6d %8dms %12d\n",
				name, br.alg, br.score, br.objective, deltaStr, pctStr,
				br.assigned, br.hard, br.soft, br.duration, br.candidates)

			// Accumulate for summary.
			if summary[br.alg] == nil {
				summary[br.alg] = &algStats{}
			}
			s := summary[br.alg]
			s.count++
			s.totalObj += br.objective
			s.totalDelta += (br.objective - baseline)
			s.totalCands += br.candidates
		}
	}

	// Print summary.
	fmt.Println()
	fmt.Println("Benchmark Summary:")
	fmt.Println()
	fmt.Printf("%-28s %10s %15s %11s %13s %12s\n",
		"Algorithm", "Datasets", "Avg Objective", "Avg Delta", "Avg Delta %", "Candidates")
	fmt.Println(strings.Repeat("-", 92))

	// Get constructive average objective for percentage calculation.
	constructiveAvgObj := 0
	if cs, ok := summary["constructive"]; ok && cs.count > 0 {
		constructiveAvgObj = cs.totalObj / cs.count
	}

	for _, alg := range algs {
		s, ok := summary[alg]
		if !ok || s.count == 0 {
			continue
		}

		avgObj := s.totalObj / s.count
		avgDelta := s.totalDelta / s.count

		avgDeltaStr := "0"
		if avgDelta > 0 {
			avgDeltaStr = fmt.Sprintf("+%d", avgDelta)
		} else if avgDelta < 0 {
			avgDeltaStr = fmt.Sprintf("%d", avgDelta)
		}

		pctStr := "0.0%"
		if constructiveAvgObj > 0 && avgDelta != 0 {
			pct := float64(avgDelta) / float64(constructiveAvgObj) * 100
			if pct > 0 {
				pctStr = fmt.Sprintf("+%.1f%%", pct)
			} else {
				pctStr = fmt.Sprintf("%.1f%%", pct)
			}
		}

		fmt.Printf("%-28s %10d %15d %11s %13s %12d\n",
			alg, s.count, avgObj, avgDeltaStr, pctStr, s.totalCands)
	}
}

func runConvertNRP() {
	if len(os.Args) < 4 {
		fmt.Fprintln(os.Stderr, "Usage: owp convert-nrp <nrp-input> <output-dataset>")
		os.Exit(1)
	}

	inputPath := os.Args[2]
	outputPath := os.Args[3]

	input, err := nrp.LoadNRP(inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading NRP file: %v\n", err)
		os.Exit(1)
	}

	dataset := nrp.Convert(input)

	if err := nrp.WriteDataset(dataset, outputPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing dataset: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Converted %d nurses and %d shift demands into OWP dataset.\n", len(input.Nurses), len(dataset.BusinessEvents))
	fmt.Printf("Output: %s\n", outputPath)
}

// parseAlgorithm reads the --algorithm flag from remaining args.
// Defaults to "constructive".
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
		if v != "sa" && v != "lahc" {
			fmt.Fprintf(os.Stderr, "Invalid --pfrs-mode: %s (must be sa or lahc)\n", v)
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

// buildCapacityLookup reads capacity from each resource's details for display.
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

func runValidateINRC2() {
	if len(os.Args) < 6 {
		fmt.Fprintln(os.Stderr, "Usage: owp validate-inrc2 <scenario-file> <week-file> <history-file> <solution-file>")
		os.Exit(1)
	}

	scenarioPath := os.Args[2]
	weekPath := os.Args[3]
	historyPath := os.Args[4]
	solutionPath := os.Args[5]

	sc, err := inrc2.LoadScenario(scenarioPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading scenario: %v\n", err)
		os.Exit(1)
	}

	wd, err := inrc2.LoadWeekData(weekPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading week data: %v\n", err)
		os.Exit(1)
	}

	hist, err := inrc2.LoadHistory(historyPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading history: %v\n", err)
		os.Exit(1)
	}

	sol, err := inrc2.LoadSolution(solutionPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading solution: %v\n", err)
		os.Exit(1)
	}

	result := inrc2.Score(sc, wd, hist, sol)

	fmt.Println("=== INRC-II Validation ===")
	fmt.Println()
	fmt.Printf("Scenario: %s\n", sc.ID)
	fmt.Printf("Week: %d\n", sol.Week)
	fmt.Printf("Assignments: %d\n", len(sol.Assignments))
	fmt.Println()

	fmt.Printf("Hard Violations: %d\n", result.HardViolations)
	if result.HardViolations > 0 {
		for _, v := range result.HardDetails {
			fmt.Printf("  [%s] %s (nurse=%s, day=%s)\n", v.Code, v.Message, v.Nurse, inrc2.DayName(v.Day))
		}
	}
	fmt.Println()

	fmt.Printf("Soft Penalty: %d\n", result.SoftPenalty)
	if len(result.SoftDetails) > 0 {
		fmt.Println("  Breakdown:")
		for _, d := range result.SoftDetails {
			if d.Nurse != "" {
				fmt.Printf("    [%s] nurse=%s penalty=%d\n", d.Constraint, d.Nurse, d.Penalty)
			} else {
				fmt.Printf("    [%s] penalty=%d\n", d.Constraint, d.Penalty)
			}
		}
	}
	fmt.Println()

	fmt.Printf("Total Objective: %d\n", result.TotalObjective)
	fmt.Println()
	fmt.Println("Done.")
}

func runSolveINRC2() {
	if len(os.Args) < 6 {
		fmt.Fprintln(os.Stderr, "Usage: owp solve-inrc2 <scenario-file> <week-file> <history-file> <solution-output-file> [--algorithm tabu-search] [--profile default]")
		os.Exit(1)
	}

	scenarioPath := os.Args[2]
	weekPath := os.Args[3]
	historyPath := os.Args[4]
	outputPath := os.Args[5]
	algorithm := parseAlgorithm(os.Args[5:])
	profileName := parseProfile(os.Args[5:])

	algProfile, ok := optimisation.GetProfile(profileName)
	if !ok {
		fmt.Fprintf(os.Stderr, "Unknown profile: %s\n", profileName)
		os.Exit(1)
	}
	algProfile = applyProfileOverrides(os.Args[5:], algProfile)

	sc, err := inrc2.LoadScenario(scenarioPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading scenario: %v\n", err)
		os.Exit(1)
	}

	wd, err := inrc2.LoadWeekData(weekPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading week data: %v\n", err)
		os.Exit(1)
	}

	hist, err := inrc2.LoadHistory(historyPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading history: %v\n", err)
		os.Exit(1)
	}

	var sol inrc2.Solution
	var scoreResult inrc2.ScoreResult

	if algorithm == "parallel-feasible-roster-search" {
		pfrsConfig := parsePFRSConfig(os.Args[5:])
		pfrsSol, pfrsStats, pfrsScore, err := inrc2.SolveWeekPFRS(sc, wd, hist, pfrsConfig)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error solving: %v\n", err)
			os.Exit(1)
		}
		sol = pfrsSol
		scoreResult = pfrsScore

		if err := inrc2.WriteSolution(sol, outputPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing solution: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("=== INRC-II Solve (PFRS) ===\n\n")
		fmt.Printf("Scenario: %s\n", sc.ID)
		fmt.Printf("Week: %d\n", sol.Week)
		fmt.Printf("Algorithm: parallel-feasible-roster-search\n")
		fmt.Printf("Mode: %s\n", pfrsConfig.Mode)
		fmt.Printf("Assignments: %d\n", len(sol.Assignments))
		fmt.Printf("Hard Violations: %d\n", scoreResult.HardViolations)
		fmt.Printf("Soft Penalty: %d\n", scoreResult.SoftPenalty)
		fmt.Printf("Workers Started: %d\n", pfrsStats.WorkersStarted)
		fmt.Printf("Branches Created: %d\n", pfrsStats.BranchesCreated)
		fmt.Printf("Best Updates: %d\n", pfrsStats.BestUpdates)
		fmt.Printf("Invalid Moves Rejected: %d\n", pfrsStats.InvalidMovesRejected)
		fmt.Printf("Iterations: %d\n", pfrsStats.TotalIterations)
		fmt.Printf("Candidates Evaluated: %d\n", pfrsStats.CandidatesEvaluated)
		fmt.Printf("Duration: %dms\n", pfrsStats.DurationMs)
		fmt.Printf("Output: %s\n", outputPath)
		fmt.Println("\nDone.")
		return
	}

	owpSol, _, err := inrc2.SolveWeek(sc, wd, hist, algorithm, algProfile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error solving: %v\n", err)
		os.Exit(1)
	}
	sol = owpSol

	if err := inrc2.WriteSolution(sol, outputPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing solution: %v\n", err)
		os.Exit(1)
	}

	// Score and display.
	scoreResult = inrc2.Score(sc, wd, hist, sol)

	fmt.Printf("=== INRC-II Solve ===\n\n")
	fmt.Printf("Scenario: %s\n", sc.ID)
	fmt.Printf("Week: %d\n", sol.Week)
	fmt.Printf("Algorithm: %s\n", algorithm)
	fmt.Printf("Assignments: %d\n", len(sol.Assignments))
	fmt.Printf("Hard Violations: %d\n", scoreResult.HardViolations)
	fmt.Printf("Soft Penalty: %d\n", scoreResult.SoftPenalty)
	fmt.Printf("Output: %s\n", outputPath)
	fmt.Println("\nDone.")
}

func runBenchmarkINRC2() {
	// Default instance and profile.
	defaultInstance := "n012w8"
	defaultProfile := "research"
	defaultBasePath := "../../examples/inrc2/testdatasets_json"

	// Resolve instance directory.
	var dir string
	profileName := defaultProfile

	if len(os.Args) >= 3 && !strings.HasPrefix(os.Args[2], "--") {
		// Explicit instance name or path supplied.
		arg := os.Args[2]
		if _, err := os.Stat(arg); err == nil {
			dir = arg
		} else {
			// Try as instance name under test datasets.
			candidate := filepath.Join(defaultBasePath, arg)
			if _, err := os.Stat(candidate); err == nil {
				dir = candidate
			} else {
				// Try competition datasets.
				candidate = filepath.Join("../../examples/inrc2/datasets_json", arg)
				if _, err := os.Stat(candidate); err == nil {
					dir = candidate
				} else {
					fmt.Fprintf(os.Stderr, "Instance not found: %s\n", arg)
					os.Exit(1)
				}
			}
		}
		profileName = parseProfile(os.Args[2:])
	} else {
		// No instance argument — use default n010w4. No fallback.
		dir = filepath.Join(defaultBasePath, defaultInstance)
		if _, err := os.Stat(dir); err != nil {
			// Also try datasets_json path.
			dir = filepath.Join("../../examples/inrc2/datasets_json", defaultInstance)
			if _, err := os.Stat(dir); err != nil {
				fmt.Fprintln(os.Stderr, "Default INRC-II benchmark instance not found. Please ensure examples/inrc2/testdatasets_json/n012w8 exists.")
				os.Exit(1)
			}
		}
		if len(os.Args) >= 3 {
			profileName = parseProfile(os.Args[2:])
		}
	}

	algProfile, ok := optimisation.GetProfile(profileName)
	if !ok {
		algProfile = optimisation.ResearchProfile()
		profileName = "research"
	}

	// Apply explicit CLI overrides.
	algProfile = applyProfileOverrides(os.Args[1:], algProfile)

	// --time is not supported. Reject clearly.
	timeBudget := parseTimeBudget(os.Args[1:])
	if timeBudget > 0 {
		fmt.Fprintln(os.Stderr, "--time is not supported. Use explicit algorithm tuning flags such as --tabu-max-iterations or --sa-max-iterations.")
		os.Exit(1)
	}

	// Support --algorithm to filter to a single algorithm.
	benchAlgorithm := parseAlgorithm(os.Args[1:])

	// Find scenario file.
	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading directory: %v\n", err)
		os.Exit(1)
	}

	var scenarioFile string
	var weekFiles []string
	var histFiles []string

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

	if scenarioFile == "" {
		fmt.Fprintln(os.Stderr, "No scenario file found")
		os.Exit(1)
	}
	if len(histFiles) == 0 || len(weekFiles) == 0 {
		fmt.Fprintln(os.Stderr, "No history or week files found")
		os.Exit(1)
	}

	sort.Strings(weekFiles)
	sort.Strings(histFiles)

	sc, err := inrc2.LoadScenario(scenarioFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	algs := optimisation.Available()
	sort.Strings(algs)

	// Filter to single algorithm if --algorithm supplied in benchmark-inrc2.
	if parseStringFlag(os.Args[1:], "--algorithm") != "" {
		algs = []string{benchAlgorithm}
	}

	hist, err := inrc2.LoadHistory(histFiles[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	numWeeks := sc.NumberOfWeeks
	if numWeeks > len(weekFiles) {
		numWeeks = len(weekFiles)
	}

	// Header.
	fmt.Println("=================================================")
	fmt.Println("  Open Workforce Platform Benchmark")
	fmt.Printf("  Instance: %s\n", sc.ID)
	fmt.Printf("  Profile:  %s\n", profileName)
	fmt.Printf("  Weeks:    %d\n", numWeeks)
	fmt.Println("=================================================")
	fmt.Println()
	if len(algs) == 1 {
		displayEffectiveConfig(algs[0], algProfile)
	}

	// Accumulate results per algorithm across all weeks.
	type algResult struct {
		alg          string
		totalPenalty int
		totalHard    int
		totalSoft    int
		totalAssign  int
		totalMs      int64
		totalCands   int
	}

	results := make(map[string]*algResult)
	for _, alg := range algs {
		results[alg] = &algResult{alg: alg}
	}

	currentHist := hist
	for w := 0; w < numWeeks; w++ {
		wd, err := inrc2.LoadWeekData(weekFiles[w])
		if err != nil {
			fmt.Fprintf(os.Stderr, "  Error loading week %d: %v\n", w, err)
			continue
		}

		// Track the solution to use for history advancement when running a single algorithm.
		var weekSolForHistory inrc2.Solution
		hasSolForHistory := false

		for _, alg := range algs {
			fmt.Printf("\r  Solving week %d/%d %s...          ", w+1, numWeeks, alg)

			var sol inrc2.Solution
			var scoreResult inrc2.ScoreResult
			var durationMs int64
			var candidatesEval int

			if alg == "parallel-feasible-roster-search" {
				pfrsConfig := parsePFRSConfig(os.Args[1:])
				pfrsSol, pfrsStats, pfrsScore, err := inrc2.SolveWeekPFRS(sc, wd, currentHist, pfrsConfig)
				if err != nil {
					fmt.Fprintf(os.Stderr, "\n  Error %s week %d: %v\n", alg, w, err)
					continue
				}
				sol = pfrsSol
				scoreResult = pfrsScore
				durationMs = pfrsStats.DurationMs
				candidatesEval = pfrsStats.CandidatesEvaluated
			} else {
				owpSol, planResult, err := inrc2.SolveWeek(sc, wd, currentHist, alg, algProfile)
				if err != nil {
					fmt.Fprintf(os.Stderr, "\n  Error %s week %d: %v\n", alg, w, err)
					continue
				}
				sol = owpSol
				scoreResult = inrc2.Score(sc, wd, currentHist, sol)
				stats := planResult.Statistics()
				durationMs = stats.DurationMs
				candidatesEval = stats.CandidatesEvaluated
			}

			r := results[alg]
			r.totalPenalty += scoreResult.SoftPenalty
			r.totalHard += scoreResult.HardViolations
			r.totalSoft += len(scoreResult.SoftDetails)
			r.totalAssign += len(sol.Assignments)
			r.totalMs += durationMs
			r.totalCands += candidatesEval

			// Capture solution for single-algorithm history advancement.
			if len(algs) == 1 {
				weekSolForHistory = sol
				hasSolForHistory = true
			}
		}

		// Advance history for next week.
		// When running a single algorithm, use its solution for history advancement
		// so the algorithm gets consistent history across weeks.
		// When running multiple algorithms, use constructive for fair comparison.
		if len(algs) == 1 && hasSolForHistory {
			currentHist = inrc2.UpdateHistory(sc, currentHist, weekSolForHistory)
		} else {
			sol, _, _ := inrc2.SolveWeek(sc, wd, currentHist, "constructive", algProfile)
			currentHist = inrc2.UpdateHistory(sc, currentHist, sol)
		}
	}
	fmt.Print("\r                                              \r")

	// Separate valid (Hard == 0) from invalid results.
	var valid []*algResult
	var invalid []*algResult
	for _, alg := range algs {
		r := results[alg]
		if r.totalHard == 0 {
			valid = append(valid, r)
		} else {
			invalid = append(invalid, r)
		}
	}

	// Check --show-invalid flag.
	showInvalid := false
	for _, arg := range os.Args[1:] {
		if arg == "--show-invalid" {
			showInvalid = true
		}
	}

	// Sort valid: penalty asc, then soft asc, then runtime asc, then name asc.
	sort.Slice(valid, func(i, j int) bool {
		if valid[i].totalPenalty != valid[j].totalPenalty {
			return valid[i].totalPenalty < valid[j].totalPenalty
		}
		if valid[i].totalSoft != valid[j].totalSoft {
			return valid[i].totalSoft < valid[j].totalSoft
		}
		if valid[i].totalMs != valid[j].totalMs {
			return valid[i].totalMs < valid[j].totalMs
		}
		return valid[i].alg < valid[j].alg
	})

	// Sort invalid: hard asc (least invalid first), then penalty asc.
	sort.Slice(invalid, func(i, j int) bool {
		if invalid[i].totalHard != invalid[j].totalHard {
			return invalid[i].totalHard < invalid[j].totalHard
		}
		return invalid[i].totalPenalty < invalid[j].totalPenalty
	})

	// Print valid league table.
	if len(valid) > 0 {
		fmt.Printf("%-6s %-28s %10s %8s %12s %10s %12s\n",
			"Rank", "Algorithm", "Penalty", "Soft", "Assignments", "Runtime", "Candidates")
		fmt.Println(strings.Repeat("-", 92))

		for rank, r := range valid {
			fmt.Printf("%-6d %-28s %10d %8d %12d %8dms %12d\n",
				rank+1, r.alg, r.totalPenalty, r.totalSoft,
				r.totalAssign, r.totalMs, r.totalCands)
		}
	} else {
		fmt.Println("No valid solutions (Hard = 0) found.")
	}

	// Print invalid/rejected section only when --show-invalid is supplied.
	if len(invalid) > 0 && showInvalid {
		fmt.Println()
		fmt.Println("Rejected (Invalid Solutions):")
		fmt.Printf("       %-28s %10s %8s %8s %12s %10s %12s\n",
			"Algorithm", "Penalty", "Hard", "Soft", "Assignments", "Runtime", "Candidates")
		fmt.Println("       " + strings.Repeat("-", 92))

		for _, r := range invalid {
			fmt.Printf("       %-28s %10d %8d %8d %12d %8dms %12d\n",
				r.alg, r.totalPenalty, r.totalHard, r.totalSoft,
				r.totalAssign, r.totalMs, r.totalCands)
		}
	}

	// Summary.
	fmt.Println()
	fmt.Println("Summary:")

	if len(valid) > 0 {
		fmt.Printf("  Best algorithm:    %s (penalty: %d)\n", valid[0].alg, valid[0].totalPenalty)

		fastest := valid[0]
		for _, r := range valid {
			if r.totalMs < fastest.totalMs {
				fastest = r
			}
		}
		fmt.Printf("  Fastest valid:     %s (%dms)\n", fastest.alg, fastest.totalMs)

		totalPenalty := 0
		totalMs := int64(0)
		totalSoft := 0
		for _, r := range valid {
			totalPenalty += r.totalPenalty
			totalMs += r.totalMs
			totalSoft += r.totalSoft
		}
		n := len(valid)
		fmt.Printf("  Average penalty:   %d\n", totalPenalty/n)
		fmt.Printf("  Average runtime:   %dms\n", totalMs/int64(n))
		fmt.Printf("  Average soft:      %d\n", totalSoft/n)
	} else {
		fmt.Println("  No valid solution found.")
		if showInvalid && len(invalid) > 0 {
			fmt.Printf("  Least invalid:     %s (hard: %d, penalty: %d)\n",
				invalid[0].alg, invalid[0].totalHard, invalid[0].totalPenalty)
		}
	}

	fmt.Println()
	fmt.Println("Done.")
}
