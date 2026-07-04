package cvrp

import (
	"fmt"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/cvrp/loader"
)

// LoadDataset reads a CVRPLIB instance file in TSPLIB format.
// This is a convenience wrapper that parses and converts in one step.
//
// For full metadata access (vehicles, best known, comments), use
// loader.LoadFile() directly, then call ConvertInstance().
func LoadDataset(path string) (*Dataset, error) {
	inst, err := loader.LoadFile(path)
	if err != nil {
		return nil, err
	}
	return ConvertInstance(inst)
}

// ConvertInstance transforms a parsed loader.Instance into the CVRP domain Dataset.
// This is the bridge between format-neutral parsing and the domain model.
func ConvertInstance(inst *loader.Instance) (*Dataset, error) {
	if inst.Capacity == 0 {
		return nil, fmt.Errorf("cannot convert to CVRP dataset: no capacity specified")
	}

	depotID := inst.DepotID()

	// Find depot node.
	var depotNode *loader.Node
	for i := range inst.Nodes {
		if inst.Nodes[i].ID == depotID {
			depotNode = &inst.Nodes[i]
			break
		}
	}
	if depotNode == nil {
		return nil, fmt.Errorf("depot node %d not found in coordinates", depotID)
	}

	// Build customer list.
	customers := inst.CustomerNodes()
	cvrpCustomers := make([]Customer, 0, len(customers))
	for _, n := range customers {
		cvrpCustomers = append(cvrpCustomers, Customer{
			ID:     n.ID,
			X:      n.X,
			Y:      n.Y,
			Demand: inst.NodeDemand(n.ID),
		})
	}

	ds := &Dataset{
		Name:      inst.Name,
		Dimension: inst.Dimension,
		Capacity:  inst.Capacity,
		Depot: Depot{
			ID: depotNode.ID,
			X:  depotNode.X,
			Y:  depotNode.Y,
		},
		Customers: cvrpCustomers,
	}

	return ds, nil
}
