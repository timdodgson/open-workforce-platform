// Package searchdef holds the stable search engine contract for BYOD domains.
package searchdef

import "math/rand"

type Solution interface{}
type Move interface{}

type MoveResult struct {
	Valid bool
	Move  Move
}

type Problem interface {
	CreateInitialSolution() (Solution, error)
	CloneSolution(s Solution) Solution
	Evaluate(s Solution) int
	TryMove(s Solution, rng *rand.Rand) MoveResult
	UndoMove(s Solution, m Move)
	SolutionFingerprint(s Solution) string
	SerializeSolution(s Solution) ([]byte, error)
}

type DatasetLoader interface {
	LoadDataset(path string) (Problem, error)
}
