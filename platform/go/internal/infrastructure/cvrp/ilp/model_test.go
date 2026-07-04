package ilp

import (
	"os"
	"strings"
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/cvrp"
)

func TestBuildModel_SmallInstance(t *testing.T) {
	ds := &cvrp.Dataset{
		Name:      "test-3",
		Dimension: 4,
		Capacity:  30,
		Depot:     cvrp.Depot{ID: 1, X: 0, Y: 0},
		Customers: []cvrp.Customer{
			{ID: 2, X: 10, Y: 0, Demand: 10},
			{ID: 3, X: 0, Y: 10, Demand: 15},
			{ID: 4, X: 10, Y: 10, Demand: 10},
		},
	}

	tmpFile, err := os.CreateTemp("", "cvrp-test-*.lp")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	info, err := BuildModel(ds, tmpFile.Name(), 0)
	if err != nil {
		t.Fatalf("BuildModel: %v", err)
	}

	// 3 customers, capacity 30, total demand 35 → need 2 vehicles.
	if info.Vehicles != 2 {
		t.Errorf("Vehicles = %d, want 2", info.Vehicles)
	}
	if info.Nodes != 4 {
		t.Errorf("Nodes = %d, want 4", info.Nodes)
	}
	if info.Variables <= 0 {
		t.Error("Expected positive variable count")
	}
	if info.Constraints <= 0 {
		t.Error("Expected positive constraint count")
	}

	// Read the file and verify it's valid LP format.
	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	lp := string(content)

	if !strings.Contains(lp, "Minimize") {
		t.Error("LP file missing Minimize section")
	}
	if !strings.Contains(lp, "Subject To") {
		t.Error("LP file missing Subject To section")
	}
	if !strings.Contains(lp, "Binary") {
		t.Error("LP file missing Binary section")
	}
	if !strings.Contains(lp, "End") {
		t.Error("LP file missing End")
	}

	// Check visit constraints exist.
	if !strings.Contains(lp, "visit_in_1:") {
		t.Error("Missing visit_in constraint for customer 1")
	}
	if !strings.Contains(lp, "visit_out_1:") {
		t.Error("Missing visit_out constraint for customer 1")
	}
	if !strings.Contains(lp, "depot_out:") {
		t.Error("Missing depot_out constraint")
	}
	if !strings.Contains(lp, "depot_in:") {
		t.Error("Missing depot_in constraint")
	}
	// MTZ constraint.
	if !strings.Contains(lp, "mtz_") {
		t.Error("Missing MTZ subtour elimination constraints")
	}

	t.Logf("Model: %d vars, %d constraints, %d nodes, %d vehicles",
		info.Variables, info.Constraints, info.Nodes, info.Vehicles)
}

func TestBuildModel_VehicleCount(t *testing.T) {
	// Total demand = 80, capacity = 25 → ceil(80/25) = 4 vehicles.
	ds := &cvrp.Dataset{
		Name:      "test-vehicles",
		Dimension: 5,
		Capacity:  25,
		Depot:     cvrp.Depot{ID: 1, X: 0, Y: 0},
		Customers: []cvrp.Customer{
			{ID: 2, X: 1, Y: 0, Demand: 20},
			{ID: 3, X: 2, Y: 0, Demand: 20},
			{ID: 4, X: 3, Y: 0, Demand: 20},
			{ID: 5, X: 4, Y: 0, Demand: 20},
		},
	}

	tmpFile, _ := os.CreateTemp("", "cvrp-test-*.lp")
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	info, err := BuildModel(ds, tmpFile.Name(), 0)
	if err != nil {
		t.Fatalf("BuildModel: %v", err)
	}
	if info.Vehicles != 4 {
		t.Errorf("Vehicles = %d, want 4 (80/25 rounded up)", info.Vehicles)
	}
}

func TestBuildModel_FromCVRPLIB(t *testing.T) {
	ds, err := cvrp.LoadDataset("../testdata/A-n10-k2.vrp")
	if err != nil {
		t.Skipf("Test instance not found: %v", err)
	}

	tmpFile, _ := os.CreateTemp("", "cvrp-test-*.lp")
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	info, err := BuildModel(ds, tmpFile.Name(), 0)
	if err != nil {
		t.Fatalf("BuildModel: %v", err)
	}

	t.Logf("A-n10-k2: %d vars, %d constraints, %d nodes, %d vehicles",
		info.Variables, info.Constraints, info.Nodes, info.Vehicles)

	// Should produce a valid model.
	if info.Variables <= 0 || info.Constraints <= 0 {
		t.Error("Invalid model dimensions")
	}
}
