// Package ilp provides a disjunctive ILP formulation for the Job Shop Scheduling Problem.
package ilp

import (
	"fmt"
	"os"
	"strings"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/jobshop"
)

// ModelInfo describes the generated LP model.
type ModelInfo struct {
	Variables   int
	Constraints int
	Jobs        int
	Machines    int
	Operations  int
}

// BuildModel writes a JSS LP model in CPLEX LP format.
func BuildModel(ds *jobshop.Dataset, modelPath string) (ModelInfo, error) {
	numOps := len(ds.AllOps)
	if numOps == 0 {
		return ModelInfo{}, fmt.Errorf("no operations in dataset")
	}

	bigM := 0
	for _, op := range ds.AllOps {
		bigM += op.Duration
	}
	if bigM == 0 {
		bigM = 1
	}

	opIndex := make(map[[2]int]int, numOps)
	for k, op := range ds.AllOps {
		opIndex[[2]int{op.JobID, op.OpIndex}] = k
	}

	var b strings.Builder

	b.WriteString("Minimize\n obj: C\n\n")
	b.WriteString("Subject To\n")
	conCount := 0

	// Makespan bounds.
	for k, op := range ds.AllOps {
		b.WriteString(fmt.Sprintf(" mk_%d: C - s_%d >= %d\n", k, k, op.Duration))
		conCount++
	}

	// Job precedence.
	for _, job := range ds.JobList {
		for op := 0; op < len(job.Operations)-1; op++ {
			curr := job.Operations[op]
			next := job.Operations[op+1]
			ki := opIndex[[2]int{curr.JobID, curr.OpIndex}]
			kj := opIndex[[2]int{next.JobID, next.OpIndex}]
			b.WriteString(fmt.Sprintf(" prec_%d_%d: s_%d - s_%d >= %d\n",
				curr.JobID, next.OpIndex, kj, ki, curr.Duration))
			conCount++
		}
	}

	// Disjunctive constraints per machine.
	machineOps := make(map[int][]int)
	for k, op := range ds.AllOps {
		machineOps[op.Machine] = append(machineOps[op.Machine], k)
	}

	yCount := 0
	for machine, ops := range machineOps {
		for i := 0; i < len(ops); i++ {
			for j := i + 1; j < len(ops); j++ {
				a, c := ops[i], ops[j]
				opA := ds.AllOps[a]
				opC := ds.AllOps[c]
				b.WriteString(fmt.Sprintf(" disj_%d_%d_a: s_%d - s_%d + %d y_%d_%d <= %d\n",
					machine, yCount, c, a, bigM, a, c, bigM-opA.Duration))
				b.WriteString(fmt.Sprintf(" disj_%d_%d_b: s_%d - s_%d - %d y_%d_%d <= %d\n",
					machine, yCount, a, c, bigM, a, c, bigM-opC.Duration))
				conCount += 2
				yCount++
			}
		}
	}

	b.WriteString("\nBounds\n")
	b.WriteString(" 0 <= C\n")
	for k := range ds.AllOps {
		b.WriteString(fmt.Sprintf(" 0 <= s_%d\n", k))
	}

	b.WriteString("\nBinary\n")
	for machine, ops := range machineOps {
		for i := 0; i < len(ops); i++ {
			for j := i + 1; j < len(ops); j++ {
				a, c := ops[i], ops[j]
				_ = machine
				b.WriteString(fmt.Sprintf(" y_%d_%d\n", a, c))
			}
		}
	}

	b.WriteString("\nGenerals\n")
	b.WriteString(" C\n")
	for k := range ds.AllOps {
		b.WriteString(fmt.Sprintf(" s_%d\n", k))
	}

	b.WriteString("\nEnd\n")

	if err := os.WriteFile(modelPath, []byte(b.String()), 0644); err != nil {
		return ModelInfo{}, fmt.Errorf("write model: %w", err)
	}

	numVars := numOps + 1 + yCount // s + C + y
	return ModelInfo{
		Variables:   numVars,
		Constraints: conCount,
		Jobs:        ds.Jobs,
		Machines:    ds.Machines,
		Operations:  numOps,
	}, nil
}
