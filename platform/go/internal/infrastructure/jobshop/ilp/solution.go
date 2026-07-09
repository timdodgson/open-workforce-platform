package ilp

import (
	"encoding/json"
	"fmt"
	"math"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/jobshop"
	nrpilp "github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/ilp"
)

type solutionJSON struct {
	Makespan   int                      `json:"makespan"`
	Jobs       int                      `json:"jobs"`
	Machines   int                      `json:"machines"`
	Operations []jobshop.ScheduledOp    `json:"operations"`
}

// ExtractSolution maps HiGHS start-time variables to a JSS schedule JSON.
func ExtractSolution(ds *jobshop.Dataset, output nrpilp.SolverOutput) ([]byte, error) {
	if len(output.SolutionValues) == 0 {
		return nil, fmt.Errorf("no solution values in solver output")
	}

	makespan := int(math.Round(output.Objective))
	if c, ok := output.SolutionValues["C"]; ok {
		makespan = int(math.Round(c))
	}

	ops := make([]jobshop.ScheduledOp, 0, len(ds.AllOps))
	for k, op := range ds.AllOps {
		key := fmt.Sprintf("s_%d", k)
		startF, ok := output.SolutionValues[key]
		if !ok {
			return nil, fmt.Errorf("missing variable %s", key)
		}
		start := int(math.Round(startF))
		ops = append(ops, jobshop.ScheduledOp{
			JobID:    op.JobID,
			OpIndex:  op.OpIndex,
			Machine:  op.Machine,
			Start:    start,
			End:      start + op.Duration,
			Duration: op.Duration,
		})
	}

	sched := &jobshop.Schedule{Ops: ops, Makespan: makespan}
	violations := jobshop.Validate(ds, sched)
	if len(violations) > 0 {
		return nil, fmt.Errorf("extracted schedule has %d hard violations", len(violations))
	}

	out := solutionJSON{
		Makespan:   jobshop.ComputeMakespan(sched),
		Jobs:       ds.Jobs,
		Machines:   ds.Machines,
		Operations: ops,
	}
	return json.Marshal(out)
}
