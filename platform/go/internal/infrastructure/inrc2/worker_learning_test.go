package inrc2

import (
	"os"
	"strings"
	"testing"
)

func TestWorkerLearningCSVHeader(t *testing.T) {
	header := WorkerLearningCSVHeader()
	fields := strings.Split(header, ",")

	// Should have all fields.
	if len(fields) != 38 {
		t.Errorf("Header has %d fields, want 38", len(fields))
	}

	// Check key fields exist.
	must := []string{
		"problem_type", "instance", "algorithm", "week", "depth",
		"global_best", "improved", "improvement_amount", "roi",
	}
	for _, m := range must {
		found := false
		for _, f := range fields {
			if f == m {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Header missing field: %s", m)
		}
	}
}

func TestWorkerLearningCSVRow(t *testing.T) {
	r := WorkerLearningRecord{
		ProblemType:        "nrp",
		Instance:           "n012w8",
		Algorithm:          "sa",
		RunSeed:            42,
		Week:               3,
		Phase:              "search",
		Depth:              5,
		ParentWorkerID:     12,
		FamilyID:           2,
		BeamRank:           1,
		BeamScore:          3500,
		Entropy:            2.3,
		Diversity:          0.15,
		BeamHealth:         0.85,
		Temperature:        5.0,
		IterationsAlloc:    60000,
		GlobalBest:         3200,
		ParentObjective:    3400,
		DistanceFromBest:   200,
		PlateauLength:      15000,
		RecentImprovRate:   0.5,
		WorkerCount:        48,
		ActiveFamilies:     6,
		Improved:           true,
		ProducedGlobalBest: true,
		ImprovementAmount:  150,
		FinalObjective:     3250,
		RuntimeMs:          80,
		CandidatesEval:     60000,
		Accepted:           3200,
		Rejected:           45000,
		PlateauCount:       3,
		BranchesSpawned:    2,
	}

	r.ComputeDerived()

	row := WorkerLearningCSVRow(r)
	fields := strings.Split(row, ",")

	if len(fields) != 38 {
		t.Errorf("Row has %d fields, want 38", len(fields))
	}

	// Check first field is problem type.
	if fields[0] != "nrp" {
		t.Errorf("field[0] = %q, want nrp", fields[0])
	}

	// Check derived ROI is computed.
	if r.ROI <= 0 {
		t.Errorf("ROI = %f, want > 0", r.ROI)
	}
	if r.ImprovPer100K <= 0 {
		t.Errorf("ImprovPer100K = %f, want > 0", r.ImprovPer100K)
	}
}

func TestWriteWorkerLearningCSV(t *testing.T) {
	records := []WorkerLearningRecord{
		{
			ProblemType: "cvrp", Instance: "A-n32-k5", Algorithm: "sa",
			Week: 1, Depth: 0, Improved: true, ImprovementAmount: 50,
			RuntimeMs: 100, CandidatesEval: 500000, FinalObjective: 800,
		},
		{
			ProblemType: "cvrp", Instance: "A-n32-k5", Algorithm: "lahc",
			Week: 1, Depth: 0, Improved: true, ImprovementAmount: 80,
			RuntimeMs: 95, CandidatesEval: 500000, FinalObjective: 784,
			ProducedGlobalBest: true,
		},
	}

	path := t.TempDir() + "/worker_learning.csv"
	err := WriteWorkerLearningCSV(path, records)
	if err != nil {
		t.Fatalf("WriteWorkerLearningCSV: %v", err)
	}

	data, _ := os.ReadFile(path)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 3 { // header + 2 rows
		t.Errorf("got %d lines, want 3", len(lines))
	}

	// Verify header.
	if !strings.HasPrefix(lines[0], "problem_type,") {
		t.Errorf("header doesn't start with problem_type: %s", lines[0][:30])
	}
}

func TestComputeDerived(t *testing.T) {
	r := WorkerLearningRecord{
		ImprovementAmount: 200,
		RuntimeMs:         50,
		CandidatesEval:    1000000, // 1M = 10 × 100K
	}

	r.ComputeDerived()

	// ROI = 200 / 50 = 4.0
	if r.ROI != 4.0 {
		t.Errorf("ROI = %f, want 4.0", r.ROI)
	}

	// ImprovPerCPU = 200 / 50 * 1000 = 4000
	if r.ImprovPerCPU != 4000.0 {
		t.Errorf("ImprovPerCPU = %f, want 4000.0", r.ImprovPerCPU)
	}

	// ImprovPer100K = 200 / 10 = 20
	if r.ImprovPer100K != 20.0 {
		t.Errorf("ImprovPer100K = %f, want 20.0", r.ImprovPer100K)
	}
}
