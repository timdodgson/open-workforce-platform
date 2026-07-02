package inrc2_test

import (
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2"
)

func TestRosterHammingDistance_Identical(t *testing.T) {
	r := inrc2.NewRoster([]string{"A", "B"}, 7)
	r.Set(0, 0, inrc2.ShiftAssignment{ShiftType: "Early", Skill: "Nurse"})
	r.Set(1, 3, inrc2.ShiftAssignment{ShiftType: "Late", Skill: "HeadNurse"})

	dist := inrc2.RosterHammingDistance(r, r.Clone())
	if dist != 0.0 {
		t.Errorf("identical rosters should have distance 0, got %f", dist)
	}
}

func TestRosterHammingDistance_CompletelyDifferent(t *testing.T) {
	r1 := inrc2.NewRoster([]string{"A"}, 7)
	r2 := inrc2.NewRoster([]string{"A"}, 7)

	// Fill r1 with shifts, leave r2 empty.
	for d := 0; d < 7; d++ {
		r1.Set(0, d, inrc2.ShiftAssignment{ShiftType: "Early", Skill: "Nurse"})
	}

	dist := inrc2.RosterHammingDistance(r1, r2)
	if dist != 1.0 {
		t.Errorf("completely different rosters should have distance 1.0, got %f", dist)
	}
}

func TestRosterHammingDistance_Partial(t *testing.T) {
	r1 := inrc2.NewRoster([]string{"A", "B"}, 7)
	r2 := inrc2.NewRoster([]string{"A", "B"}, 7)

	// Same for nurse A, different for nurse B on one day.
	r1.Set(0, 0, inrc2.ShiftAssignment{ShiftType: "Early", Skill: "Nurse"})
	r2.Set(0, 0, inrc2.ShiftAssignment{ShiftType: "Early", Skill: "Nurse"})
	r1.Set(1, 0, inrc2.ShiftAssignment{ShiftType: "Late", Skill: "Nurse"})
	r2.Set(1, 0, inrc2.ShiftAssignment{ShiftType: "Night", Skill: "Nurse"})

	// Total cells: 2 nurses * 7 days = 14. One cell different.
	dist := inrc2.RosterHammingDistance(r1, r2)
	expected := 1.0 / 14.0
	if dist < expected-0.001 || dist > expected+0.001 {
		t.Errorf("expected distance ~%.4f, got %.4f", expected, dist)
	}
}

func TestRosterHammingDistance_NilRoster(t *testing.T) {
	r := inrc2.NewRoster([]string{"A"}, 7)

	dist := inrc2.RosterHammingDistance(r, nil)
	if dist != 1.0 {
		t.Errorf("nil roster should give distance 1.0, got %f", dist)
	}

	dist = inrc2.RosterHammingDistance(nil, r)
	if dist != 1.0 {
		t.Errorf("nil roster should give distance 1.0, got %f", dist)
	}
}

func TestRosterFingerprint_Deterministic(t *testing.T) {
	r := inrc2.NewRoster([]string{"A", "B"}, 7)
	r.Set(0, 0, inrc2.ShiftAssignment{ShiftType: "Early", Skill: "Nurse"})

	fp1 := inrc2.RosterFingerprint(r)
	fp2 := inrc2.RosterFingerprint(r)

	if fp1 != fp2 {
		t.Errorf("fingerprint should be deterministic: %s != %s", fp1, fp2)
	}
	if len(fp1) != 12 {
		t.Errorf("fingerprint should be 12 chars, got %d", len(fp1))
	}
}

func TestRosterFingerprint_DifferentForDifferentRosters(t *testing.T) {
	r1 := inrc2.NewRoster([]string{"A"}, 7)
	r2 := inrc2.NewRoster([]string{"A"}, 7)
	r1.Set(0, 0, inrc2.ShiftAssignment{ShiftType: "Early", Skill: "Nurse"})
	r2.Set(0, 0, inrc2.ShiftAssignment{ShiftType: "Late", Skill: "Nurse"})

	fp1 := inrc2.RosterFingerprint(r1)
	fp2 := inrc2.RosterFingerprint(r2)

	if fp1 == fp2 {
		t.Error("different rosters should have different fingerprints")
	}
}

func TestSearchTree_WinningLineage(t *testing.T) {
	tree := inrc2.SearchTree{
		Nodes: []inrc2.TreeNode{
			{WorkerID: 0, ParentWorkerID: -1, Depth: 0, BestPenalty: 500},
			{WorkerID: 1, ParentWorkerID: 0, Depth: 1, BestPenalty: 450},
			{WorkerID: 2, ParentWorkerID: 0, Depth: 1, BestPenalty: 480},
			{WorkerID: 3, ParentWorkerID: 1, Depth: 2, BestPenalty: 400, ProducedGlobalBest: true},
		},
	}

	lineage := tree.WinningLineage(400)
	if len(lineage) != 3 {
		t.Fatalf("expected lineage of 3, got %d", len(lineage))
	}
	if lineage[0].WorkerID != 0 {
		t.Errorf("first in lineage should be root (0), got %d", lineage[0].WorkerID)
	}
	if lineage[1].WorkerID != 1 {
		t.Errorf("second should be worker 1, got %d", lineage[1].WorkerID)
	}
	if lineage[2].WorkerID != 3 {
		t.Errorf("third should be winner (3), got %d", lineage[2].WorkerID)
	}
}

func TestSearchTree_DepthSummaries(t *testing.T) {
	tree := inrc2.SearchTree{
		Nodes: []inrc2.TreeNode{
			{WorkerID: 0, Depth: 0, BestPenalty: 500, ImprovementFromStart: 100},
			{WorkerID: 1, Depth: 1, BestPenalty: 450, ImprovementFromStart: 50},
			{WorkerID: 2, Depth: 1, BestPenalty: 400, ImprovementFromStart: 100, ProducedGlobalBest: true},
		},
	}

	summaries := tree.DepthSummaries()
	if len(summaries) != 2 {
		t.Fatalf("expected 2 depth levels, got %d", len(summaries))
	}
	if summaries[0].WorkerCount != 1 {
		t.Errorf("depth 0 should have 1 worker, got %d", summaries[0].WorkerCount)
	}
	if summaries[1].WorkerCount != 2 {
		t.Errorf("depth 1 should have 2 workers, got %d", summaries[1].WorkerCount)
	}
	if summaries[1].BestPenalty != 400 {
		t.Errorf("depth 1 best penalty should be 400, got %d", summaries[1].BestPenalty)
	}
}
