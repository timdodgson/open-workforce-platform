package optimisation

import (
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/plan"
)

// hillClimbingAlgorithm implements Algorithm using local search.
type hillClimbingAlgorithm struct{}

func init() {
	register(&hillClimbingAlgorithm{})
}

func (h *hillClimbingAlgorithm) Name() string {
	return "hill-climbing"
}

func (h *hillClimbingAlgorithm) Solve(ctx OptimisationContext) (plan.OptimisedPlan, error) {
	items := ctx.Items()
	capacities := ctx.Resources()
	priorities := ctx.WorkItems()

	if err := validate(items, capacities); err != nil {
		return plan.OptimisedPlan{}, err
	}

	sorted := orderByPriority(items, priorities)
	assignments, unassigned, _ := assignItems(sorted, capacities, priorities, ctx)

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

	improved := true
	for improved {
		improved = false

		// Phase 1: Try placement moves for unassigned items.
		for ui := 0; ui < len(unassigned); ui++ {
			unassignedID := unassigned[ui]
			requiredSkill := requiredSkillOf[unassignedID]

			moves := GenerateMoves(unassignedID, requiredSkill, assignments, capacities, resourceIndex, requiredSkillOf, durationOf)

			for _, m := range moves {
				newAssignments, ok := ApplyMove(m, assignments)
				if ok && scheduleFeasible(newAssignments, capacities, priorities, ctx) {
					newScore := ObjectiveScore(newAssignments, ctx)
					if newScore > currentScore {
						assignments = newAssignments
						unassigned = append(unassigned[:ui], unassigned[ui+1:]...)
						currentScore = newScore
						improved = true
						break
					}
				}
			}

			if improved {
				break
			}
		}

		if improved {
			continue
		}

		// Phase 2: Try swap moves that might enable a placement.
		if len(unassigned) == 0 {
			break
		}

		swaps := GenerateSwapMoves(assignments, capacities, resourceIndex, requiredSkillOf, durationOf)
		for _, swap := range swaps {
			swapped, ok := ApplyMove(swap, copyAssignments(assignments))
			if !ok || !scheduleFeasible(swapped, capacities, priorities, ctx) {
				continue
			}

			for ui := 0; ui < len(unassigned); ui++ {
				unassignedID := unassigned[ui]
				requiredSkill := requiredSkillOf[unassignedID]

				placementMoves := GenerateMoves(unassignedID, requiredSkill, swapped, capacities, resourceIndex, requiredSkillOf, durationOf)
				for _, pm := range placementMoves {
					placed, ok := ApplyMove(pm, swapped)
					if ok && scheduleFeasible(placed, capacities, priorities, ctx) {
						newScore := ObjectiveScore(placed, ctx)
						if newScore > currentScore {
							assignments = placed
							unassigned = append(unassigned[:ui], unassigned[ui+1:]...)
							currentScore = newScore
							improved = true
							break
						}
					}
				}
				if improved {
					break
				}
			}
			if improved {
				break
			}
		}
	}

	return buildResult(assignments, unassigned, totalItems, capacities, ctx)
}
