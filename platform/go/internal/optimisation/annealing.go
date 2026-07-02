package optimisation

import (
	"math"
	"math/rand"
	"time"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/assignment"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/plan"
)

type simulatedAnnealingAlgorithm struct{}

func init() {
	register(&simulatedAnnealingAlgorithm{})
}

func (sa *simulatedAnnealingAlgorithm) Name() string {
	return "simulated-annealing"
}

func (sa *simulatedAnnealingAlgorithm) Solve(ctx OptimisationContext) (plan.OptimisedPlan, error) {
	startTime := time.Now()
	items := ctx.Items()
	capacities := ctx.Resources()
	priorities := ctx.WorkItems()

	if err := validate(items, capacities); err != nil {
		return plan.OptimisedPlan{}, err
	}

	sorted := orderByPriority(items, priorities)
	assignments, unassigned, _ := assignItems(sorted, capacities, priorities, ctx)

	// Warm start: if existing plan provided, use it instead.
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

	// SA parameters from profile.
	profile := ctx.Profile()
	maxIter := profile.SAMaxIterations
	temperature := profile.SAInitialTemperature
	coolingRate := profile.SACoolingRate
	minTemp := profile.SAMinTemperature

	// Deterministic PRNG (fixed seed for reproducibility).
	rng := rand.New(rand.NewSource(42))

	candidatesEvaluated := 0
	improvementsAccepted := 0
	iterationsRun := 0

	for iteration := 0; iteration < maxIter; iteration++ {
		if temperature < minTemp {
			break
		}
		iterationsRun++

		// Generate candidate moves.
		var allMoves []CandidateMove

		// Placement moves for unassigned items.
		for _, uid := range unassigned {
			skill := requiredSkillOf[uid]
			moves := GenerateMoves(uid, skill, assignments, capacities, resourceIndex, requiredSkillOf, durationOf)
			allMoves = append(allMoves, moves...)
		}

		// Neighbourhood moves.
		allMoves = append(allMoves, GenerateAllNeighbourhoodMoves(assignments, capacities, resourceIndex, requiredSkillOf, durationOf)...)

		if len(allMoves) == 0 {
			break
		}

		// Select a candidate deterministically using PRNG.
		moveIdx := rng.Intn(len(allMoves))
		candidate := allMoves[moveIdx]

		candidatesEvaluated++
		trial, ok := ApplyMove(candidate, copyAssignments(assignments))
		if !ok || !scheduleFeasible(trial, capacities, priorities, ctx) {
			// Cool even on failed attempt.
			temperature *= coolingRate
			continue
		}

		candidateScore := ObjectiveScore(trial, ctx)
		delta := float64(candidateScore - currentScore)

		// Metropolis acceptance criterion.
		accept := false
		if delta > 0 {
			// Improving move — always accept.
			accept = true
		} else if delta == 0 {
			// Equal — accept with 50% probability.
			accept = rng.Float64() < 0.5
		} else {
			// Worsening move — accept with probability exp(delta / temperature).
			probability := math.Exp(delta / temperature)
			accept = rng.Float64() < probability
		}

		if accept {
			assignments = trial
			currentScore = candidateScore
			improvementsAccepted++

			// Update unassigned list if this was a placement.
			if candidate.Type == Placement || candidate.Type == Displacement {
				for i, uid := range unassigned {
					if uid == candidate.WorkItemID {
						unassigned = append(unassigned[:i], unassigned[i+1:]...)
						break
					}
				}
			}

			// Track best-ever.
			if currentScore > bestScore {
				bestAssignments = copyAssignments(assignments)
				bestUnassigned = copyStrings(unassigned)
				bestScore = currentScore
			}
		}

		// Cool temperature.
		temperature *= coolingRate
	}

	stats := plan.Statistics{
		Algorithm:            "simulated-annealing",
		DurationMs:           time.Since(startTime).Milliseconds(),
		Iterations:           iterationsRun,
		CandidatesEvaluated:  candidatesEvaluated,
		ImprovementsAccepted: improvementsAccepted,
		FinalObjectiveScore:  bestScore,
	}

	return buildResult(bestAssignments, bestUnassigned, totalItems, capacities, ctx, stats)
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
