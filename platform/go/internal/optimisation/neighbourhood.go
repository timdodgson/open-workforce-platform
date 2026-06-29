package optimisation

import (
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/assignment"
)

// MoveType identifies the kind of candidate move.
type MoveType int

const (
	// Placement places an unassigned work item directly onto a resource.
	Placement MoveType = iota
	// Displacement moves an existing item to make room, then places a new item.
	Displacement
	// SwapMove exchanges two assigned items between their resources.
	SwapMove
)

// CandidateMove describes a candidate change to an assignment plan.
//
// The Type field determines which other fields are relevant:
//
//	Placement:    WorkItemID, TargetResource
//	Displacement: WorkItemID, TargetResource, DisplacedItemID, DisplacedTarget
//	SwapMove:     WorkItemID, TargetResource, SwapItemID, SwapFrom
type CandidateMove struct {
	Type MoveType

	// WorkItemID is the item being placed (Placement/Displacement)
	// or the first item in a swap (SwapMove).
	WorkItemID     string
	TargetResource string

	// Displacement: the existing item being moved and where it goes.
	DisplacedItemID string
	DisplacedTarget string

	// SwapMove: the second item (currently on TargetResource) and
	// where WorkItemID currently is (SwapItemID goes here).
	SwapItemID string
	SwapFrom   string
}

// IsDisplacement returns true if this move displaces an existing item.
func (m CandidateMove) IsDisplacement() bool {
	return m.Type == Displacement
}

// IsSwap returns true if this move exchanges two assigned items.
func (m CandidateMove) IsSwap() bool {
	return m.Type == SwapMove
}

// GenerateMoves returns all valid candidate moves that would place the given
// unassigned work item onto a resource.
func GenerateMoves(
	workItemID string,
	requiredSkill string,
	assignments []assignment.Assignment,
	capacities []ResourceCapacity,
	resourceIndex map[string]int,
	requiredSkillOf map[string]string,
	durationOf map[string]int,
) []CandidateMove {
	remaining := computeRemaining(assignments, capacities, resourceIndex, durationOf)
	duration := getDuration(durationOf, workItemID)
	var moves []CandidateMove

	// Direct placement moves.
	for i, rc := range capacities {
		if canAccept(rc, remaining[i], requiredSkill, duration) {
			moves = append(moves, CandidateMove{
				WorkItemID:     workItemID,
				TargetResource: rc.ResourceID,
			})
		}
	}

	// Displacement moves.
	for _, existing := range assignments {
		srcIdx := resourceIndex[existing.ResourceID()]
		srcRC := capacities[srcIdx]

		if !srcRC.Available {
			continue
		}
		if requiredSkill != "" && !hasSkill(srcRC.Skills, requiredSkill) {
			continue
		}
		// Check if source can't fit the new item (otherwise direct placement would work).
		if remaining[srcIdx] >= duration {
			continue
		}

		existingSkill := requiredSkillOf[existing.WorkItemID()]
		existingDuration := getDuration(durationOf, existing.WorkItemID())

		// After removing existing item, can the new item fit?
		if remaining[srcIdx]+existingDuration < duration {
			continue
		}

		for di, destRC := range capacities {
			if di == srcIdx {
				continue
			}
			if canAccept(destRC, remaining[di], existingSkill, existingDuration) {
				moves = append(moves, CandidateMove{
					Type:            Displacement,
					WorkItemID:      workItemID,
					TargetResource:  srcRC.ResourceID,
					DisplacedItemID: existing.WorkItemID(),
					DisplacedTarget: destRC.ResourceID,
				})
			}
		}
	}

	return moves
}

// ApplyMove applies a candidate move to an assignments slice, returning the new slice.
func ApplyMove(m CandidateMove, assignments []assignment.Assignment) ([]assignment.Assignment, bool) {
	if m.IsSwap() {
		return applySwap(m, assignments)
	}

	if m.IsDisplacement() {
		found := false
		for i, a := range assignments {
			if a.WorkItemID() == m.DisplacedItemID && a.ResourceID() == m.TargetResource {
				moved, err := assignment.New(m.DisplacedTarget, m.DisplacedItemID)
				if err != nil {
					return assignments, false
				}
				assignments[i] = moved
				found = true
				break
			}
		}
		if !found {
			return assignments, false
		}
	}

	placed, err := assignment.New(m.TargetResource, m.WorkItemID)
	if err != nil {
		return assignments, false
	}
	assignments = append(assignments, placed)
	return assignments, true
}

// applySwap exchanges two assigned items between their resources.
func applySwap(m CandidateMove, assignments []assignment.Assignment) ([]assignment.Assignment, bool) {
	foundA := false
	foundB := false

	for i, a := range assignments {
		if !foundA && a.WorkItemID() == m.WorkItemID && a.ResourceID() == m.SwapFrom {
			moved, err := assignment.New(m.TargetResource, m.WorkItemID)
			if err != nil {
				return assignments, false
			}
			assignments[i] = moved
			foundA = true
		} else if !foundB && a.WorkItemID() == m.SwapItemID && a.ResourceID() == m.TargetResource {
			moved, err := assignment.New(m.SwapFrom, m.SwapItemID)
			if err != nil {
				return assignments, false
			}
			assignments[i] = moved
			foundB = true
		}
		if foundA && foundB {
			break
		}
	}

	if !foundA || !foundB {
		return assignments, false
	}
	return assignments, true
}

// GenerateSwapMoves returns all valid swap candidate moves for the current assignments.
//
// A swap is valid if both resources can accommodate the incoming item's duration
// after the outgoing item is removed.
func GenerateSwapMoves(
	assignments []assignment.Assignment,
	capacities []ResourceCapacity,
	resourceIndex map[string]int,
	requiredSkillOf map[string]string,
	durationOf map[string]int,
) []CandidateMove {
	remaining := computeRemaining(assignments, capacities, resourceIndex, durationOf)
	var moves []CandidateMove

	for i := 0; i < len(assignments); i++ {
		for j := i + 1; j < len(assignments); j++ {
			a := assignments[i]
			b := assignments[j]

			if a.ResourceID() == b.ResourceID() {
				continue
			}

			aIdx := resourceIndex[a.ResourceID()]
			bIdx := resourceIndex[b.ResourceID()]
			aRC := capacities[aIdx]
			bRC := capacities[bIdx]

			aSkill := requiredSkillOf[a.WorkItemID()]
			bSkill := requiredSkillOf[b.WorkItemID()]
			aDuration := getDuration(durationOf, a.WorkItemID())
			bDuration := getDuration(durationOf, b.WorkItemID())

			// Check if A can go to B's resource (after B leaves).
			if !bRC.Available {
				continue
			}
			if aSkill != "" && !hasSkill(bRC.Skills, aSkill) {
				continue
			}
			// Remaining at B + B's duration (freed) must fit A's duration.
			if remaining[bIdx]+bDuration < aDuration {
				continue
			}

			// Check if B can go to A's resource (after A leaves).
			if !aRC.Available {
				continue
			}
			if bSkill != "" && !hasSkill(aRC.Skills, bSkill) {
				continue
			}
			// Remaining at A + A's duration (freed) must fit B's duration.
			if remaining[aIdx]+aDuration < bDuration {
				continue
			}

			moves = append(moves, CandidateMove{
				Type:           SwapMove,
				WorkItemID:     a.WorkItemID(),
				TargetResource: b.ResourceID(),
				SwapItemID:     b.WorkItemID(),
				SwapFrom:       a.ResourceID(),
			})
		}
	}

	return moves
}

// computeRemaining calculates remaining capacity per resource given current assignments.
func computeRemaining(assignments []assignment.Assignment, capacities []ResourceCapacity, resourceIndex map[string]int, durationOf map[string]int) []int {
	remaining := make([]int, len(capacities))
	for i, rc := range capacities {
		remaining[i] = rc.Capacity
	}
	for _, a := range assignments {
		if idx, ok := resourceIndex[a.ResourceID()]; ok {
			remaining[idx] -= getDuration(durationOf, a.WorkItemID())
		}
	}
	return remaining
}

// getDuration returns the duration for a work item, defaulting to 1 if not found.
func getDuration(durationOf map[string]int, workItemID string) int {
	if d, ok := durationOf[workItemID]; ok && d > 0 {
		return d
	}
	return 1
}
