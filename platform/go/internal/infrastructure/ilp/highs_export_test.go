package ilp

import "time"

// ParseNativeOutputForTest exposes parseNativeOutput for testing.
func ParseNativeOutputForTest(output string, elapsed time.Duration) SolverOutput {
	return parseNativeOutput(output, elapsed)
}

// ProgressPointExported is an exported version of progressPoint for testing.
type ProgressPointExported struct {
	Elapsed   float64
	Incumbent float64
	Bound     float64
	Gap       float64
	Nodes     int
	LpIters   int
}

// ParseProgressLineForTest exposes parseProgressLine for testing.
func ParseProgressLineForTest(line string) (ProgressPointExported, bool) {
	pt, ok := parseProgressLine(line)
	if !ok {
		return ProgressPointExported{}, false
	}
	return ProgressPointExported{
		Elapsed:   pt.elapsed,
		Incumbent: pt.incumbent,
		Bound:     pt.bound,
		Gap:       pt.gap,
		Nodes:     pt.nodes,
		LpIters:   pt.lpIters,
	}, true
}
