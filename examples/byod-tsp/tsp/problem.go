package tsp

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/timdodgson/open-workforce-platform/owp-sdk/searchdef"
)

// Problem is a minimal symmetric TSP implementing searchdef.Problem (BYOD demo).
type Problem struct {
	dataset *Dataset
}

func NewProblem(ds *Dataset) *Problem {
	return &Problem{dataset: ds}
}

type tourSolution struct {
	order []int
}

type swapMove struct {
	i, j int
}

func (p *Problem) CreateInitialSolution() (searchdef.Solution, error) {
	order := make([]int, len(p.dataset.Cities))
	for i := range order {
		order[i] = i
	}
	return &tourSolution{order: order}, nil
}

func (p *Problem) CloneSolution(s searchdef.Solution) searchdef.Solution {
	src := s.(*tourSolution)
	clone := make([]int, len(src.order))
	copy(clone, src.order)
	return &tourSolution{order: clone}
}

func (p *Problem) Evaluate(s searchdef.Solution) int {
	order := s.(*tourSolution).order
	if len(order) == 0 {
		return 0
	}
	total := 0
	for i := 0; i < len(order); i++ {
		a := order[i]
		b := order[(i+1)%len(order)]
		total += p.dataset.Distance(a, b)
	}
	return total
}

func (p *Problem) TryMove(s searchdef.Solution, rng *rand.Rand) searchdef.MoveResult {
	order := s.(*tourSolution).order
	n := len(order)
	if n < 2 {
		return searchdef.MoveResult{Valid: false}
	}
	i := rng.Intn(n)
	j := rng.Intn(n - 1)
	if j >= i {
		j++
	}
	order[i], order[j] = order[j], order[i]
	return searchdef.MoveResult{Valid: true, Move: swapMove{i: i, j: j}}
}

func (p *Problem) UndoMove(s searchdef.Solution, m searchdef.Move) {
	mv := m.(swapMove)
	order := s.(*tourSolution).order
	order[mv.i], order[mv.j] = order[mv.j], order[mv.i]
}

func (p *Problem) SolutionFingerprint(s searchdef.Solution) string {
	order := s.(*tourSolution).order
	h := md5.Sum([]byte(fmt.Sprint(order)))
	return fmt.Sprintf("%x", h[:8])
}

func (p *Problem) SerializeSolution(s searchdef.Solution) ([]byte, error) {
	order := s.(*tourSolution).order
	out := make([]int, len(order))
	copy(out, order)
	return json.Marshal(map[string]interface{}{
		"tour":  out,
		"tourLength": p.Evaluate(s),
	})
}
