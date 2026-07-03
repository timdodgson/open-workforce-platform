package ilp

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// HighsSolver uses the native HiGHS binary to solve LP/MIP models.
// Requires: highs binary on PATH (Apache static build for parallel on Windows).
// Download from: https://github.com/ERGO-Code/HiGHS/releases
type HighsSolver struct {
	// Parallel enables parallel tree search. When true, HiGHS decides thread count.
	Parallel bool
}

// Name returns the solver identifier.
func (h *HighsSolver) Name() string {
	return "HiGHS"
}

// Available checks if the native highs binary is on PATH.
func (h *HighsSolver) Available() bool {
	_, err := exec.LookPath("highs")
	return err == nil
}

// Solve runs the HiGHS solver on the given LP model file.
func (h *HighsSolver) Solve(modelPath string, timeLimit time.Duration) (SolverOutput, error) {
	tmpDir, err := os.MkdirTemp("", "ilp-solver-*")
	if err != nil {
		return SolverOutput{}, fmt.Errorf("failed to create temp dir: %w", err)
	}

	solutionPath := filepath.Join(tmpDir, "solution.sol")
	progressCSVPath := filepath.Join(tmpDir, "progress.csv")

	// Build command arguments.
	args := []string{
		modelPath,
		"--time_limit", fmt.Sprintf("%.1f", timeLimit.Seconds()),
		"--solution_file", solutionPath,
	}

	if h.Parallel {
		args = append(args, "--parallel", "on")
	}

	cmd := exec.Command("highs", args...)

	// Capture stdout via a pipe for progress parsing.
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		os.RemoveAll(tmpDir)
		return SolverOutput{}, fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	cmd.Stderr = os.Stderr

	start := time.Now()
	if err := cmd.Start(); err != nil {
		os.RemoveAll(tmpDir)
		return SolverOutput{}, fmt.Errorf("failed to start highs: %w", err)
	}

	// Parse progress from stdout in real-time.
	var progressPoints []progressPoint
	var allOutput strings.Builder
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		allOutput.WriteString(line)
		allOutput.WriteString("\n")
		// Forward to stderr so user sees progress.
		fmt.Fprintln(os.Stderr, line)

		// Try to parse as MIP progress line.
		if pt, ok := parseProgressLine(line); ok {
			progressPoints = append(progressPoints, pt)
		}
	}

	err = cmd.Wait()
	elapsed := time.Since(start)

	// Write progress CSV.
	if len(progressPoints) > 0 {
		writeProgressCSV(progressCSVPath, progressPoints)
	}

	// Parse the final result from output.
	output := allOutput.String()
	result := parseNativeOutput(output, elapsed)
	result.ProgressCSVPath = progressCSVPath

	// Try to parse solution file for variable values.
	if result.Status == "OPTIMAL" || result.Status == "FEASIBLE" {
		if vars, err := parseNativeSolution(solutionPath); err == nil {
			result.SolutionValues = vars
		}
	}

	if err != nil && result.Status == "" {
		result.Status = "ERROR"
		os.RemoveAll(tmpDir)
		return result, fmt.Errorf("highs exited with error: %w", err)
	}

	return result, nil
}

// progressPoint holds a single progress data point from the MIP solve.
type progressPoint struct {
	elapsed    float64
	incumbent  float64
	bound      float64
	gap        float64
	nodes      int
	lpIters    int
	hasIncumb  bool
}

// MIP progress line pattern from HiGHS output.
// Example: " B     1234    567    890  12.3%   1234.5          2345.6       10.1%    123   45   67    12345    12.3s"
var progressRe = regexp.MustCompile(
	`^\s*\S+\s+` + // Src
		`(\d+)\s+` + // Proc (nodes)
		`\d+\s+` + // InQueue
		`\d+\s+` + // Leaves
		`[\d.]+%\s+` + // Expl%
		`([-\d.e+inf]+)\s+` + // BestBound
		`([-\d.e+inf]+)\s+` + // BestSol
		`([\d.]+)%\s+` + // Gap%
		`\d+\s+` + // Cuts
		`\d+\s+` + // InLp
		`\d+\s+` + // Confl
		`(\d+)\s+` + // LpIters
		`([\d.]+)s`) // Time

func parseProgressLine(line string) (progressPoint, bool) {
	m := progressRe.FindStringSubmatch(line)
	if m == nil {
		return progressPoint{}, false
	}

	nodes, _ := strconv.Atoi(m[1])
	bound := parseFloat(m[2])
	incumbent := parseFloat(m[3])
	gap, _ := strconv.ParseFloat(m[4], 64)
	lpIters, _ := strconv.Atoi(m[5])
	elapsed, _ := strconv.ParseFloat(m[6], 64)

	pt := progressPoint{
		elapsed:   elapsed,
		bound:     bound,
		incumbent: incumbent,
		gap:       gap,
		nodes:     nodes,
		lpIters:   lpIters,
		hasIncumb: m[3] != "inf" && m[3] != "-inf",
	}
	return pt, true
}

func parseFloat(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "inf" || s == "+inf" {
		return 0
	}
	if s == "-inf" {
		return 0
	}
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

func writeProgressCSV(path string, points []progressPoint) {
	f, err := os.Create(path)
	if err != nil {
		return
	}
	defer f.Close()

	f.WriteString("elapsed,incumbent,bound,gap,nodes,lpIters\n")
	for _, p := range points {
		incStr := ""
		if p.hasIncumb {
			incStr = fmt.Sprintf("%.1f", p.incumbent)
		}
		boundStr := ""
		if p.bound > 0 {
			boundStr = fmt.Sprintf("%.1f", p.bound)
		}
		fmt.Fprintf(f, "%.1f,%s,%s,%.2f,%d,%d\n",
			p.elapsed, incStr, boundStr, p.gap, p.nodes, p.lpIters)
	}
}

// parseNativeOutput extracts status, objective, and bound from HiGHS console output.
func parseNativeOutput(output string, elapsed time.Duration) SolverOutput {
	result := SolverOutput{
		RuntimeSeconds: elapsed.Seconds(),
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.Contains(line, "Solving report") ||
			strings.Contains(line, "Status") {
			if strings.Contains(line, "Optimal") {
				result.Status = "OPTIMAL"
			} else if strings.Contains(line, "Time limit reached") ||
				strings.Contains(line, "TIME_LIMIT") {
				result.Status = "FEASIBLE"
			} else if strings.Contains(line, "Infeasible") {
				result.Status = "INFEASIBLE"
			}
		}

		if strings.HasPrefix(line, "Objective value") || strings.HasPrefix(line, "Primal bound") {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				if v, err := strconv.ParseFloat(parts[len(parts)-1], 64); err == nil {
					result.Objective = v
				}
			}
		}

		if strings.HasPrefix(line, "Dual bound") || strings.Contains(line, "Best bound") {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				if v, err := strconv.ParseFloat(parts[len(parts)-1], 64); err == nil {
					result.LowerBound = v
				}
			}
		}

		// HiGHS prints "Optimal - objective = 1234.0" or similar
		if strings.Contains(line, "objective") && strings.Contains(line, "=") {
			parts := strings.Split(line, "=")
			if len(parts) == 2 {
				if v, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64); err == nil {
					result.Objective = v
				}
			}
		}
	}

	// If we found an objective but no status, infer from exit.
	if result.Status == "" && result.Objective > 0 {
		result.Status = "FEASIBLE"
	}

	return result
}

// parseNativeSolution reads a HiGHS .sol file and extracts variable values.
func parseNativeSolution(path string) (map[string]float64, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	vars := make(map[string]float64)
	lines := strings.Split(string(data), "\n")
	inColumns := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "Columns" {
			inColumns = true
			continue
		}
		if line == "Rows" || line == "" {
			if inColumns {
				break
			}
			continue
		}
		if inColumns {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				name := parts[0]
				val, err := strconv.ParseFloat(parts[1], 64)
				if err == nil && val > 0.5 {
					vars[name] = val
				}
			}
		}
	}

	return vars, nil
}

