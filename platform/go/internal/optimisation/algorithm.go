package optimisation

import (
	"fmt"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/plan"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/workitem"
)

// Algorithm represents an optimisation strategy.
//
// Each algorithm exposes its name and a solve operation that accepts
// an OptimisationContext and returns an Optimised Plan.
type Algorithm interface {
	Name() string
	Solve(ctx OptimisationContext) (plan.OptimisedPlan, error)
}

// registry holds registered algorithms by name.
var registry = map[string]Algorithm{}

// register adds an algorithm to the registry.
func register(a Algorithm) {
	registry[a.Name()] = a
}

// Get returns the algorithm registered under the given name.
//
// Returns an error if no algorithm is found.
func Get(name string) (Algorithm, error) {
	a, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("unknown algorithm: %q", name)
	}
	return a, nil
}

// Available returns the names of all registered algorithms.
func Available() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	return names
}

// Solve is a convenience function that runs the constructive algorithm.
func Solve(items []workitem.WorkItem, capacities []ResourceInput, priorities []WorkItemInput) (plan.OptimisedPlan, error) {
	a, err := Get("constructive")
	if err != nil {
		return plan.OptimisedPlan{}, err
	}
	return a.Solve(NewContext(items, capacities, priorities))
}

// SolveHillClimbing is a convenience function that runs the hill-climbing algorithm.
func SolveHillClimbing(items []workitem.WorkItem, capacities []ResourceInput, priorities []WorkItemInput) (plan.OptimisedPlan, error) {
	a, err := Get("hill-climbing")
	if err != nil {
		return plan.OptimisedPlan{}, err
	}
	return a.Solve(NewContext(items, capacities, priorities))
}
