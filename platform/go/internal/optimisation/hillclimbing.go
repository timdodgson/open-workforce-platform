package optimisation

import (
	"time"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/plan"
)

type hillClimbingAlgorithm struct{}

func init() {
	register(&hillClimbingAlgorithm{})
}

func (h *hillClimbingAlgorithm) Name() string {
	return "hill-climbing"
}

func (h *hillClimbingAlgorithm) Solve(ctx OptimisationContext) (plan.OptimisedPlan, error) {
	startTime := time.Now()
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

	iterations := 0
	candidatesEvaluated := 0
	improvementsAccepted := 0

	improved := true
	for improved {
		improved = false
		iterations++

		for ui := 0; ui < len(unassigned); ui++ {
			unassignedID := unassigned[ui]
			requiredSkill := requiredSkillOf[unassignedID]

			moves := GenerateMoves(unassignedID, requiredSkill, assignments, capacities, resourceIndex, requiredSkillOf, durationOf)

			for _, m := range moves {
				candidatesEvaluated++
				newAssignments, ok := ApplyMove(m, assignments)
				if ok && scheduleFeasible(newAssignments, capacities, priorities, ctx) {
					newScore := ObjectiveScore(newAssignments, ctx)
					if newScore > currentScore {
						assignments = newAssignments
						unassigned = append(unassigned[:ui], unassigned[ui+1:]...)
						currentScore = newScore
						improvementsAccepted++
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

		if len(unassigned) == 0 {
			break
		}

		swaps := GenerateSwapMoves(assignments, capacities, resourceIndex, requiredSkillOf, durationOf)
		for _, swap := range swaps {
			candidatesEvaluated++
			swapped, ok := ApplyMove(swap, copyAssignments(assignments))
			if !ok || !scheduleFeasible(swapped, capacities, priorities, ctx) {
				continue
			}

			for ui := 0; ui < len(unassigned); ui++ {
				unassignedID := unassigned[ui]
				requiredSkill := requiredSkillOf[unassignedID]

				placementMoves := GenerateMoves(unassignedID, requiredSkill, swapped, capacities, resourceIndex, requiredSkillOf, durationOf)
				for _, pm := range placementMoves {
					candidatesEvaluated++
					placed, ok := ApplyMove(pm, swapped)
					if ok && scheduleFeasible(placed, capacities, priorities, ctx) {
						newScore := ObjectiveScore(placed, ctx)
						if newScore > currentScore {
							assignments = placed
							unassigned = append(unassigned[:ui], unassigned[ui+1:]...)
							currentScore = newScore
							improvementsAccepted++
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

	stats := plan.Statistics{
		Algorithm:            "hill-climbing",
		DurationMs:           time.Since(startTime).Milliseconds(),
		Iterations:           iterations,
		CandidatesEvaluated:  candidatesEvaluated,
		ImprovementsAccepted: improvementsAccepted,
		FinalObjectiveScore:  currentScore,
	}

	return buildResult(assignments, unassigned, totalItems, capacities, ctx, stats)
}
