// feature_extract.go — builds policy FeatureVectors from search integration contexts.
package optimisation

import (
	"time"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation/policy"
)

// FeatureExtractor builds FeatureVectors from context-specific inputs.
// Each integration style (worker, portfolio, search) has a dedicated method.
type FeatureExtractor struct{}

// NewFeatureExtractor creates a FeatureExtractor.
func NewFeatureExtractor() *FeatureExtractor {
	return &FeatureExtractor{}
}

// FromWorkerContext builds a FeatureVector from a WorkerContext.
func (fe *FeatureExtractor) FromWorkerContext(ctx WorkerContext, runID string, instance string) FeatureVector {
	budgetConsumed := 0.0
	if ctx.AllocatedIters > 0 {
		budgetConsumed = 0.0 // at spawn time, worker hasn't started
	}

	return FeatureVector{
		SchemaVersion:      policy.FeatureSchemaVersion,
		Timestamp:          time.Now(),
		RunID:              runID,
		Problem:            "nrp",
		Instance:           instance,
		Algorithm:          ctx.Algorithm,
		IterationBudget:    ctx.AllocatedIters,
		IterationsComplete: 0,
		BudgetConsumed:     budgetConsumed,
		Temperature:        0,
		CurrentObjective:   ctx.ParentObjective,
		BestObjective:      ctx.GlobalBest,
		ParentObjective:    ctx.ParentObjective,
		DistanceFromBest:   ctx.DistanceFromBest,
		GapToReference:     -1,
		PlateauLength:      0,
		ImprovementRate:    ctx.RecentImprovRate,
		AcceptanceRate:     0,
		Diversity:          0,
		Entropy:            ctx.Entropy,
		WorkerCount:        ctx.WorkerCount,
		BranchDepth:        ctx.Depth,
		Week:               ctx.Week,
		BeamHealth:         ctx.BeamHealth,
		ElapsedMs:          0,
		TimeRatio:          0,
		DecisionType:       "worker",
	}
}

// FromSearchProgress builds a FeatureVector from SearchProgress.
func (fe *FeatureExtractor) FromSearchProgress(
	p SearchProgress, runID string, problem string, instance string, elapsedMs int64,
) FeatureVector {
	budgetConsumed := 0.0
	if p.IterationsTotal > 0 {
		budgetConsumed = float64(p.IterationsComplete) / float64(p.IterationsTotal)
	}
	acceptanceRate := 0.0
	if p.CandidatesEval > 0 {
		acceptanceRate = float64(p.Accepted) / float64(p.CandidatesEval)
	}

	return FeatureVector{
		SchemaVersion:      policy.FeatureSchemaVersion,
		Timestamp:          time.Now(),
		RunID:              runID,
		Problem:            problem,
		Instance:           instance,
		Algorithm:          p.Algorithm,
		IterationBudget:    p.IterationsTotal,
		IterationsComplete: p.IterationsComplete,
		BudgetConsumed:     budgetConsumed,
		Temperature:        p.Temperature,
		CurrentObjective:   p.CurrentPenalty,
		BestObjective:      p.BestPenalty,
		ParentObjective:    p.InitialPenalty,
		DistanceFromBest:   p.CurrentPenalty - p.BestPenalty,
		GapToReference:     -1,
		PlateauLength:      p.PlateauLength,
		ImprovementRate:    p.ImprovementRate,
		AcceptanceRate:     acceptanceRate,
		Diversity:          0,
		Entropy:            0,
		WorkerCount:        0,
		BranchDepth:        0,
		Week:               0,
		BeamHealth:         0,
		ElapsedMs:          elapsedMs,
		TimeRatio:          0,
		DecisionType:       "search",
	}
}

// FromPortfolioContext builds a FeatureVector for a specific strategy within a portfolio.
func (fe *FeatureExtractor) FromPortfolioContext(
	ctx PortfolioContext, strategy string, runID string,
) FeatureVector {
	wins := 0
	total := 0
	for _, entry := range ctx.PreviousResults {
		if entry.Strategy == strategy {
			total++
			if entry.Improved {
				wins++
			}
		}
	}
	winRate := 0.0
	if total > 0 {
		winRate = float64(wins) / float64(total)
	}

	return FeatureVector{
		SchemaVersion:      policy.FeatureSchemaVersion,
		Timestamp:          time.Now(),
		RunID:              runID,
		Problem:            ctx.ProblemType,
		Instance:           ctx.Instance,
		Algorithm:          strategy,
		IterationBudget:    ctx.TotalBudget,
		IterationsComplete: 0,
		BudgetConsumed:     0,
		Temperature:        0,
		CurrentObjective:   0,
		BestObjective:      0,
		ParentObjective:    0,
		DistanceFromBest:   0,
		GapToReference:     winRate,
		PlateauLength:      0,
		ImprovementRate:    0,
		AcceptanceRate:     0,
		Diversity:          0,
		Entropy:            0,
		WorkerCount:        len(ctx.Strategies),
		BranchDepth:        0,
		Week:               0,
		BeamHealth:         0,
		ElapsedMs:          0,
		TimeRatio:          0,
		DecisionType:       "portfolio",
	}
}
