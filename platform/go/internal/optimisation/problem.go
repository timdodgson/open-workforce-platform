// Package optimisation — Problem interface
//
// Problem defines the contract between optimisation algorithms and problem domains.
// Algorithms (SA, LAHC, Tabu, Portfolio, Beam Search) depend only on this interface.
// Each problem domain (NRP, VRP, etc.) provides its own implementation.
package optimisation

import "math/rand"

// Solution is an opaque handle to a problem-specific solution.
// Algorithms never inspect or modify it directly — they interact through the Problem interface.
type Solution interface{}

// Move is an opaque handle to a problem-specific move.
// Algorithms use it to undo rejected moves.
type Move interface{}

// MoveResult describes the outcome of TryMove.
type MoveResult struct {
	// Valid is true if the move passed hard constraint validation and was applied.
	Valid bool

	// Move is the applied move handle — pass to UndoMove if the move is rejected.
	// Nil if Valid is false.
	Move Move
}

// Problem defines the contract that any optimisation problem must satisfy.
//
// Implementations encapsulate all domain knowledge: solution representation,
// move generation, hard constraint validation, soft constraint evaluation,
// and serialisation. Algorithms know nothing about the problem domain.
//
// Performance contract: Evaluate and TryMove are called millions of times
// per second in concurrent workers. Implementations must be allocation-free
// in the hot path where possible. Each worker owns its own Problem instance.
type Problem interface {
	// --- Solution Lifecycle ---

	// CreateInitialSolution builds a valid starting solution for the problem.
	// Returns an error if no feasible solution can be constructed.
	CreateInitialSolution() (Solution, error)

	// CloneSolution creates a deep copy of the solution.
	// The clone must be fully independent — mutations to one must not affect the other.
	CloneSolution(s Solution) Solution

	// --- Evaluation ---

	// Evaluate returns the objective value (penalty/cost) of a solution.
	// Lower is better. Called millions of times — must be fast.
	Evaluate(s Solution) int

	// --- Neighbourhood ---

	// TryMove generates a random move, validates it against hard constraints,
	// and applies it to the solution if valid. Returns the result.
	//
	// If MoveResult.Valid is false, the solution is unchanged.
	// If MoveResult.Valid is true, the solution has been mutated.
	//
	// The rng parameter provides the random source for move generation.
	// Each worker has its own rng — no synchronisation needed.
	TryMove(s Solution, rng *rand.Rand) MoveResult

	// UndoMove reverts a previously applied move.
	// Called when the algorithm rejects the move after evaluation.
	// The move must be the most recently applied (stack discipline).
	UndoMove(s Solution, m Move)

	// --- Serialisation ---

	// SolutionFingerprint returns a short hash/string identifying the solution state.
	// Used for diversity measurement in beam search.
	SolutionFingerprint(s Solution) string

	// SerializeSolution converts the solution into a format suitable for
	// dashboard visualisation or file export. Returns JSON-serialisable data.
	SerializeSolution(s Solution) ([]byte, error)
}

// DatasetLoader defines how a problem loads its input data.
// Separate from Problem because loading happens once, before workers start.
type DatasetLoader interface {
	// LoadDataset reads problem instance data from the given path/identifier.
	// Returns a configured Problem ready for optimisation.
	LoadDataset(path string) (Problem, error)
}
