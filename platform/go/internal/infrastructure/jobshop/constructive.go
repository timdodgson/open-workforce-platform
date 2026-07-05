package jobshop

// BuildInitialSchedule constructs a valid schedule using a simple dispatch rule.
// Uses Shortest Processing Time (SPT) priority: at each decision point, schedule
// the ready operation with the shortest duration.
//
// Guarantees:
//   - All precedence constraints satisfied (operations in job order)
//   - No machine overlaps
//   - Feasible schedule (may not be optimal)
func BuildInitialSchedule(ds *Dataset) *Schedule {
	numJobs := ds.Jobs
	numMachines := ds.Machines

	// Track next operation index per job.
	nextOp := make([]int, numJobs)

	// Track earliest available time per machine.
	machineAvail := make([]int, numMachines)

	// Track completion time of last operation per job (precedence).
	jobReady := make([]int, numJobs)

	scheduled := make([]ScheduledOp, 0, len(ds.AllOps))
	totalOps := numJobs * numMachines
	placed := 0

	for placed < totalOps {
		// Find the next operation to schedule (earliest start time, tie-break by shortest duration).
		bestJob := -1
		bestStart := int(^uint(0) >> 1)
		bestDuration := int(^uint(0) >> 1)

		for j := 0; j < numJobs; j++ {
			if nextOp[j] >= numMachines {
				continue // job complete
			}
			op := ds.JobList[j].Operations[nextOp[j]]
			// Earliest start = max(machine available, job precedence ready).
			earliest := machineAvail[op.Machine]
			if jobReady[j] > earliest {
				earliest = jobReady[j]
			}

			// SPT: prefer earlier start, then shorter duration.
			if earliest < bestStart || (earliest == bestStart && op.Duration < bestDuration) {
				bestJob = j
				bestStart = earliest
				bestDuration = op.Duration
			}
		}

		if bestJob == -1 {
			break // shouldn't happen if data is valid
		}

		op := ds.JobList[bestJob].Operations[nextOp[bestJob]]
		start := bestStart
		end := start + op.Duration

		scheduled = append(scheduled, ScheduledOp{
			JobID:    bestJob,
			OpIndex:  nextOp[bestJob],
			Machine:  op.Machine,
			Start:    start,
			End:      end,
			Duration: op.Duration,
		})

		// Update state.
		machineAvail[op.Machine] = end
		jobReady[bestJob] = end
		nextOp[bestJob]++
		placed++
	}

	// Compute makespan.
	makespan := 0
	for _, op := range scheduled {
		if op.End > makespan {
			makespan = op.End
		}
	}

	return &Schedule{
		Ops:      scheduled,
		Makespan: makespan,
	}
}
