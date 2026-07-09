package ilp

import (
	"os"
	"strings"
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/jobshop"
)

func TestBuildModel_SmallInstance(t *testing.T) {
	ds := &jobshop.Dataset{
		Name:     "test-2x2",
		Jobs:     2,
		Machines: 2,
		JobList: []jobshop.Job{
			{ID: 0, Operations: []jobshop.Operation{
				{JobID: 0, OpIndex: 0, Machine: 0, Duration: 3},
				{JobID: 0, OpIndex: 1, Machine: 1, Duration: 2},
			}},
			{ID: 1, Operations: []jobshop.Operation{
				{JobID: 1, OpIndex: 0, Machine: 1, Duration: 2},
				{JobID: 1, OpIndex: 1, Machine: 0, Duration: 4},
			}},
		},
	}
	for _, job := range ds.JobList {
		ds.AllOps = append(ds.AllOps, job.Operations...)
	}

	tmpFile, err := os.CreateTemp("", "jss-test-*.lp")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	info, err := BuildModel(ds, tmpFile.Name())
	if err != nil {
		t.Fatalf("BuildModel: %v", err)
	}

	if info.Operations != 4 {
		t.Errorf("Operations = %d, want 4", info.Operations)
	}

	content, _ := os.ReadFile(tmpFile.Name())
	lp := string(content)
	for _, want := range []string{"Minimize", "obj: C", "prec_0_1:", "disj_", "Binary", "End"} {
		if !strings.Contains(lp, want) {
			t.Errorf("LP missing %q", want)
		}
	}
}

func TestBuildModel_FromFT06(t *testing.T) {
	ds, err := jobshop.LoadDataset("testdata/ft06.txt")
	if err != nil {
		t.Skipf("testdata not found: %v", err)
	}

	tmpFile, _ := os.CreateTemp("", "jss-test-*.lp")
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	info, err := BuildModel(ds, tmpFile.Name())
	if err != nil {
		t.Fatalf("BuildModel: %v", err)
	}

	if info.Operations != 36 {
		t.Errorf("Operations = %d, want 36", info.Operations)
	}
	t.Logf("ft06: %d vars, %d constraints", info.Variables, info.Constraints)
}
