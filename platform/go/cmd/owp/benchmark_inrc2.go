package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2/legacysearch"
)

const (
	inrc2TestDatasets  = "../../examples/inrc2/testdatasets_json"
	inrc2CompDatasets  = "../../examples/inrc2/datasets_json"
	defaultINRC2Bench  = "n012w8"
	defaultINRC2Prof   = "research"
)

func runBenchmarkINRC2() {
	dir, profileName := resolveBenchmarkINRC2Dir(os.Args)

	algProfile, ok := legacysearch.GetProfile(profileName)
	if !ok {
		algProfile = legacysearch.ResearchProfile()
		profileName = "research"
	}
	algProfile = applyProfileOverrides(os.Args[1:], algProfile)

	if parseTimeBudget(os.Args[1:]) > 0 {
		fmt.Fprintln(os.Stderr, "--time is not supported. Use explicit algorithm tuning flags such as --tabu-max-iterations or --sa-max-iterations.")
		os.Exit(1)
	}

	algorithmFilter := parseStringFlag(os.Args[1:], "--algorithm")

	scenarioFile, weekFiles, histFiles, err := inrc2.ScanInstanceDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if scenarioFile == "" {
		fmt.Fprintln(os.Stderr, "No scenario file found")
		os.Exit(1)
	}
	if len(histFiles) == 0 || len(weekFiles) == 0 {
		fmt.Fprintln(os.Stderr, "No history or week files found")
		os.Exit(1)
	}

	sc, err := inrc2.LoadScenario(scenarioFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	algs := legacysearch.Available()
	sort.Strings(algs)
	if algorithmFilter != "" {
		algs = []string{strings.TrimSpace(algorithmFilter)}
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

	pfrsConfig := parsePFRSConfig(os.Args[1:])
	results := inrc2.RunAlgorithmBenchmark(inrc2.AlgorithmBenchmarkParams{
		Scenario:   sc,
		WeekFiles:  weekFiles,
		History:    hist,
		NumWeeks:   numWeeks,
		Algorithms: algs,
		AlgProfile: algProfile,
		PFRSConfig: pfrsConfig,
		OnWeekStart: func(w int, alg string) {
			fmt.Printf("\r  Solving week %d/%d %s...          ", w+1, numWeeks, alg)
		},
	})
	fmt.Print("\r                                              \r")

	showInvalid := parseShowInvalidFlag(os.Args[1:])
	valid, invalid := inrc2.RankAlgorithmBenchmarkResults(algs, results)
	printINRC2BenchmarkResults(valid, invalid, showInvalid)
	printINRC2BenchmarkSummary(valid, invalid, showInvalid)
}

func resolveBenchmarkINRC2Dir(args []string) (dir, profileName string) {
	profileName = defaultINRC2Prof

	if len(args) >= 3 && !strings.HasPrefix(args[2], "--") {
		arg := args[2]
		if _, err := os.Stat(arg); err == nil {
			dir = arg
		} else if candidate := filepath.Join(inrc2TestDatasets, arg); statOK(candidate) {
			dir = candidate
		} else if candidate := filepath.Join(inrc2CompDatasets, arg); statOK(candidate) {
			dir = candidate
		} else {
			fmt.Fprintf(os.Stderr, "Instance not found: %s\n", arg)
			os.Exit(1)
		}
		profileName = parseProfile(args[2:])
		return dir, profileName
	}

	dir = filepath.Join(inrc2TestDatasets, defaultINRC2Bench)
	if !statOK(dir) {
		dir = filepath.Join(inrc2CompDatasets, defaultINRC2Bench)
		if !statOK(dir) {
			fmt.Fprintln(os.Stderr, "Default INRC-II benchmark instance not found. Please ensure examples/inrc2/testdatasets_json/n012w8 exists.")
			os.Exit(1)
		}
	}
	if len(args) >= 3 {
		profileName = parseProfile(args[2:])
	}
	return dir, profileName
}

func statOK(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
