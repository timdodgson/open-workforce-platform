package cvrp

import "fmt"

// --- CVRP Neighbourhood Framework ---
//
// Provides four move types for metaheuristic search:
//   1. Relocate — move one customer to a different position
//   2. InterSwap — swap two customers between different routes
//   3. IntraSwap — swap two customer positions within the same route
//   4. TwoOpt — reverse a segment within a route
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
	FromRoute int
	FromPos   int
	ToRoute   int
	ToPos     int

	// Telemetry — populated on move generation for dashboard/debugging.
	CustomerA int // customer index involved (first)
	CustomerB int // customer index involved (second, -1 for relocate/2-opt)
	DemandA   int // demand of customer A
	DemandB   int // demand of customer B (0 for relocate/2-opt)
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
	default:
		return "unknown move"
	}
}

// AffectedRoutes returns the route indices modified by this move.
func (m Move) AffectedRoutes() []int {
	switch m.Type {
	case Relocate, Swap:
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
