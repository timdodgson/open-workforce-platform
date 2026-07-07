package cvrp

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

// CVRPProblem implements optimisation.Problem for the Capacitated Vehicle Routing Problem.
//
// Each worker owns its own CVRPProblem instance. The struct contains the
// pre-computed distance matrix and all instance data needed for evaluation.
type CVRPProblem struct {
	dataset    *Dataset
	distMatrix [][]int // rounded Euclidean distances [nodeIdx][nodeIdx]
	numNodes   int     // depot + customers
}

// NewCVRPProblem creates a CVRP problem instance from a loaded dataset.
func NewCVRPProblem(ds *Dataset) *CVRPProblem {
	numNodes := 1 + len(ds.Customers) // depot + customers
	// Build distance matrix with depot at index 0, customers at index 1..n.
	distMatrix := make([][]int, numNodes)
	for i := range distMatrix {
		distMatrix[i] = make([]int, numNodes)
	}

	// Node coordinates: index 0 = depot, 1..n = customers in order.
	nodeX := make([]float64, numNodes)
	nodeY := make([]float64, numNodes)
	nodeX[0] = ds.Depot.X
	nodeY[0] = ds.Depot.Y
	for i, c := range ds.Customers {
		nodeX[i+1] = c.X
		nodeY[i+1] = c.Y
	}

	for i := 0; i < numNodes; i++ {
		for j := i + 1; j < numNodes; j++ {
			d := DistanceRounded(nodeX[i], nodeY[i], nodeX[j], nodeY[j])
			distMatrix[i][j] = d
			distMatrix[j][i] = d
		}
	}

	return &CVRPProblem{
		dataset:    ds,
		distMatrix: distMatrix,
		numNodes:   numNodes,
	}
}

// --- Solution representation for the optimiser ---

// cvrpSolution is the mutable solution representation used by the optimiser.
// Routes are stored as slices of customer indices (0-based into dataset.Customers).
type cvrpSolution struct {
	routes [][]int   // each route is a slice of customer indices (0-based)
	loads  []int     // current load for each route
	costs  []float64 // distance cost per route (float for incremental updates)
}

// --- optimisation.Problem implementation ---

// CreateInitialSolution builds a feasible initial solution using nearest-neighbour heuristic.
func (p *CVRPProblem) CreateInitialSolution() (optimisation.Solution, error) {
	sol, err := p.BuildInitialSolution(NearestNeighbour)
	if err != nil {
		return nil, err
	}
	return sol, nil
}

// CloneSolution creates a deep copy of the solution.
func (p *CVRPProblem) CloneSolution(s optimisation.Solution) optimisation.Solution {
	src := s.(*cvrpSolution)
	clone := &cvrpSolution{
		routes: make([][]int, len(src.routes)),
		loads:  make([]int, len(src.loads)),
		costs:  make([]float64, len(src.costs)),
	}
	for i := range src.routes {
		clone.routes[i] = make([]int, len(src.routes[i]))
		copy(clone.routes[i], src.routes[i])
	}
	copy(clone.loads, src.loads)
	copy(clone.costs, src.costs)
	return clone
}

// Evaluate returns the total route distance (lower is better).
func (p *CVRPProblem) Evaluate(s optimisation.Solution) int {
	sol := s.(*cvrpSolution)
	total := 0
	for _, route := range sol.routes {
		total += p.routeDistance(route)
	}
	// Add penalty for capacity violations (to allow infeasible intermediate states).
	for i, load := range sol.loads {
		if load > p.dataset.Capacity {
			excess := load - p.dataset.Capacity
			total += excess * 1000 // heavy penalty per unit excess
			_ = i
		}
	}
	return total
}

// TryMove generates a random neighbourhood move, validates feasibility,
// and applies it. Delegates to the neighbourhood framework.
func (p *CVRPProblem) TryMove(s optimisation.Solution, rng *rand.Rand) optimisation.MoveResult {
	sol := s.(*cvrpSolution)
	return p.GenerateMove(sol, rng)
}

// UndoMove reverts a previously applied move.
func (p *CVRPProblem) UndoMove(s optimisation.Solution, m optimisation.Move) {
	sol := s.(*cvrpSolution)
	mv := m.(Move)
	p.UndoMoveOnSolution(sol, mv)
}

// SolutionFingerprint returns a hash of the route structure.
func (p *CVRPProblem) SolutionFingerprint(s optimisation.Solution) string {
	sol := s.(*cvrpSolution)
	h := md5.New()
	for _, route := range sol.routes {
		for _, c := range route {
			fmt.Fprintf(h, "%d,", c)
		}
		fmt.Fprint(h, "|")
	}
	return fmt.Sprintf("%x", h.Sum(nil)[:6])
}

// SerializeSolution converts to JSON for dashboard/export.
func (p *CVRPProblem) SerializeSolution(s optimisation.Solution) ([]byte, error) {
	sol := s.(*cvrpSolution)

	type routeJSON struct {
		Customers []int `json:"customers"`
		Load      int   `json:"load"`
		Distance  int   `json:"distance"`
	}
	type solutionJSON struct {
		Routes    []routeJSON `json:"routes"`
		TotalCost int         `json:"totalCost"`
		Vehicles  int         `json:"vehicles"`
		Feasible  bool        `json:"feasible"`
	}

	out := solutionJSON{
		Vehicles: len(sol.routes),
		Feasible: true,
	}

	totalCost := 0
	for i, route := range sol.routes {
		dist := p.routeDistance(route)
		totalCost += dist
		feasible := sol.loads[i] <= p.dataset.Capacity
		if !feasible {
			out.Feasible = false
		}
		// Convert customer indices back to IDs.
		custIDs := make([]int, len(route))
		for j, idx := range route {
			custIDs[j] = p.dataset.Customers[idx].ID
		}
		out.Routes = append(out.Routes, routeJSON{
			Customers: custIDs,
			Load:      sol.loads[i],
			Distance:  dist,
		})
	}
	out.TotalCost = totalCost

	return json.Marshal(out)
}

// --- Helpers ---

// routeDistance computes the total distance for a route (depot → customers → depot).
func (p *CVRPProblem) routeDistance(route []int) int {
	if len(route) == 0 {
		return 0
	}
	dist := p.distMatrix[0][route[0]+1] // depot to first customer
	for i := 0; i < len(route)-1; i++ {
		dist += p.distMatrix[route[i]+1][route[i+1]+1]
	}
	dist += p.distMatrix[route[len(route)-1]+1][0] // last customer to depot
	return dist
}

// routeLoad computes the total demand for a route.
func (p *CVRPProblem) routeLoad(route []int) int {
	load := 0
	for _, idx := range route {
		load += p.dataset.Customers[idx].Demand
	}
	return load
}

// --- Accessors ---

// Dataset returns the loaded problem instance.
func (p *CVRPProblem) Dataset() *Dataset { return p.dataset }

// --- Compile-time interface check ---

var _ optimisation.Problem = (*CVRPProblem)(nil)
