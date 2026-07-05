package jobshop

// --- JSS Scoring and Validation ---

// Violation describes a hard constraint violation.
type Violation struct {
	Code   string // "PRECEDENCE", "OVERLAP", "MACHINE_MISMATCH"
	Detail string
}

// Validate checks a schedule for hard constraint violations.
func Validate(ds *Dataset, sched *Schedule) []Violation {
	var violations []Violation

	// Build lookup: (jobID, opIndex) → ScheduledOp.
	lookup := make(map[[2]int]*ScheduledOp)
	for i := range sched.Ops {
		key := [2]int{sched.Ops[i].JobID, sched.Ops[i].OpIndex}
		lookup[key] = &sched.Ops[i]
	}

	// Check precedence: each operation starts after previous in same job ends.
	for j := 0; j < ds.Jobs; j++ {
		for op := 1; op < ds.Machines; op++ {
			prev := lookup[[2]int{j, op - 1}]
			curr := lookup[[2]int{j, op}]
			if prev == nil || curr == nil {
				continue
			}
			if curr.Start < prev.End {
				violations = append(violations, Violation{
					Code:   "PRECEDENCE",
					Detail: formatPrecedenceViolation(j, op, prev.End, curr.Start),
				})
			}
		}
	}

	// Check machine overlap: no two operations on the same machine at the same time.
	// Group by machine.
	machineOps := make(map[int][]ScheduledOp)
	for _, op := range sched.Ops {
		machineOps[op.Machine] = append(machineOps[op.Machine], op)
	}
	for m, ops := range machineOps {
		for i := 0; i < len(ops); i++ {
			for j := i + 1; j < len(ops); j++ {
				if overlaps(ops[i], ops[j]) {
					violations = append(violations, Violation{
						Code:   "OVERLAP",
						Detail: formatOverlapViolation(m, ops[i], ops[j]),
					})
				}
			}
		}
	}

	// Check machine assignment: each operation runs on its designated machine.
	for _, op := range sched.Ops {
		expected := ds.JobList[op.JobID].Operations[op.OpIndex].Machine
		if op.Machine != expected {
			violations = append(violations, Violation{
				Code:   "MACHINE_MISMATCH",
				Detail: formatMachineMismatch(op.JobID, op.OpIndex, expected, op.Machine),
			})
		}
	}

	return violations
}

// ComputeMakespan returns the maximum completion time.
func ComputeMakespan(sched *Schedule) int {
	makespan := 0
	for _, op := range sched.Ops {
		if op.End > makespan {
			makespan = op.End
		}
	}
	return makespan
}

// --- Helpers ---

func overlaps(a, b ScheduledOp) bool {
	return a.Start < b.End && b.Start < a.End
}

func formatPrecedenceViolation(job, op, prevEnd, currStart int) string {
	return "Job " + itoa(job) + " op " + itoa(op) + ": starts at " + itoa(currStart) + " before previous ends at " + itoa(prevEnd)
}

func formatOverlapViolation(machine int, a, b ScheduledOp) string {
	return "Machine " + itoa(machine) + ": job " + itoa(a.JobID) + " op " + itoa(a.OpIndex) +
		" [" + itoa(a.Start) + "-" + itoa(a.End) + "] overlaps job " + itoa(b.JobID) + " op " + itoa(b.OpIndex) +
		" [" + itoa(b.Start) + "-" + itoa(b.End) + "]"
}

func formatMachineMismatch(job, op, expected, actual int) string {
	return "Job " + itoa(job) + " op " + itoa(op) + ": expected machine " + itoa(expected) + ", got " + itoa(actual)
}

func itoa(n int) string {
	if n < 0 {
		return "-" + itoa(-n)
	}
	if n < 10 {
		return string(rune('0' + n))
	}
	return itoa(n/10) + string(rune('0'+n%10))
}
