package cvrp

import (
	"testing"
)

func TestLoadBenchmarkInstances(t *testing.T) {
	instances := []struct {
		path      string
		name      string
		customers int
		capacity  int
	}{
		{"../../../../../examples/cvrp/A-n32-k5.vrp", "A-n32-k5", 31, 100},
		{"../../../../../examples/cvrp/A-n45-k6.vrp", "A-n45-k6", 44, 100},
		{"../../../../../examples/cvrp/A-n60-k9.vrp", "A-n60-k9", 59, 100},
		{"../../../../../examples/cvrp/A-n80-k10.vrp", "A-n80-k10", 79, 100},
	}

	for _, inst := range instances {
		ds, err := LoadDataset(inst.path)
		if err != nil {
			t.Errorf("%s: LoadDataset failed: %v", inst.name, err)
			continue
		}
		if ds.Name != inst.name {
			t.Errorf("%s: Name = %q", inst.name, ds.Name)
		}
		if len(ds.Customers) != inst.customers {
			t.Errorf("%s: Customers = %d, want %d", inst.name, len(ds.Customers), inst.customers)
		}
		if ds.Capacity != inst.capacity {
			t.Errorf("%s: Capacity = %d, want %d", inst.name, ds.Capacity, inst.capacity)
		}

		// Verify constructive solution works.
		p := NewCVRPProblem(ds)
		sol, err := p.CreateInitialSolution()
		if err != nil {
			t.Errorf("%s: CreateInitialSolution failed: %v", inst.name, err)
			continue
		}
		cs := sol.(*cvrpSolution)
		feasible, violations := p.ValidateFull(cs)
		if !feasible {
			t.Errorf("%s: Initial solution infeasible: %+v", inst.name, violations)
		}
		cost := p.Evaluate(sol)
		t.Logf("%s: %d customers, constructive=%d, routes=%d", inst.name, inst.customers, cost, len(cs.routes))
	}
}
