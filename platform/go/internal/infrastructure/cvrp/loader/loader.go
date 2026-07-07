// Package loader provides CVRPLIB dataset parsing, independent of the optimiser.
//
// It reads standard TSPLIB-format VRP instance files and produces a structured
// Instance that can be converted into any VRP domain model. Designed to be
// extended for future VRP variants (VRPTW, MDVRP, OVRP, etc.) without
// modifying the core parser.
package loader

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// Instance represents a parsed CVRPLIB problem instance.
// This is a format-neutral intermediate representation — it contains
// everything from the file without domain-specific interpretation.
type Instance struct {
	// --- Metadata ---
	Name           string
	Comment        string
	Type           string // "CVRP", "TSP", "VRPTW", etc.
	Dimension      int    // total nodes (depot + customers)
	Capacity       int    // vehicle capacity (0 if not specified)
	Vehicles       int    // number of vehicles (0 = unlimited/unspecified)
	BestKnown      int    // best known solution value (0 if unknown)
	EdgeWeightType string // "EUC_2D", "CEIL_2D", "ATT", "EXPLICIT", etc.
	DistanceType   string // "rounded" (nint), "exact", "ceiling" — derived from EdgeWeightType

	// --- Node Data ---
	Nodes    []Node // all nodes including depot(s)
	Demands  []int  // demand[nodeID] — indexed by node ID (1-based in file, stored 0-based)
	DepotIDs []int  // node IDs that are depots

	// --- Optional: Explicit Distance Matrix ---
	// Present when EdgeWeightType is "EXPLICIT".
	ExplicitDistances [][]int

	// --- Optional: Time Windows (for VRPTW) ---
	TimeWindows []TimeWindow // empty for plain CVRP

	// --- Source ---
	FilePath string
}

// Node represents a location (depot or customer) with coordinates.
type Node struct {
	ID int
	X  float64
	Y  float64
}

// TimeWindow represents a service time window (for future VRPTW support).
type TimeWindow struct {
	NodeID    int
	ReadyTime int
	DueTime   int
	Service   int
}

// LoadFile parses a CVRPLIB instance file from the given path.
func LoadFile(path string) (*Instance, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	defer f.Close()

	inst, err := Parse(f)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	inst.FilePath = path
	return inst, nil
}

// Parse reads a CVRPLIB instance from any io.Reader.
// This allows testing without file I/O.
func Parse(r io.Reader) (*Instance, error) {
	inst := &Instance{}
	scanner := bufio.NewScanner(r)

	section := ""
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || line == "EOF" {
			continue
		}

		// Detect section headers (no colon, specific keywords).
		if isSectionHeader(line) {
			section = normaliseSectionName(line)
			continue
		}

		// Key-value header lines (contain ":").
		if strings.Contains(line, ":") && section == "" {
			key, val := parseHeaderLine(line)
			switch key {
			case "NAME":
				inst.Name = val
			case "COMMENT":
				inst.Comment = val
			case "TYPE":
				inst.Type = val
			case "DIMENSION":
				inst.Dimension, _ = strconv.Atoi(val)
			case "CAPACITY":
				inst.Capacity, _ = strconv.Atoi(val)
			case "VEHICLES":
				inst.Vehicles, _ = strconv.Atoi(val)
			case "BEST_KNOWN":
				inst.BestKnown, _ = strconv.Atoi(val)
			case "EDGE_WEIGHT_TYPE":
				inst.EdgeWeightType = val
				inst.DistanceType = classifyDistanceType(val)
			}
			continue
		}

		// Section data parsing.
		switch section {
		case "node_coord":
			if node, ok := parseNode(line); ok {
				inst.Nodes = append(inst.Nodes, node)
			}
		case "demand":
			if id, d, ok := parseDemandLine(line); ok {
				// Grow demand slice to accommodate node ID.
				for len(inst.Demands) <= id {
					inst.Demands = append(inst.Demands, 0)
				}
				inst.Demands[id] = d
			}
		case "depot":
			id, err := strconv.Atoi(strings.TrimSpace(line))
			if err != nil || id == -1 {
				section = "" // end of depot section
				continue
			}
			inst.DepotIDs = append(inst.DepotIDs, id)
		case "time_window":
			if tw, ok := parseTimeWindow(line); ok {
				inst.TimeWindows = append(inst.TimeWindows, tw)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading: %w", err)
	}

	// Validation.
	if inst.Dimension == 0 && len(inst.Nodes) > 0 {
		inst.Dimension = len(inst.Nodes)
	}
	if err := inst.Validate(); err != nil {
		return nil, err
	}

	return inst, nil
}

// Validate checks the instance for minimum required data.
func (inst *Instance) Validate() error {
	if inst.Dimension == 0 {
		return fmt.Errorf("invalid instance: DIMENSION not specified")
	}
	if len(inst.Nodes) == 0 {
		return fmt.Errorf("invalid instance: no node coordinates found")
	}
	if inst.Capacity == 0 && inst.Type == "CVRP" {
		return fmt.Errorf("invalid instance: CAPACITY not specified for CVRP")
	}
	return nil
}

// DepotID returns the primary depot node ID.
// Defaults to 1 if no depot section was specified.
func (inst *Instance) DepotID() int {
	if len(inst.DepotIDs) > 0 {
		return inst.DepotIDs[0]
	}
	return 1
}

// CustomerNodes returns all nodes except depot(s).
func (inst *Instance) CustomerNodes() []Node {
	depotSet := make(map[int]bool, len(inst.DepotIDs))
	for _, id := range inst.DepotIDs {
		depotSet[id] = true
	}
	if len(depotSet) == 0 {
		depotSet[1] = true // default depot
	}

	var customers []Node
	for _, n := range inst.Nodes {
		if !depotSet[n.ID] {
			customers = append(customers, n)
		}
	}
	return customers
}

// NodeDemand returns the demand for a given node ID.
// Returns 0 for depot or unknown nodes.
func (inst *Instance) NodeDemand(nodeID int) int {
	if nodeID < len(inst.Demands) {
		return inst.Demands[nodeID]
	}
	return 0
}

// --- Parsing helpers ---

func parseHeaderLine(line string) (string, string) {
	parts := strings.SplitN(line, ":", 2)
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
}

func isSectionHeader(line string) bool {
	upper := strings.ToUpper(line)
	switch upper {
	case "NODE_COORD_SECTION", "DEMAND_SECTION", "DEPOT_SECTION",
		"EDGE_WEIGHT_SECTION", "TIME_WINDOW_SECTION",
		"SVC_TIME_SECTION", "DISPLAY_DATA_SECTION":
		return true
	}
	return false
}

func normaliseSectionName(line string) string {
	upper := strings.ToUpper(line)
	switch upper {
	case "NODE_COORD_SECTION":
		return "node_coord"
	case "DEMAND_SECTION":
		return "demand"
	case "DEPOT_SECTION":
		return "depot"
	case "TIME_WINDOW_SECTION":
		return "time_window"
	default:
		return strings.ToLower(strings.TrimSuffix(upper, "_SECTION"))
	}
}

func parseNode(line string) (Node, bool) {
	fields := strings.Fields(line)
	if len(fields) < 3 {
		return Node{}, false
	}
	id, err := strconv.Atoi(fields[0])
	if err != nil {
		return Node{}, false
	}
	x, err := strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return Node{}, false
	}
	y, err := strconv.ParseFloat(fields[2], 64)
	if err != nil {
		return Node{}, false
	}
	return Node{ID: id, X: x, Y: y}, true
}

func parseDemandLine(line string) (int, int, bool) {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return 0, 0, false
	}
	id, err := strconv.Atoi(fields[0])
	if err != nil {
		return 0, 0, false
	}
	d, err := strconv.Atoi(fields[1])
	if err != nil {
		return 0, 0, false
	}
	return id, d, true
}

func parseTimeWindow(line string) (TimeWindow, bool) {
	fields := strings.Fields(line)
	if len(fields) < 4 {
		return TimeWindow{}, false
	}
	id, _ := strconv.Atoi(fields[0])
	ready, _ := strconv.Atoi(fields[1])
	due, _ := strconv.Atoi(fields[2])
	service, _ := strconv.Atoi(fields[3])
	return TimeWindow{NodeID: id, ReadyTime: ready, DueTime: due, Service: service}, true
}

func classifyDistanceType(ewt string) string {
	switch strings.ToUpper(ewt) {
	case "EUC_2D":
		return "rounded"
	case "CEIL_2D":
		return "ceiling"
	case "ATT":
		return "att"
	case "EXPLICIT":
		return "explicit"
	case "GEO":
		return "geo"
	default:
		return "rounded" // default to EUC_2D behaviour
	}
}
