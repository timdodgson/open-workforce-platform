// Package optimisation provides the generic Search Intelligence abstraction.
//
// Search Intelligence is a universal concept that allows AI to advise any solver
// architecture without changing its core behaviour. The abstraction supports three
// integration styles, each suited to a different solver architecture:
//
//   - WorkerAssist: for beam search / parallel worker-spawning solvers (NRP PFRS)
//   - PortfolioAssist: for multi-strategy portfolio solvers (CVRP, JSS, VRPTW)
//   - SearchAssist: for single-search solvers (any algorithm running alone)
//
// CLI flag --worker-decision-mode controls all three styles (historical name).
// It sets SearchConfig.AssistMode for CVRP/JSS/VRPTW and wires inrc2 worker
// engines for tune-pfrs. Modes: off, shadow, assist, adaptive.
//
// Each integration style defines its own input/recommendation types, but all share
// the same lifecycle: evaluate at a decision point, recommend an action, record the
// outcome, and learn from the result.
//
// All three styles support four modes: off, shadow, assist, adaptive.
//
// Current status:
//   - WorkerAssist interface: production via inrc2.WorkerDecisionEngine (PFRS; parallel type, not WorkerAssist)
//   - SearchAssist: IMPLEMENTED (SA/LAHC/Tabu on CVRP, JSS, VRPTW via SearchHookRunner)
//   - PortfolioAssist: IMPLEMENTED (portfolio mode; rule-based + learned model fallback)
//   - Adaptive mode: IMPLEMENTED (AdaptiveSearchAssist for search; assist+recorder for PFRS workers)
//   - SI 2.0 policies (RulePolicy, LearnedPolicy, HybridPolicy): scaffolded; PolicySearchHookRunner not yet wired into search.go
package optimisation

// --- Core Types ---

// Confidence represents the model's certainty in a recommendation (0.0 to 1.0).
type Confidence float64

// SafetyStatus indicates whether a recommendation passed safety checks.
type SafetyStatus string

const (
	SafetyPassed   SafetyStatus = "passed"
	SafetyRejected SafetyStatus = "rejected"
)

// --- WorkerAssist ---
//
// Integration style for beam search / parallel worker-spawning solvers.
// The AI evaluates each worker at spawn time and recommends whether to run it.
//
// Used by: NRP PFRS (implemented and validated)
//
// Decision points: each time a new worker is about to be submitted to the work queue.
//
// Integration:
//   submitWork() calls WorkerDecisionEngine.Evaluate (inrc2; parallel to WorkerAssist interface)
//   If recommendation is Skip and safety allows → worker is not submitted
//   If recommendation is ReduceBudget → advisory (future: reduce iterations)
//   If recommendation is IncreaseBudget → advisory (future: increase iterations)
//   If recommendation is ChangeAlgo → worker runs with different algorithm
//   All decisions are logged to worker_assist.csv

// WorkerAction is the set of actions the AI can recommend for a worker.
type WorkerAction string

const (
	WorkerRun            WorkerAction = "run"
	WorkerSkip           WorkerAction = "skip"
	WorkerReduceBudget   WorkerAction = "reduce_budget"
	WorkerIncreaseBudget WorkerAction = "increase_budget"
	WorkerChangeAlgo     WorkerAction = "change_algorithm"
)

// WorkerContext captures spawn-time state for a worker decision.
type WorkerContext struct {
	Algorithm                  string
	Week                       int
	Depth                      int
	ParentObjective            int
	GlobalBest                 int
	DistanceFromBest           int
	BeamRank                   int
	Entropy                    float64
	BeamHealth                 float64
	RecentImprovRate           float64
	AllocatedIters             int
	WorkerCount                int
	IsGlobalBestLineage        bool
	ParentProducedGlobalBest   bool
	GenerationsSinceGlobalBest int
}

// WorkerRecommendation is the AI's advice for a single worker spawn.
type WorkerRecommendation struct {
	Action             WorkerAction
	Confidence         Confidence
	Reasons            []string
	SuggestedAlgorithm string // only if ChangeAlgo
	SuggestedBudget    int    // only if ReduceBudget/IncreaseBudget
}

// WorkerAssist is the interface for worker-level AI advice.
type WorkerAssist interface {
	// Evaluate returns a recommendation for the given worker context.
	Evaluate(ctx WorkerContext) WorkerRecommendation
}

// --- PortfolioAssist ---
//
// Integration style for multi-strategy portfolio solvers.
// The AI evaluates portfolio allocation between strategies and recommends adjustments.
//
// Used by: CVRP, JSS, VRPTW (portfolio mode)
//
// Decision points:
//   - Before portfolio run: allocate iteration budgets across strategies
//   - During portfolio run (if streaming): terminate/extend individual strategies
//   - After portfolio run: recommend next-run adjustments
//
// Integration:
//   Before: PortfolioAssist.AllocateBudget(strategies, totalBudget, history)
//   During: PortfolioAssist.EvaluateProgress(strategyProgress) — optional
//   After: PortfolioAssist.ReviewOutcome(results) — for learning

// StrategyBudget represents an iteration budget allocation for one strategy.
type StrategyBudget struct {
	Strategy   string
	Iterations int
	Priority   int // higher = more promising
}

// PortfolioContext captures the state at portfolio decision time.
type PortfolioContext struct {
	Strategies      []string
	TotalBudget     int
	ProblemType     string
	Instance        string
	PreviousResults []PortfolioHistoryEntry // past runs on same instance
}

// PortfolioHistoryEntry records one past portfolio run result.
type PortfolioHistoryEntry struct {
	Strategy    string
	BestPenalty int
	Iterations  int
	RuntimeMs   int64
	Improved    bool
}

// PortfolioAction is the set of actions the AI can recommend for a portfolio.
type PortfolioAction string

const (
	PortfolioAllocate     PortfolioAction = "allocate"      // set iteration budgets
	PortfolioSkipStrategy PortfolioAction = "skip_strategy" // don't run one strategy
	PortfolioTerminate    PortfolioAction = "terminate"     // stop a running strategy early
	PortfolioExtend       PortfolioAction = "extend"        // give more iterations
	PortfolioRestart      PortfolioAction = "restart"       // restart with different seed
	PortfolioAdjustParams PortfolioAction = "adjust_params" // change temperature/LAHC/tabu
)

// PortfolioRecommendation is the AI's advice for portfolio management.
type PortfolioRecommendation struct {
	Action     PortfolioAction
	Confidence Confidence
	Reasons    []string
	Budgets    []StrategyBudget   // for Allocate action
	Target     string             // strategy name for Skip/Terminate/Extend
	Adjustment map[string]float64 // parameter adjustments for AdjustParams
}

// PortfolioAssist is the interface for portfolio-level AI advice.
type PortfolioAssist interface {
	// AllocateBudget recommends how to distribute iterations across strategies.
	AllocateBudget(ctx PortfolioContext) PortfolioRecommendation

	// EvaluateProgress checks mid-run progress and may recommend early termination.
	// Returns nil if no action is recommended.
	EvaluateProgress(strategy string, iterationsUsed int, currentBest int, globalBest int) *PortfolioRecommendation

	// ReviewOutcome records the final result for learning.
	ReviewOutcome(strategy string, result PortfolioHistoryEntry)
}

// --- SearchAssist ---
//
// Integration style for single-search algorithms.
// The AI monitors search progress and recommends parameter adjustments or early stop.
//
// Used by: any single-algorithm run (SA, LAHC, Tabu)
//
// Decision points:
//   - Periodically during search (every N iterations)
//   - When stagnation is detected
//   - When improvement rate changes significantly
//
// Integration:
//   Search loop calls SearchAssist.Checkpoint(progress) periodically
//   If recommendation is EarlyStop → search terminates
//   If recommendation is Restart → search restarts from current best
//   If recommendation is AdjustTemp → temperature is modified
//   If recommendation is AdjustBudget → remaining iterations are changed

// SearchAction is the set of actions the AI can recommend during a search.
type SearchAction string

const (
	SearchContinue     SearchAction = "continue"      // no change
	SearchEarlyStop    SearchAction = "early_stop"    // terminate search
	SearchRestart      SearchAction = "restart"       // restart from best known
	SearchAdjustTemp   SearchAction = "adjust_temp"   // change temperature
	SearchAdjustLAHC   SearchAction = "adjust_lahc"   // change LAHC buffer length
	SearchAdjustTabu   SearchAction = "adjust_tabu"   // change tabu tenure
	SearchAdjustBudget SearchAction = "adjust_budget" // change remaining iterations
)

// SearchProgress captures the current state of a running search.
type SearchProgress struct {
	Algorithm          string
	IterationsComplete int
	IterationsTotal    int
	CurrentPenalty     int
	BestPenalty        int
	InitialPenalty     int
	ImprovementRate    float64 // improvements per 10K iterations (recent window)
	Temperature        float64 // current SA temperature (0 if not SA)
	PlateauLength      int     // iterations since last improvement
	Accepted           int
	Rejected           int
	CandidatesEval     int
}

// SearchRecommendation is the AI's advice during a running search.
type SearchRecommendation struct {
	Action         SearchAction
	Confidence     Confidence
	Reasons        []string
	NewTemperature float64 // for AdjustTemp
	NewLAHCLength  int     // for AdjustLAHC
	NewTabuTenure  int     // for AdjustTabu
	NewBudget      int     // for AdjustBudget (total iterations)
}

// SearchAssist is the interface for search-level AI advice.
type SearchAssist interface {
	// Checkpoint evaluates progress and may recommend an action.
	// Called periodically during search (e.g. every 10K iterations).
	// Returns nil if no action is recommended (equivalent to "continue").
	Checkpoint(progress SearchProgress) *SearchRecommendation
}
