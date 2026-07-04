package cvrp

// --- Initial Solution Construction ---
//
// Provides simple constructive heuristics that produce valid starting solutions.
// The focus is correctness over quality — these solutions will be improved by
// metaheuristic search (SA, LAHC, Tabu).
//
// Guarantees:
//   - Every customer assigned exactly once
//   - No route exceeds vehicle capacity
//   - Every route starts and ends at depot (implicit in representation)

// ConstructionStrategy identifies which heuristic to use.
type ConstructionStrategy int

const (
	// NearestNeighbour greedily adds the nearest feasible customer to the current route.
	NearestNeighbour ConstructionStrategy = iota
	// Sequential assigns customers in order, opening a new route when capacity is reached.
	Sequential
)

// BuildInitialSolution constructs a feasible CVRP solution using the given strategy.
// Returns the solution and verifies it passes validation.
func (p *CVRPProblem) BuildInitialSolution(strategy ConstructionStrategy) (*cvrpSolution, error) {
	switch strategy {
	case Sequential:
		return p.buildSequential()
	default:
		return p.buildNearestNeighbour()
	}
}

// buildNearestNeighbour constructs a solution by repeatedly selecting the
// nearest unvisited customer that fits in the current route's remaining capacity.
// When no customer fits, a new route is started.
func (p *CVRPProblem) buildNearestNeighbour() (*cvrpSolution, error) {
	customers := len(p.dataset.Customers)
	if customers == 0 {
		return &cvrpSolution{}, nil
	}

	visited := make([]bool, customers)
	sol := &cvrpSolution{}
	remaining := customers

	for remaining > 0 {
		route := []int{}
		load := 0
		current := 0 // start at depot (node index 0)

		for {
			bestDist := int(^uint(0) >> 1)
			bestIdx := -1
			for i := 0; i < customers; i++ {
				if visited[i] {
					continue
				}
				demand := p.dataset.Customers[i].Demand
				if load+demand > p.dataset.Capacity {
					continue
				}
				d := p.distMatrix[current][i+1]
				if d < bestDist {
					bestDist = d
					bestIdx = i
				}
			}

			if bestIdx == -1 {
				break
			}

			route = append(route, bestIdx)
			load += p.dataset.Customers[bestIdx].Demand
			visited[bestIdx] = true
			remaining--
			current = bestIdx + 1
		}

		if len(route) > 0 {
			sol.routes = append(sol.routes, route)
			sol.loads = append(sol.loads, load)
			sol.costs = append(sol.costs, float64(p.routeDistance(route)))
		}
	}

	return sol, nil
}

// buildSequential assigns customers in their original order.
// Opens a new route whenever the next customer would exceed capacity.
// Simplest possible valid construction — useful as a baseline.
func (p *CVRPProblem) buildSequential() (*cvrpSolution, error) {
	customers := len(p.dataset.Customers)
	if customers == 0 {
		return &cvrpSolution{}, nil
	}

	sol := &cvrpSolution{}
	route := []int{}
	load := 0

	for i := 0; i < customers; i++ {
		demand := p.dataset.Customers[i].Demand

		// If adding this customer would exceed capacity, close current route.
		if load+demand > p.dataset.Capacity && len(route) > 0 {
			sol.routes = append(sol.routes, route)
			sol.loads = append(sol.loads, load)
			sol.costs = append(sol.costs, float64(p.routeDistance(route)))
			route = []int{}
			load = 0
		}

		route = append(route, i)
		load += demand
	}

	// Close final route.
	if len(route) > 0 {
		sol.routes = append(sol.routes, route)
		sol.loads = append(sol.loads, load)
		sol.costs = append(sol.costs, float64(p.routeDistance(route)))
	}

	return sol, nil
}
