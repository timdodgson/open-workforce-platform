package optimisation

import (
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/plan"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/workitem"
)

// hillClimbingAlgorithm implements Algorithm using local search.
type hillClimbingAlgorithm struct{}

func init() {
	register(&hillClimbingAlgorithm{})
}

func (h *hillClimbingAlgorithm) Name() string {
	return "hill-climbing"
}

// Solve starts from the constructive solution and attempts to improve it
// by exploring neighbouring assignment plans.
//
// It first tries placement moves for unassigned items. If no placement
// improves the score, it tries swap moves that might enable a subsequent
// placement. The algorithm stops when no improving combination can be found.
func (h *hillClimbingAlgorithm) Solve(items []workitem.WorkItem, capacities []ResourceCapacity, priorities []WorkItemPriority) (plan.OptimisedPlan, error) {
	if err := validate(items, capacities); err != nil {
		return plan.OptimisedPlan{}, err
	}

	sorted := orderByPriority(items, priorities)
	assignments, unassigned := assignItems(sorted, capacities, priorities)

	requiredSkillOf := make(map[string]string, len(priorities))
	for _, p := range priorities {
		requiredSkillOf[p.WorkItemID] = p.RequiredSkill
	}

	resourceIndex := make(map[string]int, len(capacities))
	for i, rc := range capacities {
		resourceIndex[rc.ResourceID] = i
	}

	totalItems := len(items)
	currentScore := calculateScore(len(assignments), totalItems)

	improved := true
	for improved {
		improved = false

		// Phase 1: Try placement moves for unassigned items.
		for ui := 0; ui < len(unassigned); ui++ {
			unassignedID := unassigned[ui]
			requiredSkill := requiredSkillOf[unassignedID]

			moves := GenerateMoves(unassignedID, requiredSkill, assignments, capacities, resourceIndex, requiredSkillOf)

			for _, m := range moves {
				newAssignments, ok := ApplyMove(m, assignments)
				if ok {
					newScore := calculateScore(len(newAssignments), totalItems)
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

		swaps := GenerateSwapMoves(assignments, capacities, resourceIndex, requiredSkillOf)
		for _, swap := range swaps {
			swapped, ok := ApplyMove(swap, copyAssignments(assignments))
			if !ok {
				continue
			}

			// After the swap, try to place an unassigned item.
			for ui := 0; ui < len(unassigned); ui++ {
				unassignedID := unassigned[ui]
				requiredSkill := requiredSkillOf[unassignedID]

				placementMoves := GenerateMoves(unassignedID, requiredSkill, swapped, capacities, resourceIndex, requiredSkillOf)
				for _, pm := range placementMoves {
					placed, ok := ApplyMove(pm, swapped)
					if ok {
						newScore := calculateScore(len(placed), totalItems)
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

	return buildResult(assignments, unassigned, totalItems, capacities)
}
