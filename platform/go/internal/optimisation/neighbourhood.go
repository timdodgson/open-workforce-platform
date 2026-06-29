package optimisation

import (
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/assignment"
)

// CandidateMove describes a candidate change to an assignment plan.
//
// A candidate move places a work item onto a target resource.
// If a displacement is required (the target is full), the move also
// specifies which existing item would be moved and where.
type CandidateMove struct {
	// WorkItemID is the item being placed.
	WorkItemID     string
	TargetResource string

	// Displacement fields are set when an existing item must be moved
	// to free a slot. Empty strings indicate direct placement.
	DisplacedItemID string
	DisplacedTarget string
}

// IsDisplacement returns true if this move requires displacing an existing item.
func (m CandidateMove) IsDisplacement() bool {
	return m.DisplacedItemID != ""
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
