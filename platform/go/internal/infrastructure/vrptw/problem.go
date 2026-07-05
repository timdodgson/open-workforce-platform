package vrptw

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

// VRPTWProblem implements optimisation.Problem for the VRPTW.
type VRPTWProblem struct {
	dataset    *Dataset
	distMatrix [][]int     // rounded Euclidean distances [nodeIdx][nodeIdx]
	timeMatrix [][]float64 // exact travel times (distance = time in Solomon instances)
	numNodes   int         // depot + customers
}

// NewVRPTWProblem creates a VRPTW problem instance from a loaded dataset.
func NewVRPTWProblem(ds *Dataset) *VRPTWProblem {
	numNodes := 1 + len(ds.Customers)

	// Build distance and time matrices.
	// In Solomon instances, travel time == Euclidean distance.
	distMatrix := make([][]int, numNodes)
	timeMatrix := make([][]float64, numNodes)
	for i := range distMatrix {
		distMatrix[i] = make([]int, numNodes)
		timeMatrix[i] = make([]float64, numNodes)
	}

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
			t := DistanceExact(nodeX[i], nodeY[i], nodeX[j], nodeY[j])
			distMatrix[i][j] = d
			distMatrix[j][i] = d
			timeMatrix[i][j] = t
			timeMatrix[j][i] = t
		}
	}

	return &VRPTWProblem{
		dataset:    ds,
		distMatrix: distMatrix,
		timeMatrix: timeMatrix,
		numNodes:   numNodes,
	}
}

// --- Internal solution representation ---

type vrptwSolution struct {
	routes [][]int // each route: slice of customer indices (0-based into dataset.Customers)
	loads  []int   // current load per route
}

// --- optimisation.Problem implementation ---

func (p *VRPTWProblem) CreateInitialSolution() (optimisation.Solution, error) {
	sol := p.buildTimeOrdered()
	return sol, nil
}

func (p *VRPTWProblem) CloneSolution(s optimisation.Solution) optimisation.Solution {
	src := s.(*vrptwSolution)
	clone := &vrptwSolution{
		routes: make([][]int, len(src.routes)),
		loads:  make([]int, len(src.loads)),
	}
	for i := range src.routes {
		clone.routes[i] = make([]int, len(src.routes[i]))
		copy(clone.routes[i], src.routes[i])
	}
	copy(clone.loads, src.loads)
	return clone
}

func (p *VRPTWProblem) Evaluate(s optimisation.Solution) int {
	sol := s.(*vrptwSolution)
	total := 0
	for _, route := range sol.routes {
		total += p.routeDistance(route)
	}
	// Penalty for constraint violations.
	total += p.penaltyForViolations(sol)
	return total
}

func (p *VRPTWProblem) TryMove(s optimisation.Solution, rng *rand.Rand) optimisation.MoveResult {
	sol := s.(*vrptwSolution)
	return p.generateMove(sol, rng)
}

func (p *VRPTWProblem) UndoMove(s optimisation.Solution, m optimisation.Move) {
	sol := s.(*vrptwSolution)
	mv := m.(vrptwMove)
	p.undoMove(sol, mv)
}

func (p *VRPTWProblem) SolutionFingerprint(s optimisation.Solution) string {
	sol := s.(*vrptwSolution)
	h := md5.New()
	for _, route := range sol.routes {
		for _, c := range route {
			fmt.Fprintf(h, "%d,", c)
		}
		fmt.Fprint(h, "|")
	}
	return fmt.Sprintf("%x", h.Sum(nil)[:6])
}

func (p *VRPTWProblem) SerializeSolution(s optimisation.Solution) ([]byte, error) {
	sol := s.(*vrptwSolution)

	type routeJSON struct {
		Customers []int `json:"customers"`
		Load      int   `json:"load"`
		Distance  int   `json:"distance"`
		Feasible  bool  `json:"feasible"`
	}
	type solutionJSON struct {
		Routes       []routeJSON `json:"routes"`
		TotalCost    int         `json:"totalCost"`
		Vehicles     int         `json:"vehicles"`
		Feasible     bool        `json:"feasible"`
		TWViolations int         `json:"timeWindowViolations"`
	}

	out := solutionJSON{
		Vehicles: len(sol.routes),
		Feasible: true,
	}

	totalCost := 0
	twViolations := 0
	for i, route := range sol.routes {
		dist := p.routeDistance(route)
		totalCost += dist
		routeFeasible, violations := p.checkRouteFeasibility(route)
		if !routeFeasible {
			out.Feasible = false
		}
		twViolations += violations

		custIDs := make([]int, len(route))
		for j, idx := range route {
			custIDs[j] = p.dataset.Customers[idx].ID
		}
		out.Routes = append(out.Routes, routeJSON{
			Customers: custIDs,
			Load:      sol.loads[i],
			Distance:  dist,
			Feasible:  routeFeasible,
		})
	}
	out.TotalCost = totalCost
	out.TWViolations = twViolations

	return json.Marshal(out)
}

// --- Route Evaluation ---

// routeDistance computes the total travel distance for a route (depot → customers → depot).
func (p *VRPTWProblem) routeDistance(route []int) int {
	if len(route) == 0 {
		return 0
	}
	dist := p.distMatrix[0][route[0]+1]
	for i := 0; i < len(route)-1; i++ {
		dist += p.distMatrix[route[i]+1][route[i+1]+1]
	}
	dist += p.distMatrix[route[len(route)-1]+1][0]
	return dist
}

// checkRouteFeasibility checks capacity and time window constraints for a route.
// Returns (feasible, numberOfTWViolations).
func (p *VRPTWProblem) checkRouteFeasibility(route []int) (bool, int) {
	if len(route) == 0 {
		return true, 0
	}

	violations := 0
	currentTime := float64(p.dataset.Depot.ReadyTime)

	// Depart depot → first customer.
	currentTime += p.timeMatrix[0][route[0]+1]

	for i, custIdx := range route {
		c := p.dataset.Customers[custIdx]

		// Wait if arrived before ready time.
		if currentTime < float64(c.ReadyTime) {
			currentTime = float64(c.ReadyTime)
		}

		// Check if arrived after due date (violation).
		if currentTime > float64(c.DueDate) {
			violations++
		}

		// Service time.
		currentTime += float64(c.Service)

		// Travel to next customer or depot.
		if i < len(route)-1 {
			currentTime += p.timeMatrix[custIdx+1][route[i+1]+1]
		} else {
			// Return to depot.
			currentTime += p.timeMatrix[custIdx+1][0]
		}
	}

	// Check depot return time.
	if currentTime > float64(p.dataset.Depot.DueDate) {
		violations++
	}

	return violations == 0, violations
}

// penaltyForViolations computes the total penalty for all constraint violations.
func (p *VRPTWProblem) penaltyForViolations(sol *vrptwSolution) int {
	penalty := 0

	for i, route := range sol.routes {
		// Capacity violation.
		if sol.loads[i] > p.dataset.Capacity {
			excess := sol.loads[i] - p.dataset.Capacity
			penalty += excess * 1000
		}

		// Time window violations.
		_, twViolations := p.checkRouteFeasibility(route)
		penalty += twViolations * 500
	}

	return penalty
}

// IsFeasible returns true if the solution has no constraint violations.
func (p *VRPTWProblem) IsFeasible(s optimisation.Solution) bool {
	sol := s.(*vrptwSolution)
	for i, route := range sol.routes {
		if sol.loads[i] > p.dataset.Capacity {
			return false
		}
		feasible, _ := p.checkRouteFeasibility(route)
		if !feasible {
			return false
		}
	}
	return true
}

// TotalDistance returns the pure distance without penalties.
func (p *VRPTWProblem) TotalDistance(s optimisation.Solution) int {
	sol := s.(*vrptwSolution)
	total := 0
	for _, route := range sol.routes {
		total += p.routeDistance(route)
	}
	return total
}

// RouteCount returns the number of routes (vehicles used).
func (p *VRPTWProblem) RouteCount(s optimisation.Solution) int {
	sol := s.(*vrptwSolution)
	count := 0
	for _, route := range sol.routes {
		if len(route) > 0 {
			count++
		}
	}
	return count
}

// --- Constructive Heuristic ---

// buildTimeOrdered constructs an initial solution using a time-window-aware
// insertion heuristic. Customers are sorted by ready time and inserted into
// routes respecting both capacity and time window constraints.
func (p *VRPTWProblem) buildTimeOrdered() *vrptwSolution {
	n := len(p.dataset.Customers)
	if n == 0 {
		return &vrptwSolution{}
	}

	// Sort customer indices by ready time (ascending).
	indices := make([]int, n)
	for i := range indices {
		indices[i] = i
	}
	// Simple insertion sort (fine for <1000 customers).
	for i := 1; i < n; i++ {
		for j := i; j > 0 && p.dataset.Customers[indices[j]].ReadyTime < p.dataset.Customers[indices[j-1]].ReadyTime; j-- {
			indices[j], indices[j-1] = indices[j-1], indices[j]
		}
	}

	sol := &vrptwSolution{}
	assigned := make([]bool, n)

	for {
		// Start a new route.
		route := []int{}
		load := 0
		currentTime := float64(p.dataset.Depot.ReadyTime)
		currentNode := 0 // depot

		for {
			bestIdx := -1
			bestCost := math.MaxFloat64

			for _, idx := range indices {
				if assigned[idx] {
					continue
				}
				c := p.dataset.Customers[idx]

				// Capacity check.
				if load+c.Demand > p.dataset.Capacity {
					continue
				}

				// Time feasibility check.
				travelTime := p.timeMatrix[currentNode][idx+1]
				arrivalTime := currentTime + travelTime

				// Wait if early.
				serviceStart := arrivalTime
				if serviceStart < float64(c.ReadyTime) {
					serviceStart = float64(c.ReadyTime)
				}

				// Check if too late.
				if serviceStart > float64(c.DueDate) {
					continue
				}

				// Check if we can return to depot after serving this customer.
				finishTime := serviceStart + float64(c.Service)
				returnTime := finishTime + p.timeMatrix[idx+1][0]
				if returnTime > float64(p.dataset.Depot.DueDate) {
					continue
				}

				// Use distance as insertion cost.
				cost := travelTime
				if cost < bestCost {
					bestCost = cost
					bestIdx = idx
				}
			}

			if bestIdx < 0 {
				break // no more customers fit this route
			}

			// Insert customer.
			c := p.dataset.Customers[bestIdx]
			travelTime := p.timeMatrix[currentNode][bestIdx+1]
			arrivalTime := currentTime + travelTime
			serviceStart := arrivalTime
			if serviceStart < float64(c.ReadyTime) {
				serviceStart = float64(c.ReadyTime)
			}

			route = append(route, bestIdx)
			load += c.Demand
			currentTime = serviceStart + float64(c.Service)
			currentNode = bestIdx + 1
			assigned[bestIdx] = true
		}

		if len(route) == 0 {
			break // all customers assigned or no feasible insertion
		}

		sol.routes = append(sol.routes, route)
		sol.loads = append(sol.loads, load)
	}

	// Check if any customers remain unassigned (shouldn't happen with correct instances).
	for i, a := range assigned {
		if !a {
			// Force-assign to a new single-customer route.
			sol.routes = append(sol.routes, []int{i})
			sol.loads = append(sol.loads, p.dataset.Customers[i].Demand)
		}
	}

	return sol
}

// --- Accessors ---

func (p *VRPTWProblem) Dataset() *Dataset { return p.dataset }

// --- Compile-time interface check ---
var _ optimisation.Problem = (*VRPTWProblem)(nil)
