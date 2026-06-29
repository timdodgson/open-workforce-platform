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
//
// It considers:
//   - direct placement (resource has available capacity and matching skill)
//   - displacement (move an existing item to another resource to free a slot)
//
// The neighbourhood does not score or choose moves. That is the algorithm's
// responsibility.
func GenerateMoves(
	workItemID string,
	requiredSkill string,
	assignments []assignment.Assignment,
	capacities []ResourceCapacity,
	resourceIndex map[string]int,
	requiredSkillOf map[string]string,
) []CandidateMove {
	remaining := computeRemaining(assignments, capacities, resourceIndex)
	var moves []CandidateMove

	// Direct placement moves.
	for i, rc := range capacities {
		if canAccept(rc, remaining[i], requiredSkill) {
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
		// Source is full (otherwise direct placement above would have found it).
		if remaining[srcIdx] > 0 {
			continue
		}

		existingSkill := requiredSkillOf[existing.WorkItemID()]

		for di, destRC := range capacities {
			if di == srcIdx {
				continue
			}
			if canAccept(destRC, remaining[di], existingSkill) {
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
//
// This is a helper for algorithms. The neighbourhood generates moves;
// algorithms decide which to apply.
func ApplyMove(m CandidateMove, assignments []assignment.Assignment) ([]assignment.Assignment, bool) {
	if m.IsSwap() {
		return applySwap(m, assignments)
	}

	if m.IsDisplacement() {
		// Find and relocate the displaced item.
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

	// Place the work item.
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
// A swap exchanges two assigned items between their resources.
// A swap is valid only if both resulting assignments respect availability and skills.
// Capacity is unchanged by a swap (one item leaves, one arrives).
func GenerateSwapMoves(
	assignments []assignment.Assignment,
	capacities []ResourceCapacity,
	resourceIndex map[string]int,
	requiredSkillOf map[string]string,
) []CandidateMove {
	var moves []CandidateMove

	for i := 0; i < len(assignments); i++ {
		for j := i + 1; j < len(assignments); j++ {
			a := assignments[i]
			b := assignments[j]

			// Skip if both are on the same resource — nothing to swap.
			if a.ResourceID() == b.ResourceID() {
				continue
			}

			aIdx := resourceIndex[a.ResourceID()]
			bIdx := resourceIndex[b.ResourceID()]
			aRC := capacities[aIdx]
			bRC := capacities[bIdx]

			aSkill := requiredSkillOf[a.WorkItemID()]
			bSkill := requiredSkillOf[b.WorkItemID()]

			// Check if A can go to B's resource.
			if !bRC.Available {
				continue
			}
			if aSkill != "" && !hasSkill(bRC.Skills, aSkill) {
				continue
			}

			// Check if B can go to A's resource.
			if !aRC.Available {
				continue
			}
			if bSkill != "" && !hasSkill(aRC.Skills, bSkill) {
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
func computeRemaining(assignments []assignment.Assignment, capacities []ResourceCapacity, resourceIndex map[string]int) []int {
	remaining := make([]int, len(capacities))
	for i, rc := range capacities {
		remaining[i] = rc.Capacity
	}
	for _, a := range assignments {
		if idx, ok := resourceIndex[a.ResourceID()]; ok {
			remaining[idx]--
		}
	}
	return remaining
}
