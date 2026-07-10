package main

import (
	"fmt"
	"os"
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
	case "tune-pfrs":
		runTunePFRS()
	case "visualise-pfrs":
		runVisualisePFRS()
	case "benchmark-ilp":
		runBenchmarkILP()
	case "solve-cvrp":
		runSolveCVRP()
	case "benchmark-cvrp-ilp":
		runBenchmarkCVRPILP()
	case "benchmark-vrptw-ilp":
		runBenchmarkVRPTWILP()
	case "benchmark-jss-ilp":
		runBenchmarkJSSILP()
	case "solve-jobshop":
		runSolveJobShop()
	case "solve-vrptw":
		runSolveVRPTW()
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  owp optimise <dataset-path> ...   (deprecated — use solve-* / tune-pfrs)")
	fmt.Fprintln(os.Stderr, "  owp benchmark <datasets-directory> (deprecated — use benchmark-inrc2)")
	fmt.Fprintln(os.Stderr, "  owp convert-nrp <nrp-input> <output-dataset>")
	fmt.Fprintln(os.Stderr, "  owp validate-inrc2 <scenario-file> <week-file> <history-file> <solution-file>")
	fmt.Fprintln(os.Stderr, "  owp solve-inrc2 <scenario-file> <week-file> <history-file> <solution-output-file> [--algorithm tabu-search] [--profile default]")
	fmt.Fprintln(os.Stderr, "  owp benchmark-inrc2 [instance-name] [--profile research]")
	fmt.Fprintln(os.Stderr, "  owp tune-pfrs [--instance <name>] [--pfrs-run-label <name>] [--pfrs-mode sa|portfolio] [--pfrs-storage s3]")
	fmt.Fprintln(os.Stderr, "  owp visualise-pfrs --audit-csv <path> --output-dir <path>")
	fmt.Fprintln(os.Stderr, "  owp benchmark-ilp --instance <name> [--weeks <n>] [--time-limit <seconds>] [--parallel true] [--storage s3] [--output <path>] [--compare-pfrs <penalty>]")
	fmt.Fprintln(os.Stderr, "  owp solve-cvrp --instance <path> [--mode sa|lahc|tabu|portfolio] [--iterations <n>] [--temperature <t>] [--seed <s>] [--run-label <name>] [--worker-decision-mode off|shadow|assist|adaptive]")
	fmt.Fprintln(os.Stderr, "  owp benchmark-cvrp-ilp --instance <path.vrp> [--time-limit <seconds>] [--parallel true] [--run-label <name>]")
	fmt.Fprintln(os.Stderr, "  owp benchmark-vrptw-ilp --instance <path.txt> [--time-limit <seconds>] [--parallel true] [--run-label <name>] [--storage s3]")
	fmt.Fprintln(os.Stderr, "  owp benchmark-jss-ilp --instance <path.txt> [--time-limit <seconds>] [--parallel true] [--run-label <name>] [--storage s3]")
	fmt.Fprintln(os.Stderr, "  owp solve-jobshop --instance <path> [--mode sa|lahc|adaptive|portfolio] [--iterations <n>] [--seed <s>] [--run-label <name>] [--worker-decision-mode off|shadow|assist|adaptive]")
	fmt.Fprintln(os.Stderr, "  owp solve-vrptw --instance <path> [--mode sa|lahc|tabu|portfolio] [--iterations <n>] [--seed <s>] [--run-label <name>] [--worker-decision-mode off|shadow|assist|adaptive]")
}
