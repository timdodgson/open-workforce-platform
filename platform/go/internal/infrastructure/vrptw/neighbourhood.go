package vrptw

import (
	"math/rand"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

// --- Neighbourhood Operators ---
//
// Same five operators as CVRP but with additional time window feasibility checks.
// Moves that violate capacity are rejected immediately.
// Moves that violate time windows are penalised (not rejected) — this allows
// the search to traverse infeasible space temporarily, guided by the penalty.

// Move types.
type moveType int

const (
	relocate  moveType = iota
	swap
	intraSwap
	twoOpt
	orOpt
)

// vrptwMove stores move information for undo.
type vrptwMove struct {
	mType     moveType
	fromRoute int
	fromPos   int
	toRoute   int
	toPos     int
	chainLen  int // for or-opt
	custA     int
	custB     int
	demandA   int
	demandB   int
}

// Move type selection weights.
const (
	wRelocate  = 25
	wSwap      = 20
	wIntraSwap = 15
	wTwoOpt    = 15
	wOrOpt     = 25
)

func (p *VRPTWProblem) generateMove(sol *vrptwSolution, rng *rand.Rand) optimisation.MoveResult {
	numRoutes := len(sol.routes)
	if numRoutes == 0 {
		return optimisation.MoveResult{Valid: false}
	}

	total := wRelocate + wSwap + wIntraSwap + wTwoOpt + wOrOpt
	roll := rng.Intn(total)

	switch {
	case roll < wRelocate:
		return p.genRelocate(sol, rng)
	case roll < wRelocate+wSwap:
		return p.genSwap(sol, rng)
	case roll < wRelocate+wSwap+wIntraSwap:
		return p.genIntraSwap(sol, rng)
	case roll < wRelocate+wSwap+wIntraSwap+wTwoOpt:
		return p.genTwoOpt(sol, rng)
	default:
		return p.genOrOpt(sol, rng)
	}
}

// --- Relocate ---

func (p *VRPTWProblem) genRelocate(sol *vrptwSolution, rng *rand.Rand) optimisation.MoveResult {
	numRoutes := len(sol.routes)
	fromRoute := rng.Intn(numRoutes)
	if len(sol.routes[fromRoute]) == 0 {
		return optimisation.MoveResult{Valid: false}
	}

	fromPos := rng.Intn(len(sol.routes[fromRoute]))
	custIdx := sol.routes[fromRoute][fromPos]
	demand := p.dataset.Customers[custIdx].Demand

	toRoute := rng.Intn(numRoutes)

	// Capacity check for inter-route.
	if toRoute != fromRoute {
		if sol.loads[toRoute]+demand > p.dataset.Capacity {
			return optimisation.MoveResult{Valid: false}
		}
	}

	toLen := len(sol.routes[toRoute])
	if toRoute == fromRoute {
		toLen--
	}
	if toLen < 0 {
		toLen = 0
	}
	toPos := rng.Intn(toLen + 1)

	// Apply.
	sol.routes[fromRoute] = append(sol.routes[fromRoute][:fromPos], sol.routes[fromRoute][fromPos+1:]...)
	sol.loads[fromRoute] -= demand

	sol.routes[toRoute] = append(sol.routes[toRoute][:toPos], append([]int{custIdx}, sol.routes[toRoute][toPos:]...)...)
	sol.loads[toRoute] += demand

	return optimisation.MoveResult{
		Valid: true,
		Move: vrptwMove{
			mType:     relocate,
			fromRoute: fromRoute,
			fromPos:   fromPos,
			toRoute:   toRoute,
			toPos:     toPos,
			custA:     custIdx,
			demandA:   demand,
		},
	}
}

// --- Inter-route Swap ---

func (p *VRPTWProblem) genSwap(sol *vrptwSolution, rng *rand.Rand) optimisation.MoveResult {
	numRoutes := len(sol.routes)
	if numRoutes < 2 {
		return optimisation.MoveResult{Valid: false}
	}

	routeA := rng.Intn(numRoutes)
	if len(sol.routes[routeA]) == 0 {
		return optimisation.MoveResult{Valid: false}
	}

	routeB := rng.Intn(numRoutes - 1)
	if routeB >= routeA {
		routeB++
	}
	if len(sol.routes[routeB]) == 0 {
		return optimisation.MoveResult{Valid: false}
	}

	posA := rng.Intn(len(sol.routes[routeA]))
	posB := rng.Intn(len(sol.routes[routeB]))

	custA := sol.routes[routeA][posA]
	custB := sol.routes[routeB][posB]
	demandA := p.dataset.Customers[custA].Demand
	demandB := p.dataset.Customers[custB].Demand

	// Capacity check.
	newLoadA := sol.loads[routeA] - demandA + demandB
	newLoadB := sol.loads[routeB] - demandB + demandA
	if newLoadA > p.dataset.Capacity || newLoadB > p.dataset.Capacity {
		return optimisation.MoveResult{Valid: false}
	}

	// Apply.
	sol.routes[routeA][posA] = custB
	sol.routes[routeB][posB] = custA
	sol.loads[routeA] = newLoadA
	sol.loads[routeB] = newLoadB

	return optimisation.MoveResult{
		Valid: true,
		Move: vrptwMove{
			mType:     swap,
			fromRoute: routeA,
			fromPos:   posA,
			toRoute:   routeB,
			toPos:     posB,
			custA:     custA,
			custB:     custB,
			demandA:   demandA,
			demandB:   demandB,
		},
	}
}

// --- Intra-route Swap ---

func (p *VRPTWProblem) genIntraSwap(sol *vrptwSolution, rng *rand.Rand) optimisation.MoveResult {
	numRoutes := len(sol.routes)
	route := rng.Intn(numRoutes)
	routeLen := len(sol.routes[route])
	if routeLen < 2 {
		return optimisation.MoveResult{Valid: false}
	}

	posA := rng.Intn(routeLen)
	posB := rng.Intn(routeLen - 1)
	if posB >= posA {
		posB++
	}

	custA := sol.routes[route][posA]
	custB := sol.routes[route][posB]

	sol.routes[route][posA] = custB
	sol.routes[route][posB] = custA

	return optimisation.MoveResult{
		Valid: true,
		Move: vrptwMove{
			mType:     intraSwap,
			fromRoute: route,
			fromPos:   posA,
			toRoute:   route,
			toPos:     posB,
			custA:     custA,
			custB:     custB,
		},
	}
}

// --- 2-Opt ---

func (p *VRPTWProblem) genTwoOpt(sol *vrptwSolution, rng *rand.Rand) optimisation.MoveResult {
	numRoutes := len(sol.routes)
	route := rng.Intn(numRoutes)
	routeLen := len(sol.routes[route])
	if routeLen < 3 {
		return optimisation.MoveResult{Valid: false}
	}

	i := rng.Intn(routeLen - 1)
	j := i + 1 + rng.Intn(routeLen-i-1)

	// Reverse segment.
	for lo, hi := i, j; lo < hi; lo, hi = lo+1, hi-1 {
		sol.routes[route][lo], sol.routes[route][hi] = sol.routes[route][hi], sol.routes[route][lo]
	}

	return optimisation.MoveResult{
		Valid: true,
		Move: vrptwMove{
			mType:     twoOpt,
			fromRoute: route,
			fromPos:   i,
			toRoute:   route,
			toPos:     j,
		},
	}
}

// --- Or-opt ---

func (p *VRPTWProblem) genOrOpt(sol *vrptwSolution, rng *rand.Rand) optimisation.MoveResult {
	numRoutes := len(sol.routes)
	fromRoute := rng.Intn(numRoutes)
	fromLen := len(sol.routes[fromRoute])
	if fromLen < 2 {
		return optimisation.MoveResult{Valid: false}
	}

	maxChain := 3
	if fromLen < maxChain {
		maxChain = fromLen
	}
	chainLen := 1 + rng.Intn(maxChain)

	maxStart := fromLen - chainLen
	if maxStart < 0 {
		return optimisation.MoveResult{Valid: false}
	}
	fromPos := rng.Intn(maxStart + 1)

	chainDemand := 0
	for i := fromPos; i < fromPos+chainLen; i++ {
		chainDemand += p.dataset.Customers[sol.routes[fromRoute][i]].Demand
	}

	toRoute := rng.Intn(numRoutes)

	if toRoute != fromRoute {
		if sol.loads[toRoute]+chainDemand > p.dataset.Capacity {
			return optimisation.MoveResult{Valid: false}
		}
	}

	toLen := len(sol.routes[toRoute])
	if toRoute == fromRoute {
		toLen -= chainLen
	}
	if toLen < 0 {
		toLen = 0
	}
	toPos := rng.Intn(toLen + 1)

	// Extract chain.
	chain := make([]int, chainLen)
	copy(chain, sol.routes[fromRoute][fromPos:fromPos+chainLen])
	firstCust := chain[0]

	// Remove from source.
	sol.routes[fromRoute] = append(sol.routes[fromRoute][:fromPos], sol.routes[fromRoute][fromPos+chainLen:]...)
	sol.loads[fromRoute] -= chainDemand

	// Insert at destination.
	tail := make([]int, len(sol.routes[toRoute][toPos:]))
	copy(tail, sol.routes[toRoute][toPos:])
	sol.routes[toRoute] = append(sol.routes[toRoute][:toPos], chain...)
	sol.routes[toRoute] = append(sol.routes[toRoute], tail...)
	sol.loads[toRoute] += chainDemand

	return optimisation.MoveResult{
		Valid: true,
		Move: vrptwMove{
			mType:     orOpt,
			fromRoute: fromRoute,
			fromPos:   fromPos,
			toRoute:   toRoute,
			toPos:     toPos,
			chainLen:  chainLen,
			custA:     firstCust,
			demandA:   chainDemand,
		},
	}
}

// --- Undo ---

func (p *VRPTWProblem) undoMove(sol *vrptwSolution, mv vrptwMove) {
	switch mv.mType {
	case relocate:
		p.undoRelocate(sol, mv)
	case swap:
		p.undoSwap(sol, mv)
	case intraSwap:
		p.undoIntraSwap(sol, mv)
	case twoOpt:
		p.undoTwoOpt(sol, mv)
	case orOpt:
		p.undoOrOpt(sol, mv)
	}
}

func (p *VRPTWProblem) undoRelocate(sol *vrptwSolution, mv vrptwMove) {
	custIdx := sol.routes[mv.toRoute][mv.toPos]
	demand := p.dataset.Customers[custIdx].Demand

	sol.routes[mv.toRoute] = append(sol.routes[mv.toRoute][:mv.toPos], sol.routes[mv.toRoute][mv.toPos+1:]...)
	sol.loads[mv.toRoute] -= demand

	sol.routes[mv.fromRoute] = append(sol.routes[mv.fromRoute][:mv.fromPos], append([]int{custIdx}, sol.routes[mv.fromRoute][mv.fromPos:]...)...)
	sol.loads[mv.fromRoute] += demand
}

func (p *VRPTWProblem) undoSwap(sol *vrptwSolution, mv vrptwMove) {
	sol.routes[mv.fromRoute][mv.fromPos], sol.routes[mv.toRoute][mv.toPos] =
		sol.routes[mv.toRoute][mv.toPos], sol.routes[mv.fromRoute][mv.fromPos]
	sol.loads[mv.fromRoute] = sol.loads[mv.fromRoute] - mv.demandB + mv.demandA
	sol.loads[mv.toRoute] = sol.loads[mv.toRoute] - mv.demandA + mv.demandB
}

func (p *VRPTWProblem) undoIntraSwap(sol *vrptwSolution, mv vrptwMove) {
	sol.routes[mv.fromRoute][mv.fromPos], sol.routes[mv.fromRoute][mv.toPos] =
		sol.routes[mv.fromRoute][mv.toPos], sol.routes[mv.fromRoute][mv.fromPos]
}

func (p *VRPTWProblem) undoTwoOpt(sol *vrptwSolution, mv vrptwMove) {
	for lo, hi := mv.fromPos, mv.toPos; lo < hi; lo, hi = lo+1, hi-1 {
		sol.routes[mv.fromRoute][lo], sol.routes[mv.fromRoute][hi] =
			sol.routes[mv.fromRoute][hi], sol.routes[mv.fromRoute][lo]
	}
}

func (p *VRPTWProblem) undoOrOpt(sol *vrptwSolution, mv vrptwMove) {
	chainLen := mv.chainLen

	chain := make([]int, chainLen)
	copy(chain, sol.routes[mv.toRoute][mv.toPos:mv.toPos+chainLen])
	chainDemand := 0
	for _, c := range chain {
		chainDemand += p.dataset.Customers[c].Demand
	}

	sol.routes[mv.toRoute] = append(sol.routes[mv.toRoute][:mv.toPos], sol.routes[mv.toRoute][mv.toPos+chainLen:]...)
	sol.loads[mv.toRoute] -= chainDemand

	tail := make([]int, len(sol.routes[mv.fromRoute][mv.fromPos:]))
	copy(tail, sol.routes[mv.fromRoute][mv.fromPos:])
	sol.routes[mv.fromRoute] = append(sol.routes[mv.fromRoute][:mv.fromPos], chain...)
	sol.routes[mv.fromRoute] = append(sol.routes[mv.fromRoute], tail...)
	sol.loads[mv.fromRoute] += chainDemand
}
