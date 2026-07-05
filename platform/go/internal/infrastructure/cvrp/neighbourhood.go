package cvrp

import (
	"math/rand"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

// --- Neighbourhood Generation ---
//
// Randomly generates moves from the four neighbourhood operators.
// Each move validates capacity constraints before applying.
// Infeasible moves (capacity violation) are rejected — the generator
// returns Valid=false and the solution is unchanged.

// Move type selection weights (relative probabilities).
const (
	weightRelocate  = 25 // 25%
	weightSwap      = 20 // 20%
	weightIntraSwap = 15 // 15%
	weightTwoOpt    = 15 // 15%
	weightOrOpt     = 25 // 25% — highest value CVRP neighbourhood
)

// GenerateMove randomly selects and applies a neighbourhood move.
// Returns a MoveResult with Valid=true if the move was applied.
// The solution is only mutated if Valid is true.
func (p *CVRPProblem) GenerateMove(sol *cvrpSolution, rng *rand.Rand) optimisation.MoveResult {
	numRoutes := len(sol.routes)
	if numRoutes == 0 {
		return optimisation.MoveResult{Valid: false}
	}

	// Weighted random selection of move type.
	total := weightRelocate + weightSwap + weightIntraSwap + weightTwoOpt + weightOrOpt
	roll := rng.Intn(total)

	switch {
	case roll < weightRelocate:
		return p.generateRelocate(sol, rng)
	case roll < weightRelocate+weightSwap:
		return p.generateInterSwap(sol, rng)
	case roll < weightRelocate+weightSwap+weightIntraSwap:
		return p.generateIntraSwap(sol, rng)
	case roll < weightRelocate+weightSwap+weightIntraSwap+weightTwoOpt:
		return p.generateTwoOpt(sol, rng)
	default:
		return p.generateOrOpt(sol, rng)
	}
}

// --- Relocate ---

func (p *CVRPProblem) generateRelocate(sol *cvrpSolution, rng *rand.Rand) optimisation.MoveResult {
	numRoutes := len(sol.routes)

	fromRoute := rng.Intn(numRoutes)
	if len(sol.routes[fromRoute]) == 0 {
		return optimisation.MoveResult{Valid: false}
	}
	fromPos := rng.Intn(len(sol.routes[fromRoute]))
	custIdx := sol.routes[fromRoute][fromPos]
	demand := p.dataset.Customers[custIdx].Demand

	toRoute := rng.Intn(numRoutes)

	// Capacity check for inter-route relocate.
	if toRoute != fromRoute {
		if sol.loads[toRoute]+demand > p.dataset.Capacity {
			return optimisation.MoveResult{Valid: false}
		}
	}

	// Determine insertion position.
	toLen := len(sol.routes[toRoute])
	if toRoute == fromRoute {
		toLen--
	}
	if toLen < 0 {
		toLen = 0
	}
	toPos := rng.Intn(toLen + 1)

	// Apply: remove from source.
	sol.routes[fromRoute] = append(sol.routes[fromRoute][:fromPos], sol.routes[fromRoute][fromPos+1:]...)
	sol.loads[fromRoute] -= demand

	// Insert at destination.
	// toPos was generated for the post-removal array size (toLen accounts for removal).
	// No adjustment needed — toPos is already correct for the current array state.
	sol.routes[toRoute] = append(sol.routes[toRoute][:toPos], append([]int{custIdx}, sol.routes[toRoute][toPos:]...)...)
	sol.loads[toRoute] += demand

	return optimisation.MoveResult{
		Valid: true,
		Move: Move{
			Type:      Relocate,
			FromRoute: fromRoute,
			FromPos:   fromPos,
			ToRoute:   toRoute,
			ToPos:     toPos,
			CustomerA: custIdx,
			CustomerB: -1,
			DemandA:   demand,
		},
	}
}

// --- Inter-route Swap ---

func (p *CVRPProblem) generateInterSwap(sol *cvrpSolution, rng *rand.Rand) optimisation.MoveResult {
	numRoutes := len(sol.routes)
	if numRoutes < 2 {
		return optimisation.MoveResult{Valid: false}
	}

	routeA := rng.Intn(numRoutes)
	if len(sol.routes[routeA]) == 0 {
		return optimisation.MoveResult{Valid: false}
	}

	// Pick a different route.
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

	// Apply swap.
	sol.routes[routeA][posA] = custB
	sol.routes[routeB][posB] = custA
	sol.loads[routeA] = newLoadA
	sol.loads[routeB] = newLoadB

	return optimisation.MoveResult{
		Valid: true,
		Move: Move{
			Type:      Swap,
			FromRoute: routeA,
			FromPos:   posA,
			ToRoute:   routeB,
			ToPos:     posB,
			CustomerA: custA,
			CustomerB: custB,
			DemandA:   demandA,
			DemandB:   demandB,
		},
	}
}

// --- Intra-route Swap ---

func (p *CVRPProblem) generateIntraSwap(sol *cvrpSolution, rng *rand.Rand) optimisation.MoveResult {
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

	// Apply swap (no capacity change — same route).
	sol.routes[route][posA] = custB
	sol.routes[route][posB] = custA

	return optimisation.MoveResult{
		Valid: true,
		Move: Move{
			Type:      IntraSwap,
			FromRoute: route,
			FromPos:   posA,
			ToRoute:   route,
			ToPos:     posB,
			CustomerA: custA,
			CustomerB: custB,
			DemandA:   p.dataset.Customers[custA].Demand,
			DemandB:   p.dataset.Customers[custB].Demand,
		},
	}
}

// --- 2-Opt (intra-route segment reversal) ---

func (p *CVRPProblem) generateTwoOpt(sol *cvrpSolution, rng *rand.Rand) optimisation.MoveResult {
	numRoutes := len(sol.routes)
	route := rng.Intn(numRoutes)
	routeLen := len(sol.routes[route])
	if routeLen < 3 {
		return optimisation.MoveResult{Valid: false}
	}

	// Pick two positions to define the reversal segment [i, j].
	i := rng.Intn(routeLen - 1)
	j := i + 1 + rng.Intn(routeLen-i-1)

	// Reverse the segment in-place.
	reverseSlice(sol.routes[route], i, j)

	return optimisation.MoveResult{
		Valid: true,
		Move: Move{
			Type:      TwoOpt,
			FromRoute: route,
			FromPos:   i,
			ToRoute:   route,
			ToPos:     j,
			CustomerA: sol.routes[route][i], // first customer after reversal
			CustomerB: -1,
			DemandA:   0, // no demand change
		},
	}
}

// --- Or-opt (move a chain of 1-3 consecutive customers) ---

func (p *CVRPProblem) generateOrOpt(sol *cvrpSolution, rng *rand.Rand) optimisation.MoveResult {
	numRoutes := len(sol.routes)

	// Pick source route with enough customers for a chain.
	fromRoute := rng.Intn(numRoutes)
	fromLen := len(sol.routes[fromRoute])
	if fromLen < 2 {
		return optimisation.MoveResult{Valid: false}
	}

	// Pick chain length: 1, 2, or 3 (biased toward shorter for more valid moves).
	maxChain := 3
	if fromLen < maxChain {
		maxChain = fromLen
	}
	chainLen := 1 + rng.Intn(maxChain) // 1, 2, or 3

	// Pick start position of the chain.
	maxStart := fromLen - chainLen
	if maxStart < 0 {
		return optimisation.MoveResult{Valid: false}
	}
	fromPos := rng.Intn(maxStart + 1)

	// Calculate chain demand.
	chainDemand := 0
	for i := fromPos; i < fromPos+chainLen; i++ {
		chainDemand += p.dataset.Customers[sol.routes[fromRoute][i]].Demand
	}

	// Pick destination route.
	toRoute := rng.Intn(numRoutes)

	// Capacity check for inter-route moves.
	if toRoute != fromRoute {
		if sol.loads[toRoute]+chainDemand > p.dataset.Capacity {
			return optimisation.MoveResult{Valid: false}
		}
	}

	// Pick insertion position in destination.
	toLen := len(sol.routes[toRoute])
	if toRoute == fromRoute {
		toLen -= chainLen // account for removal
	}
	if toLen < 0 {
		toLen = 0
	}
	toPos := rng.Intn(toLen + 1)

	// Extract the chain.
	chain := make([]int, chainLen)
	copy(chain, sol.routes[fromRoute][fromPos:fromPos+chainLen])
	firstCustomer := chain[0]

	// Remove chain from source.
	sol.routes[fromRoute] = append(sol.routes[fromRoute][:fromPos], sol.routes[fromRoute][fromPos+chainLen:]...)
	sol.loads[fromRoute] -= chainDemand

	// Insert chain at destination.
	insertPos := toPos
	// No adjustment needed: toPos was calculated for post-removal array size.
	tail := make([]int, len(sol.routes[toRoute][insertPos:]))
	copy(tail, sol.routes[toRoute][insertPos:])
	sol.routes[toRoute] = append(sol.routes[toRoute][:insertPos], chain...)
	sol.routes[toRoute] = append(sol.routes[toRoute], tail...)
	sol.loads[toRoute] += chainDemand

	return optimisation.MoveResult{
		Valid: true,
		Move: Move{
			Type:      OrOpt,
			FromRoute: fromRoute,
			FromPos:   fromPos,
			ToRoute:   toRoute,
			ToPos:     insertPos,
			ChainLen:  chainLen,
			CustomerA: firstCustomer,
			CustomerB: -1,
			DemandA:   chainDemand,
		},
	}
}

// reverseSlice reverses elements in slice[i:j+1] in-place.
func reverseSlice(s []int, i, j int) {
	for i < j {
		s[i], s[j] = s[j], s[i]
		i++
		j--
	}
}

// --- Undo Operations ---

// UndoMoveOnSolution reverts a previously applied move.
func (p *CVRPProblem) UndoMoveOnSolution(sol *cvrpSolution, mv Move) {
	switch mv.Type {
	case Relocate:
		p.undoRelocate(sol, mv)
	case Swap:
		p.undoInterSwap(sol, mv)
	case IntraSwap:
		p.undoIntraSwap(sol, mv)
	case TwoOpt:
		p.undoTwoOpt(sol, mv)
	case OrOpt:
		p.undoOrOpt(sol, mv)
	}
}

func (p *CVRPProblem) undoRelocate(sol *cvrpSolution, mv Move) {
	// Customer is currently at mv.ToRoute[mv.ToPos].
	custIdx := sol.routes[mv.ToRoute][mv.ToPos]
	demand := p.dataset.Customers[custIdx].Demand

	// Remove from destination.
	sol.routes[mv.ToRoute] = append(sol.routes[mv.ToRoute][:mv.ToPos], sol.routes[mv.ToRoute][mv.ToPos+1:]...)
	sol.loads[mv.ToRoute] -= demand

	// Re-insert at original position.
	// Forward move: removed from FromPos, then inserted at ToPos (in post-removal array).
	// Undo: we just removed from ToPos. For same-route, the array is now back to
	// the post-first-removal state. Insert at FromPos restores original.
	// For inter-route: routes are independent, insert at FromPos directly.
	sol.routes[mv.FromRoute] = append(sol.routes[mv.FromRoute][:mv.FromPos], append([]int{custIdx}, sol.routes[mv.FromRoute][mv.FromPos:]...)...)
	sol.loads[mv.FromRoute] += demand
}

func (p *CVRPProblem) undoInterSwap(sol *cvrpSolution, mv Move) {
	// Swap is self-inverse.
	sol.routes[mv.FromRoute][mv.FromPos], sol.routes[mv.ToRoute][mv.ToPos] =
		sol.routes[mv.ToRoute][mv.ToPos], sol.routes[mv.FromRoute][mv.FromPos]
	sol.loads[mv.FromRoute] = p.routeLoad(sol.routes[mv.FromRoute])
	sol.loads[mv.ToRoute] = p.routeLoad(sol.routes[mv.ToRoute])
}

func (p *CVRPProblem) undoIntraSwap(sol *cvrpSolution, mv Move) {
	// Swap is self-inverse within same route.
	sol.routes[mv.FromRoute][mv.FromPos], sol.routes[mv.FromRoute][mv.ToPos] =
		sol.routes[mv.FromRoute][mv.ToPos], sol.routes[mv.FromRoute][mv.FromPos]
}

func (p *CVRPProblem) undoTwoOpt(sol *cvrpSolution, mv Move) {
	// Reverse is self-inverse.
	reverseSlice(sol.routes[mv.FromRoute], mv.FromPos, mv.ToPos)
}

func (p *CVRPProblem) undoOrOpt(sol *cvrpSolution, mv Move) {
	// Reverse of Or-opt: remove chain from ToRoute/ToPos, insert back at FromRoute/FromPos.
	chainLen := mv.ChainLen

	// Extract chain from current position.
	chain := make([]int, chainLen)
	copy(chain, sol.routes[mv.ToRoute][mv.ToPos:mv.ToPos+chainLen])
	chainDemand := 0
	for _, c := range chain {
		chainDemand += p.dataset.Customers[c].Demand
	}

	// Remove from destination.
	sol.routes[mv.ToRoute] = append(sol.routes[mv.ToRoute][:mv.ToPos], sol.routes[mv.ToRoute][mv.ToPos+chainLen:]...)
	sol.loads[mv.ToRoute] -= chainDemand

	// Re-insert at original position.
	insertPos := mv.FromPos
	tail := make([]int, len(sol.routes[mv.FromRoute][insertPos:]))
	copy(tail, sol.routes[mv.FromRoute][insertPos:])
	sol.routes[mv.FromRoute] = append(sol.routes[mv.FromRoute][:insertPos], chain...)
	sol.routes[mv.FromRoute] = append(sol.routes[mv.FromRoute], tail...)
	sol.loads[mv.FromRoute] += chainDemand
}
