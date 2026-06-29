package optimisation

import (
	"time"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/plan"
)

type tabuSearchAlgorithm struct{}

func init() {
	register(&tabuSearchAlgorithm{})
}

func (ts *tabuSearchAlgorithm) Name() string {
	return "tabu-search"
}

// tabuEntry records a move to prevent immediate reversal.
type tabuEntry struct {
	WorkItemID   string
	FromResource string
	ToResource   string
}

func (ts *tabuSearchAlgorithm) Solve(ctx OptimisationContext) (plan.OptimisedPlan, error) {
	startTime := time.Now()
	items := ctx.Items()
	capacities := ctx.Resources()
	priorities := ctx.WorkItems()

	if err := validate(items, capacities); err != nil {
		return plan.OptimisedPlan{}, err
	}

	sorted := orderByPriority(items, priorities)
	assignments, unassigned, _ := assignItems(sorted, capacities, priorities, ctx)

	// Warm start.
	if existing := ctx.ExistingAssignments(); len(existing) > 0 {
		if scheduleFeasible(existing, capacities, priorities, ctx) {
			assignedSet := make(map[string]bool, len(existing))
			for _, a := range existing {
				assignedSet[a.WorkItemID()] = true
			}
			assignments = existing
			unassigned = nil
			for _, item := range items {
				if !assignedSet[item.ID()] {
					unassigned = append(unassigned, item.ID())
				}
			}
		}
	}

	requiredSkillOf := make(map[string]string, len(priorities))
	durationOf := make(map[string]int, len(priorities))
	for _, p := range priorities {
		requiredSkillOf[p.WorkItemID] = p.RequiredSkill
		dur := p.Duration
		if dur <= 0 {
			dur = 1
		}
		durationOf[p.WorkItemID] = dur
	}

	resourceIndex := make(map[string]int, len(capacities))
	for i, rc := range capacities {
		resourceIndex[rc.ResourceID] = i
	}

	totalItems := len(items)
	currentScore := ObjectiveScore(assignments, ctx)
	bestAssignments := copyAssignments(assignments)
	bestUnassigned := copyStrings(unassigned)
	bestScore := currentScore

	var tabuList []tabuEntry
	candidatesEvaluated := 0
	improvementsAccepted := 0
	iterationsRun := 0
	maxIter := ctx.Profile().TabuMaxIterations
	listSize := ctx.Profile().TabuListSize
	aspirationEnabled := ctx.Profile().TabuAspirationEnabled

	for iteration := 0; iteration < maxIter; iteration++ {
		iterationsRun++

		// Collect all candidate moves.
		var allMoves []CandidateMove

		// Placement moves for unassigned items.
		for _, uid := range unassigned {
			skill := requiredSkillOf[uid]
			moves := GenerateMoves(uid, skill, assignments, capacities, resourceIndex, requiredSkillOf, durationOf)
			allMoves = append(allMoves, moves...)
		}

		// Neighbourhood moves (swaps, relocates, reorders).
		allMoves = append(allMoves, GenerateAllNeighbourhoodMoves(assignments, capacities, resourceIndex, requiredSkillOf, durationOf)...)

		if len(allMoves) == 0 {
			break
		}

		// Find best admissible move.
		bestMoveScore := -1 << 62 // very negative
		bestMoveIdx := -1

		for i, m := range allMoves {
			candidatesEvaluated++

			if isTabu(tabuList, m) {
				// Aspiration: allow tabu move if it beats best-ever score.
				if !aspirationEnabled {
					continue
				}
				trial, ok := ApplyMove(m, copyAssignments(assignments))
				if !ok || !scheduleFeasible(trial, capacities, priorities, ctx) {
					continue
				}
				score := ObjectiveScore(trial, ctx)
				if score <= bestScore {
					continue // Tabu and doesn't beat best — skip.
				}
				// Aspiration accepted.
				if score > bestMoveScore {
					bestMoveScore = score
					bestMoveIdx = i
				}
				continue
			}

			trial, ok := ApplyMove(m, copyAssignments(assignments))
			if !ok || !scheduleFeasible(trial, capacities, priorities, ctx) {
				continue
			}

			score := ObjectiveScore(trial, ctx)
			if score > bestMoveScore {
				bestMoveScore = score
				bestMoveIdx = i
			}
		}

		if bestMoveIdx < 0 {
			break // no admissible moves
		}

		// Apply the best move.
		chosen := allMoves[bestMoveIdx]
		newAssignments, ok := ApplyMove(chosen, assignments)
		if !ok {
			break
		}

		// Update unassigned list if this was a placement.
		if chosen.Type == Placement || chosen.Type == Displacement {
			for i, uid := range unassigned {
				if uid == chosen.WorkItemID {
					unassigned = append(unassigned[:i], unassigned[i+1:]...)
					break
				}
			}
		}

		assignments = newAssignments
		currentScore = bestMoveScore
		improvementsAccepted++

		// Track best-ever.
		if currentScore > bestScore {
			bestAssignments = copyAssignments(assignments)
			bestUnassigned = copyStrings(unassigned)
			bestScore = currentScore
		}

		// Add to tabu list.
		entry := tabuEntry{
			WorkItemID:   chosen.WorkItemID,
			FromResource: chosen.SwapFrom,
			ToResource:   chosen.TargetResource,
		}
		tabuList = append(tabuList, entry)
		if len(tabuList) > listSize {
			tabuList = tabuList[1:]
		}
	}

	stats := plan.Statistics{
		Algorithm:            "tabu-search",
		DurationMs:           time.Since(startTime).Milliseconds(),
		Iterations:           iterationsRun,
		CandidatesEvaluated:  candidatesEvaluated,
		ImprovementsAccepted: improvementsAccepted,
		FinalObjectiveScore:  bestScore,
	}

	return buildResult(bestAssignments, bestUnassigned, totalItems, capacities, ctx, stats)
}


// isTabu checks if a move matches any entry in the tabu list.
func isTabu(list []tabuEntry, m CandidateMove) bool {
	for _, entry := range list {
		if entry.WorkItemID == m.WorkItemID &&
			entry.FromResource == m.SwapFrom &&
			entry.ToResource == m.TargetResource {
			return true
		}
	}
	return false
}
