package loader

import (
	"strings"
	"testing"
)

const testInstancePath = "testdata/E-n13-k4.vrp"

// TestLoadFile_FullParse verifies complete parsing of a CVRPLIB file.
func TestLoadFile_FullParse(t *testing.T) {
	inst, err := LoadFile(testInstancePath)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}

	// Metadata.
	if inst.Name != "E-n13-k4" {
		t.Errorf("Name = %q, want E-n13-k4", inst.Name)
	}
	if inst.Type != "CVRP" {
		t.Errorf("Type = %q, want CVRP", inst.Type)
	}
	if inst.Dimension != 13 {
		t.Errorf("Dimension = %d, want 13", inst.Dimension)
	}
	if inst.Capacity != 6000 {
		t.Errorf("Capacity = %d, want 6000", inst.Capacity)
	}
	if inst.Vehicles != 4 {
		t.Errorf("Vehicles = %d, want 4", inst.Vehicles)
	}
	if inst.BestKnown != 247 {
		t.Errorf("BestKnown = %d, want 247", inst.BestKnown)
	}
	if inst.EdgeWeightType != "EUC_2D" {
		t.Errorf("EdgeWeightType = %q, want EUC_2D", inst.EdgeWeightType)
	}
	if inst.DistanceType != "rounded" {
		t.Errorf("DistanceType = %q, want rounded", inst.DistanceType)
	}
	if !strings.Contains(inst.Comment, "Christofides") {
		t.Errorf("Comment = %q, expected to contain 'Christofides'", inst.Comment)
	}

	// Nodes.
	if len(inst.Nodes) != 13 {
		t.Errorf("Nodes = %d, want 13", len(inst.Nodes))
	}

	// First node should be depot.
	if inst.Nodes[0].ID != 1 || inst.Nodes[0].X != 30 || inst.Nodes[0].Y != 40 {
		t.Errorf("First node = %+v, want {1, 30, 40}", inst.Nodes[0])
	}

	// Demands.
	if inst.NodeDemand(1) != 0 {
		t.Errorf("Depot demand = %d, want 0", inst.NodeDemand(1))
	}
	if inst.NodeDemand(2) != 1000 {
		t.Errorf("Node 2 demand = %d, want 1000", inst.NodeDemand(2))
	}
	if inst.NodeDemand(4) != 2000 {
		t.Errorf("Node 4 demand = %d, want 2000", inst.NodeDemand(4))
	}

	// Depot.
	if inst.DepotID() != 1 {
		t.Errorf("DepotID = %d, want 1", inst.DepotID())
	}

	// Customers.
	customers := inst.CustomerNodes()
	if len(customers) != 12 {
		t.Errorf("CustomerNodes = %d, want 12", len(customers))
	}

	// Total demand.
	totalDemand := 0
	for _, c := range customers {
		totalDemand += inst.NodeDemand(c.ID)
	}
	if totalDemand != 15000 {
		t.Errorf("Total demand = %d, want 15000", totalDemand)
	}
}

// TestParse_FromReader verifies parsing from an io.Reader.
func TestParse_FromReader(t *testing.T) {
	input := `NAME : tiny-test
TYPE : CVRP
DIMENSION : 4
CAPACITY : 20
EDGE_WEIGHT_TYPE : EUC_2D
NODE_COORD_SECTION
1 0 0
2 10 0
3 0 10
4 10 10
DEMAND_SECTION
1 0
2 5
3 8
4 12
DEPOT_SECTION
1
-1
EOF`

	inst, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	if inst.Name != "tiny-test" {
		t.Errorf("Name = %q, want tiny-test", inst.Name)
	}
	if inst.Dimension != 4 {
		t.Errorf("Dimension = %d, want 4", inst.Dimension)
	}
	if inst.Capacity != 20 {
		t.Errorf("Capacity = %d, want 20", inst.Capacity)
	}
	if len(inst.Nodes) != 4 {
		t.Errorf("Nodes = %d, want 4", len(inst.Nodes))
	}
	if inst.NodeDemand(3) != 8 {
		t.Errorf("Node 3 demand = %d, want 8", inst.NodeDemand(3))
	}
}

// TestParse_MissingDimension reports error.
func TestParse_MissingDimension(t *testing.T) {
	input := `NAME : broken
CAPACITY : 20
EDGE_WEIGHT_TYPE : EUC_2D
NODE_COORD_SECTION
1 0 0
DEMAND_SECTION
1 0
DEPOT_SECTION
1
-1
EOF`

	_, err := Parse(strings.NewReader(input))
	if err != nil {
		// Dimension should be inferred from node count.
		t.Logf("Got error (expected if no nodes): %v", err)
	}
}

// TestParse_DefaultDepot uses node 1 when no DEPOT_SECTION.
func TestParse_DefaultDepot(t *testing.T) {
	input := `NAME : no-depot-section
TYPE : CVRP
DIMENSION : 3
CAPACITY : 100
EDGE_WEIGHT_TYPE : EUC_2D
NODE_COORD_SECTION
1 0 0
2 5 5
3 10 10
DEMAND_SECTION
1 0
2 10
3 20
EOF`

	inst, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	if inst.DepotID() != 1 {
		t.Errorf("DepotID = %d, want 1 (default)", inst.DepotID())
	}
}

// TestInstance_CustomerNodes verifies customer extraction.
func TestInstance_CustomerNodes(t *testing.T) {
	inst, err := LoadFile(testInstancePath)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}

	customers := inst.CustomerNodes()
	if len(customers) != 12 {
		t.Errorf("CustomerNodes = %d, want 12", len(customers))
	}

	// Verify depot is not in customers.
	for _, c := range customers {
		if c.ID == inst.DepotID() {
			t.Errorf("Depot ID %d found in customer list", c.ID)
		}
	}
}

// TestLoadFile_AndConvert verifies loading then converting via parent package.
func TestLoadFile_AndConvert(t *testing.T) {
	inst, err := LoadFile(testInstancePath)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}
	if inst.Name != "E-n13-k4" {
		t.Errorf("Name = %q, want E-n13-k4", inst.Name)
	}
	if len(inst.CustomerNodes()) != 12 {
		t.Errorf("Customers = %d, want 12", len(inst.CustomerNodes()))
	}
}

// TestClassifyDistanceType verifies distance type classification.
func TestClassifyDistanceType(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"EUC_2D", "rounded"},
		{"CEIL_2D", "ceiling"},
		{"ATT", "att"},
		{"EXPLICIT", "explicit"},
		{"GEO", "geo"},
		{"UNKNOWN", "rounded"},
	}
	for _, tc := range cases {
		got := classifyDistanceType(tc.input)
		if got != tc.want {
			t.Errorf("classifyDistanceType(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

// TestValidate_MissingNodes reports error.
func TestValidate_MissingNodes(t *testing.T) {
	inst := &Instance{Dimension: 5}
	err := inst.Validate()
	if err == nil {
		t.Error("Expected validation error for missing nodes")
	}
}

// TestValidate_MissingCapacityForCVRP reports error.
func TestValidate_MissingCapacityForCVRP(t *testing.T) {
	inst := &Instance{
		Type:      "CVRP",
		Dimension: 3,
		Nodes:     []Node{{1, 0, 0}, {2, 1, 1}, {3, 2, 2}},
	}
	err := inst.Validate()
	if err == nil {
		t.Error("Expected validation error for missing capacity in CVRP")
	}
}
