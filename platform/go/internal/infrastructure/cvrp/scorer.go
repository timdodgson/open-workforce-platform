package cvrp

// --- CVRP Scoring and Validation ---
//
// Scoring: minimise total Euclidean travel distance.
// Validation: verify all hard constraints are satisfied.
//
// The scorer operates on the internal cvrpSolution representation
// through the CVRPProblem. It matches the generic Problem interface
// (Evaluate returns a single int objective value).

// Violation describes a single hard constraint violation.
type Violation struct {
	Code    string // "CAPACITY", "COVERAGE", "DUPLICATE", "EMPTY_ROUTE"
	Route   int    // route index (-1 if not route-specific)
	Detail  string // human-readable description
}

// ScoreResult contains the full scoring breakdown for a CVRP solution.
type ScoreResult struct {
	// Objective.
	TotalDistance int // sum of all route distances (depot → customers → depot)
	RouteCount   int // number of non-empty routes

	// Per-route breakdown.
	RouteDistances []int // distance for each route
	RouteLoads     []int // total demand served per route

	// Hard constraint status.
	Feasible   bool        // true if all hard constraints satisfied
	Violations []Violation // empty if feasible

	// Penalty-augmented objective (used by optimiser).
	// TotalDistance + penalty for any hard constraint violations.
	AugmentedCost int
}

// Score computes the full scoring breakdown for a solution.
// This is the detailed version — Evaluate() is the fast hot-path version.
func (p *CVRPProblem) Score(s *cvrpSolution) ScoreResult {
	result := ScoreResult{
		RouteCount: len(s.routes),
		Feasible:   true,
	}

	// Per-route distance and load.
	result.RouteDistances = make([]int, len(s.routes))
	result.RouteLoads = make([]int, len(s.routes))

	for i, route := range s.routes {
		dist := p.routeDistance(route)
		result.RouteDistances[i] = dist
		result.TotalDistance += dist
		result.RouteLoads[i] = p.routeLoad(route)
	}

	// Validate hard constraints.
	violations := p.Validate(s)
	result.Violations = violations
	result.Feasible = len(violations) == 0

	// Augmented cost = distance + penalty for violations.
	result.AugmentedCost = result.TotalDistance
	for _, v := range violations {
		switch v.Code {
		case "CAPACITY":
			// Penalty proportional to excess.
			routeIdx := v.Route
			if routeIdx >= 0 && routeIdx < len(result.RouteLoads) {
				excess := result.RouteLoads[routeIdx] - p.dataset.Capacity
				if excess > 0 {
					result.AugmentedCost += excess * 1000
				}
			}
		case "DUPLICATE":
			result.AugmentedCost += 10000 // heavy penalty per duplicate
		case "COVERAGE":
			result.AugmentedCost += 10000 // heavy penalty per missing customer
		}
	}

	return result
}

// Validate checks all hard constraints and returns any violations.
// Returns an empty slice if the solution is feasible.
func (p *CVRPProblem) Validate(s *cvrpSolution) []Violation {
	var violations []Violation

	numCustomers := len(p.dataset.Customers)

	// --- H1: Vehicle capacity ---
	for i, route := range s.routes {
		load := p.routeLoad(route)
		if load > p.dataset.Capacity {
			violations = append(violations, Violation{
				Code:   "CAPACITY",
				Route:  i,
				Detail: routeCapacityDetail(i, load, p.dataset.Capacity),
			})
		}
	}

	// --- H2: Each customer visited exactly once ---
	visitCount := make([]int, numCustomers)
	for _, route := range s.routes {
		for _, custIdx := range route {
			if custIdx >= 0 && custIdx < numCustomers {
				visitCount[custIdx]++
			}
		}
	}

	for custIdx, count := range visitCount {
		if count == 0 {
			violations = append(violations, Violation{
				Code:   "COVERAGE",
				Route:  -1,
				Detail: customerCoverageDetail(p.dataset.Customers[custIdx].ID),
			})
		} else if count > 1 {
			violations = append(violations, Violation{
				Code:   "DUPLICATE",
				Route:  -1,
				Detail: customerDuplicateDetail(p.dataset.Customers[custIdx].ID, count),
			})
		}
	}

	// --- H3: Routes start and end at depot ---
	// Implicitly satisfied by the route representation (routes are customer
	// index sequences; depot is always prepended/appended in distance calculation).
	// No explicit violation possible with this representation.

	return violations
}

// ValidateFull is a convenience that returns (feasible, violations).
func (p *CVRPProblem) ValidateFull(s *cvrpSolution) (bool, []Violation) {
	v := p.Validate(s)
	return len(v) == 0, v
}

// --- Distance calculation helpers ---

// TotalDistance computes the objective value (total travel distance) without penalty.
func (p *CVRPProblem) TotalDistance(s *cvrpSolution) int {
	total := 0
	for _, route := range s.routes {
		total += p.routeDistance(route)
	}
	return total
}

// RouteDistance computes the distance for a single route including depot legs.
func (p *CVRPProblem) RouteDistance(routeIdx int, s *cvrpSolution) int {
	if routeIdx < 0 || routeIdx >= len(s.routes) {
		return 0
	}
	return p.routeDistance(s.routes[routeIdx])
}

// --- Formatting helpers ---

func routeCapacityDetail(routeIdx, load, capacity int) string {
	return "Route " + itoa(routeIdx) + ": load " + itoa(load) + " exceeds capacity " + itoa(capacity) +
		" (excess " + itoa(load-capacity) + ")"
}

func customerCoverageDetail(customerID int) string {
	return "Customer " + itoa(customerID) + " not visited by any route"
}

func customerDuplicateDetail(customerID, count int) string {
	return "Customer " + itoa(customerID) + " visited " + itoa(count) + " times (expected 1)"
}

// itoa avoids importing strconv for simple int formatting.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	if n < 0 {
		return "-" + itoa(-n)
	}
	digits := ""
	for n > 0 {
		digits = string(rune('0'+n%10)) + digits
		n /= 10
	}
	return digits
}
