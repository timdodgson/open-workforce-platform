package jobshop

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

// JSSProblem implements optimisation.Problem for Job Shop Scheduling.
type JSSProblem struct {
	dataset *Dataset
}

// NewJSSProblem creates a JSS problem instance.
func NewJSSProblem(ds *Dataset) *JSSProblem {
	return &JSSProblem{dataset: ds}
}

// --- Internal solution representation ---
// The solution is represented as a permutation of operations.
// Each position in the permutation determines the scheduling priority.
// Decoding: process operations in permutation order, scheduling each
// at the earliest feasible time (respecting precedence + machine availability).

type jssSolution struct {
	perm []int // permutation of operation indices into dataset.AllOps
}

// --- Move types ---

type jssMove struct {
	posA int // first position in permutation
	posB int // second position in permutation
}

// --- optimisation.Problem implementation ---

func (p *JSSProblem) CreateInitialSolution() (optimisation.Solution, error) {
	// Use SPT-based constructive schedule, then encode as permutation.
	// The permutation is the order in which operations should be dispatched.
	// Start with a natural ordering (job 0 ops first, then job 1, etc.)
	// then sort by the SPT-constructed schedule start times.
	sched := BuildInitialSchedule(p.dataset)

	// Build permutation from schedule order (sorted by start time).
	type opWithStart struct {
		opIdx int
		start int
	}
	ops := make([]opWithStart, len(sched.Ops))
	for i, so := range sched.Ops {
		// Find the index in AllOps.
		opIdx := so.JobID*p.dataset.Machines + so.OpIndex
		ops[i] = opWithStart{opIdx: opIdx, start: so.Start}
	}
	sort.Slice(ops, func(i, j int) bool { return ops[i].start < ops[j].start })

	perm := make([]int, len(ops))
	for i, o := range ops {
		perm[i] = o.opIdx
	}

	return &jssSolution{perm: perm}, nil
}

func (p *JSSProblem) CloneSolution(s optimisation.Solution) optimisation.Solution {
	src := s.(*jssSolution)
	clone := &jssSolution{perm: make([]int, len(src.perm))}
	copy(clone.perm, src.perm)
	return clone
}

func (p *JSSProblem) Evaluate(s optimisation.Solution) int {
	sol := s.(*jssSolution)
	sched := p.decode(sol)
	return sched.Makespan
}

// TryMove swaps two random positions in the permutation.
// Always valid (any permutation produces a feasible schedule via decoding).
func (p *JSSProblem) TryMove(s optimisation.Solution, rng *rand.Rand) optimisation.MoveResult {
	sol := s.(*jssSolution)
	n := len(sol.perm)
	if n < 2 {
		return optimisation.MoveResult{Valid: false}
	}

	posA := rng.Intn(n)
	posB := rng.Intn(n - 1)
	if posB >= posA {
		posB++
	}

	// Apply swap.
	sol.perm[posA], sol.perm[posB] = sol.perm[posB], sol.perm[posA]

	return optimisation.MoveResult{
		Valid: true,
		Move:  jssMove{posA: posA, posB: posB},
	}
}

func (p *JSSProblem) UndoMove(s optimisation.Solution, m optimisation.Move) {
	sol := s.(*jssSolution)
	mv := m.(jssMove)
	// Swap is self-inverse.
	sol.perm[mv.posA], sol.perm[mv.posB] = sol.perm[mv.posB], sol.perm[mv.posA]
}

func (p *JSSProblem) SolutionFingerprint(s optimisation.Solution) string {
	sol := s.(*jssSolution)
	h := md5.New()
	for _, v := range sol.perm {
		fmt.Fprintf(h, "%d,", v)
	}
	return fmt.Sprintf("%x", h.Sum(nil)[:6])
}

func (p *JSSProblem) SerializeSolution(s optimisation.Solution) ([]byte, error) {
	sol := s.(*jssSolution)
	sched := p.decode(sol)

	type schedJSON struct {
		Makespan int           `json:"makespan"`
		Jobs     int           `json:"jobs"`
		Machines int           `json:"machines"`
		Ops      []ScheduledOp `json:"operations"`
	}

	out := schedJSON{
		Makespan: sched.Makespan,
		Jobs:     p.dataset.Jobs,
		Machines: p.dataset.Machines,
		Ops:      sched.Ops,
	}
	return json.Marshal(out)
}

// --- Decoding: permutation → schedule ---

// decode converts a permutation into a feasible schedule.
// Process operations in permutation order. For each operation,
// schedule it at the earliest time that satisfies:
//   - Machine availability (no overlap with other ops on same machine)
//   - Job precedence (starts after previous operation in same job ends)
func (p *JSSProblem) decode(sol *jssSolution) *Schedule {
	ds := p.dataset
	numJobs := ds.Jobs
	numMachines := ds.Machines

	// Track state.
	machineAvail := make([]int, numMachines) // earliest time each machine is free
	jobReady := make([]int, numJobs)         // earliest time next op in each job can start
	nextOpInJob := make([]int, numJobs)      // which operation index each job is up to

	scheduled := make([]ScheduledOp, 0, len(sol.perm))

	for _, opIdx := range sol.perm {
		op := ds.AllOps[opIdx]

		// Check precedence: can only schedule this op if all prior ops in the job are done.
		if op.OpIndex != nextOpInJob[op.JobID] {
			// This operation isn't ready yet (precedence not met).
			// Skip it — it will be placed later when its turn comes.
			// In a permutation-based representation, we need to handle this carefully.
			// Strategy: queue operations and process them when they become ready.
			continue
		}

		// Schedule at earliest feasible time.
		start := machineAvail[op.Machine]
		if jobReady[op.JobID] > start {
			start = jobReady[op.JobID]
		}
		end := start + op.Duration

		scheduled = append(scheduled, ScheduledOp{
			JobID:    op.JobID,
			OpIndex:  op.OpIndex,
			Machine:  op.Machine,
			Start:    start,
			End:      end,
			Duration: op.Duration,
		})

		machineAvail[op.Machine] = end
		jobReady[op.JobID] = end
		nextOpInJob[op.JobID]++
	}

	// Handle any ops that were skipped due to precedence ordering in permutation.
	// Process remaining operations in natural order.
	for placed := len(scheduled); placed < len(ds.AllOps); placed++ {
		// Find a job that has unscheduled operations.
		for j := 0; j < numJobs; j++ {
			if nextOpInJob[j] >= numMachines {
				continue
			}
			op := ds.JobList[j].Operations[nextOpInJob[j]]
			start := machineAvail[op.Machine]
			if jobReady[j] > start {
				start = jobReady[j]
			}
			end := start + op.Duration

			scheduled = append(scheduled, ScheduledOp{
				JobID:    j,
				OpIndex:  nextOpInJob[j],
				Machine:  op.Machine,
				Start:    start,
				End:      end,
				Duration: op.Duration,
			})

			machineAvail[op.Machine] = end
			jobReady[j] = end
			nextOpInJob[j]++
			break
		}
	}

	makespan := 0
	for _, op := range scheduled {
		if op.End > makespan {
			makespan = op.End
		}
	}

	return &Schedule{Ops: scheduled, Makespan: makespan}
}

// --- Accessors ---

func (p *JSSProblem) Dataset() *Dataset { return p.dataset }

// --- Compile-time interface check ---
var _ optimisation.Problem = (*JSSProblem)(nil)
