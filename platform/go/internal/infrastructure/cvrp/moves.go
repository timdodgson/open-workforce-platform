package cvrp

import "fmt"

// --- CVRP Neighbourhood Framework ---
//
// Provides five move types for metaheuristic search:
//   1. Relocate — move one customer to a different position
//   2. InterSwap — swap two customers between different routes
//   3. IntraSwap — swap two customer positions within the same route
//   4. TwoOpt — reverse a segment within a route
//   5. OrOpt — move a chain of 1-3 consecutive customers to a new position
//
// Each move is self-describing for telemetry and supports Apply/Undo.

// MoveType identifies the neighbourhood operator used.
type MoveType int

const (
	// Relocate moves a single customer from one position to another.
	Relocate MoveType = iota
	// Swap exchanges two customers between different routes.
	Swap
	// IntraSwap exchanges two customers within the same route.
	IntraSwap
	// TwoOpt reverses a segment within a single route.
	TwoOpt
	// OrOpt moves a chain of 1-3 consecutive customers to a new position.
	OrOpt
)

// String returns a human-readable name for the move type.
func (mt MoveType) String() string {
	switch mt {
	case Relocate:
		return "relocate"
	case Swap:
		return "inter-swap"
	case IntraSwap:
		return "intra-swap"
	case TwoOpt:
		return "2-opt"
	case OrOpt:
		return "or-opt"
	default:
		return "unknown"
	}
}

// Move represents a neighbourhood move that can be applied or undone.
// It carries enough information for telemetry and debugging.
type Move struct {
	Type MoveType

	// For Relocate: customer at (FromRoute, FromPos) moves to (ToRoute, ToPos).
	// For Swap/IntraSwap: customer at (FromRoute, FromPos) swaps with (ToRoute, ToPos).
	// For TwoOpt: segment [FromPos, ToPos] in FromRoute is reversed. ToRoute unused.
	// For OrOpt: chain of ChainLen customers starting at FromPos in FromRoute
	//            moves to ToPos in ToRoute.
	FromRoute int
	FromPos   int
	ToRoute   int
	ToPos     int
	ChainLen  int // OrOpt: number of consecutive customers moved (1, 2, or 3)

	// Telemetry — populated on move generation for dashboard/debugging.
	CustomerA int // customer index involved (first)
	CustomerB int // customer index involved (second, -1 for relocate/2-opt/or-opt)
	DemandA   int // demand of customer A
	DemandB   int // demand of customer B (0 for relocate/2-opt/or-opt)
}

// Description returns a human-readable description of the move for telemetry.
func (m Move) Description() string {
	switch m.Type {
	case Relocate:
		return fmt.Sprintf("relocate customer %d from route %d pos %d → route %d pos %d",
			m.CustomerA, m.FromRoute, m.FromPos, m.ToRoute, m.ToPos)
	case Swap:
		return fmt.Sprintf("swap customer %d (route %d) ↔ customer %d (route %d)",
			m.CustomerA, m.FromRoute, m.CustomerB, m.ToRoute)
	case IntraSwap:
		return fmt.Sprintf("intra-swap pos %d ↔ pos %d in route %d",
			m.FromPos, m.ToPos, m.FromRoute)
	case TwoOpt:
		return fmt.Sprintf("2-opt reverse [%d:%d] in route %d",
			m.FromPos, m.ToPos, m.FromRoute)
	case OrOpt:
		return fmt.Sprintf("or-opt move %d customers from route %d pos %d → route %d pos %d",
			m.ChainLen, m.FromRoute, m.FromPos, m.ToRoute, m.ToPos)
	default:
		return "unknown move"
	}
}

// AffectedRoutes returns the route indices modified by this move.
func (m Move) AffectedRoutes() []int {
	switch m.Type {
	case Relocate, Swap, OrOpt:
		if m.FromRoute == m.ToRoute {
			return []int{m.FromRoute}
		}
		return []int{m.FromRoute, m.ToRoute}
	case IntraSwap, TwoOpt:
		return []int{m.FromRoute}
	default:
		return nil
	}
}

// AffectedCustomers returns the customer indices involved in this move.
func (m Move) AffectedCustomers() []int {
	if m.CustomerB >= 0 {
		return []int{m.CustomerA, m.CustomerB}
	}
	return []int{m.CustomerA}
}
