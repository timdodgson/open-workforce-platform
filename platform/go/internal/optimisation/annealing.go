package optimisation

import (
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/assignment"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/plan"
)

type simulatedAnnealingAlgorithm struct{}

const maxIterations = 100

func init() {
	register(&simulatedAnnealingAlgorithm{})
}

func (sa *simulatedAnnealingAlgorithm) Name() string {
	return "simulated-annealing"
}

func (sa *simulatedAnnealingAlgorithm) Solve(ctx OptimisationContext) (plan.OptimisedPlan, error) {
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
	bestAssignments := copyAssignments(assignments)
	bestUnassigned := copyStrings(unassigned)
	bestScore := ObjectiveScore(bestAssignments, ctx)

	for iteration := 0; iteration < maxIterations; iteration++ {
		hot := iteration < maxIterations/2
		moved := false

		for ui := 0; ui < len(unassigned); ui++ {
			unassignedID := unassigned[ui]
			requiredSkill := requiredSkillOf[unassignedID]

			moves := GenerateMoves(unassignedID, requiredSkill, assignments, capacities, resourceIndex, requiredSkillOf, durationOf)

			for _, m := range moves {
				newAssignments, ok := ApplyMove(m, assignments)
				if !ok || !scheduleFeasible(newAssignments, capacities, priorities, ctx) {
					continue
				}

				newScore := ObjectiveScore(newAssignments, ctx)

				if newScore > bestScore || hot {
					assignments = newAssignments
					unassigned = append(unassigned[:ui], unassigned[ui+1:]...)

					if newScore > bestScore {
						bestAssignments = copyAssignments(assignments)
						bestUnassigned = copyStrings(unassigned)
						bestScore = newScore
					}

					moved = true
					break
				}
			}

			if moved {
				break
			}
		}

		if moved {
			continue
		}

		swaps := GenerateSwapMoves(assignments, capacities, resourceIndex, requiredSkillOf, durationOf)
		for _, swap := range swaps {
			newAssignments, ok := ApplyMove(swap, copyAssignments(assignments))
			if !ok || !scheduleFeasible(newAssignments, capacities, priorities, ctx) {
				continue
			}

			newScore := ObjectiveScore(newAssignments, ctx)

			if newScore > bestScore || hot {
				assignments = newAssignments

				if newScore > bestScore {
					bestAssignments = copyAssignments(assignments)
					bestUnassigned = copyStrings(unassigned)
					bestScore = newScore
				}

				moved = true
				break
			}
		}

		if !moved {
			break
		}
	}

	return buildResult(bestAssignments, bestUnassigned, totalItems, capacities, ctx)
}

func copyAssignments(src []assignment.Assignment) []assignment.Assignment {
	cp := make([]assignment.Assignment, len(src))
	copy(cp, src)
	return cp
}

func copyStrings(src []string) []string {
	cp := make([]string, len(src))
	copy(cp, src)
	return cp
}
