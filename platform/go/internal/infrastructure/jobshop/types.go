// Package jobshop implements the Job Shop Scheduling Problem (JSS) domain.
//
// The JSS asks: given N jobs each consisting of M ordered operations,
// where each operation requires a specific machine for a specific duration,
// find a schedule that minimises the makespan (completion time of all jobs).
//
// Hard constraints:
//   - Each operation runs on its designated machine
//   - Operations within a job execute in order (precedence)
//   - No machine processes two operations at the same time (no overlap)
//
// Objective: minimise makespan (max completion time across all jobs)
//
// This package models the domain and implements the generic
// optimisation.Problem interface.
package jobshop

// Operation represents a single task within a job.
type Operation struct {
	JobID     int // which job this belongs to
	OpIndex   int // position within the job (0-based)
	Machine   int // which machine this must run on
	Duration  int // processing time
}

// Job represents an ordered sequence of operations.
type Job struct {
	ID         int
	Operations []Operation
}

// Dataset holds a complete JSS problem instance.
type Dataset struct {
	Name     string
	Jobs     int // number of jobs
	Machines int // number of machines
	AllOps   []Operation // all operations (flat list)
	JobList  []Job       // operations grouped by job
}

// ScheduledOp represents one operation placed in a schedule.
type ScheduledOp struct {
	JobID    int
	OpIndex  int
	Machine  int
	Start    int
	End      int
	Duration int
}

// Schedule represents a complete JSS solution.
type Schedule struct {
	Ops      []ScheduledOp
	Makespan int // max End across all operations
}

// MachineSchedule returns operations scheduled on a specific machine, sorted by start time.
func (s *Schedule) MachineSchedule(machine int) []ScheduledOp {
	var result []ScheduledOp
	for _, op := range s.Ops {
		if op.Machine == machine {
			result = append(result, op)
		}
	}
	return result
}

// JobSchedule returns operations for a specific job, sorted by operation index.
func (s *Schedule) JobSchedule(jobID int) []ScheduledOp {
	var result []ScheduledOp
	for _, op := range s.Ops {
		if op.JobID == jobID {
			result = append(result, op)
		}
	}
	return result
}
