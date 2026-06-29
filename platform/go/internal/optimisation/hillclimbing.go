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
// It delegates neighbour generation to the neighbourhood component and
// accepts the first valid move that would increase the number of assignments.
//
// The algorithm stops when no improving neighbour can be found.
// It is deterministic — no randomness is involved.
func (h *hillClimbingAlgorithm) Solve(items []workitem.WorkItem, capacities []ResourceCapacity, priorities []WorkItemPriority) (plan.OptimisedPlan, error) {
	if err := validate(items, capacities); err != nil {
		return plan.OptimisedPlan{}, err
	}

	// Start from the constructive solution.
	sorted := orderByPriority(items, priorities)
	assignments, unassigned := assignItems(sorted, capacities, priorities)

	// Build lookups.
	requiredSkillOf := make(map[string]string, len(priorities))
	for _, p := range priorities {
		requiredSkillOf[p.WorkItemID] = p.RequiredSkill
	}

	resourceIndex := make(map[string]int, len(capacities))
	for i, rc := range capacities {
		resourceIndex[rc.ResourceID] = i
	}

	// Hill climbing loop: accept first improving move per iteration.
	improved := true
	for improved {
		improved = false

		for ui := 0; ui < len(unassigned); ui++ {
			unassignedID := unassigned[ui]
			requiredSkill := requiredSkillOf[unassignedID]

			moves := GenerateMoves(unassignedID, requiredSkill, assignments, capacities, resourceIndex, requiredSkillOf)

			for _, m := range moves {
				newAssignments, ok := ApplyMove(m, assignments)
				if ok {
					assignments = newAssignments
					unassigned = append(unassigned[:ui], unassigned[ui+1:]...)
					improved = true
					break
				}
			}

			if improved {
				break // restart from the top
			}
		}
	}

	return buildResult(assignments, unassigned, len(items), capacities)
}
