package ilp

import (
	"os"
	"strings"
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/vrptw"
)

func TestBuildModel_SmallInstance(t *testing.T) {
	ds := &vrptw.Dataset{
		Name:     "test-vrptw",
		Capacity: 30,
		Vehicles: 2,
		Depot: vrptw.Depot{
			ID: 0, X: 0, Y: 0, ReadyTime: 0, DueDate: 200, Service: 0,
		},
		Customers: []vrptw.Customer{
			{ID: 1, X: 10, Y: 0, Demand: 10, ReadyTime: 0, DueDate: 100, Service: 5},
			{ID: 2, X: 0, Y: 10, Demand: 15, ReadyTime: 20, DueDate: 120, Service: 5},
		},
	}

	tmpFile, err := os.CreateTemp("", "vrptw-test-*.lp")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	info, err := BuildModel(ds, tmpFile.Name(), 0)
	if err != nil {
		t.Fatalf("BuildModel: %v", err)
	}

	if info.Nodes != 3 {
		t.Errorf("Nodes = %d, want 3", info.Nodes)
	}
	if info.Variables <= 0 || info.Constraints <= 0 {
		t.Error("expected positive model size")
	}

	content, _ := os.ReadFile(tmpFile.Name())
	lp := string(content)
	for _, want := range []string{"Minimize", "Subject To", "Binary", "time_0_1:", "return_1:", "End"} {
		if !strings.Contains(lp, want) {
			t.Errorf("LP missing %q", want)
		}
	}
}

func TestBuildModel_FromSolomonSubset(t *testing.T) {
	ds, err := vrptw.LoadDataset("testdata/c101-25.txt")
	if err != nil {
		t.Skipf("testdata not found: %v", err)
	}

	tmpFile, _ := os.CreateTemp("", "vrptw-test-*.lp")
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	info, err := BuildModel(ds, tmpFile.Name(), 0)
	if err != nil {
		t.Fatalf("BuildModel: %v", err)
	}

	if info.Variables <= 0 || info.Constraints <= 0 {
		t.Error("invalid model dimensions")
	}
	t.Logf("c101-25: %d vars, %d constraints", info.Variables, info.Constraints)
}
