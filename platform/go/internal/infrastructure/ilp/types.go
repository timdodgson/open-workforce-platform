package ilp

import "time"

// BenchmarkConfig holds parameters for an ILP benchmark run.
type BenchmarkConfig struct {
	Instance     string        // Instance directory path
	Weeks        int           // Number of weeks to solve (1-8)
	TimeLimit    time.Duration // Maximum solver runtime
	SolverName   string        // "highs"
	OutputPath   string        // Path to write benchmark JSON
	ProgressPath string        // Path to write solve progress CSV
	Parallel     bool          // Enable parallel tree search
}

// BenchmarkResult holds the outcome of an ILP benchmark solve.
type BenchmarkResult struct {
	Instance               string   `json:"instance"`
	Weeks                  int      `json:"weeks"`
	Solver                 string   `json:"solver"`
	Status                 string   `json:"status"` // OPTIMAL, FEASIBLE, INFEASIBLE, TIMEOUT, ERROR
	Objective              int      `json:"objective"`
	LowerBound             int      `json:"lowerBound"`
	GapPercent             float64  `json:"gapPercent"`
	RuntimeSeconds         float64  `json:"runtimeSeconds"`
	TimeLimit              int      `json:"timeLimit"`
	SolutionPath           string   `json:"solutionPath,omitempty"`
	ProgressPath           string   `json:"progressPath,omitempty"`
	RosterPath             string   `json:"rosterPath,omitempty"`
	Notes                  string   `json:"notes,omitempty"`
	HardViolations         int      `json:"hardViolations"`
	ModelCompleteness      string   `json:"modelCompleteness"`
	SupportedConstraints   []string `json:"supportedConstraints"`
	UnsupportedConstraints []string `json:"unsupportedConstraints"`
	Threads                int      `json:"threads"`
	Parallel               bool     `json:"parallel"`
}

// ComparisonResult compares a PFRS run against an ILP benchmark.
type ComparisonResult struct {
	Instance     string  `json:"instance"`
	Weeks        int     `json:"weeks"`
	ILPObjective int     `json:"ilpObjective"`
	ILPStatus    string  `json:"ilpStatus"`
	PFRSPenalty  int     `json:"pfrsPenalty"`
	AbsoluteGap  int     `json:"absoluteGap"`
	GapPercent   float64 `json:"gapPercent"`
	ILPRuntime   float64 `json:"ilpRuntimeSeconds"`
	PFRSRuntime  float64 `json:"pfrsRuntimeSeconds"`
}

// Solver is the interface for exact benchmark solvers.
// Implementations can use HiGHS, CBC, OR-Tools, etc.
type Solver interface {
	// Name returns the solver identifier.
	Name() string

	// Available checks if the solver binary is accessible.
	Available() bool

	// Solve runs the ILP model and returns the result.
	Solve(modelPath string, timeLimit time.Duration) (SolverOutput, error)
}

// SolverOutput holds raw solver output before mapping to BenchmarkResult.
type SolverOutput struct {
	Status          string
	Objective       float64
	LowerBound      float64
	RuntimeSeconds  float64
	SolutionValues  map[string]float64 // variable name → value
	ProgressCSVPath string             // path to solve progress CSV (if captured)
}
