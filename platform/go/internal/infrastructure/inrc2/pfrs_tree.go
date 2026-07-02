package inrc2

import (
	"crypto/md5"
	"fmt"
	"os"
	"strings"
)

// --- PFRS Search Tree Visibility ---
// First-class representation of the branching tree.
// Pure observation — does not change algorithm behaviour.

// TreeNode represents a single worker in the search tree.
type TreeNode struct {
	// Identity.
	WorkerID       int
	ParentWorkerID int // -1 for root
	Depth          int
	ChildrenCount  int

	// Penalty trajectory.
	StartPenalty         int
	BestPenalty          int
	FinalPenalty         int
	ImprovementFromStart int // StartPenalty - BestPenalty
	ImprovementOverParent int // parent.BestPenalty - this.BestPenalty

	// Work metrics.
	Candidates       int
	Attempts         int
	HardRejects      int
	AcceptedImproving int
	AcceptedWorse    int
	RejectedByProb   int

	// Temperature.
	StartTemperature float64
	TempAtBest       float64
	FinalTemperature float64

	// Outcomes.
	ProducedGlobalBest bool

	// Structural metrics.
	RosterFingerprint         string  // MD5 of roster assignments
	HistoryFingerprint        string  // MD5 of history state
	HammingDistanceParent     float64 // 0.0 - 1.0, fraction of changed cells
	HammingDistanceRoot       float64
	HammingDistanceGlobalBest float64 // distance from global best at time of branching
}

// SearchTree holds all tree nodes for one PFRS execution.
type SearchTree struct {
	Nodes []TreeNode
}

// WinningLineage returns the path from root to the winning worker.
func (t *SearchTree) WinningLineage(finalPenalty int) []TreeNode {
	// Find the winning node.
	var winner *TreeNode
	for i := range t.Nodes {
		if t.Nodes[i].BestPenalty == finalPenalty && t.Nodes[i].ProducedGlobalBest {
			winner = &t.Nodes[i]
			break
		}
	}
	if winner == nil {
		return nil
	}

	// Build index.
	idx := make(map[int]*TreeNode, len(t.Nodes))
	for i := range t.Nodes {
		idx[t.Nodes[i].WorkerID] = &t.Nodes[i]
	}

	// Walk up to root.
	var lineage []TreeNode
	current := winner
	for current != nil {
		lineage = append(lineage, *current)
		if current.ParentWorkerID == -1 {
			break
		}
		current = idx[current.ParentWorkerID]
	}

	// Reverse to get root-first order.
	for i, j := 0, len(lineage)-1; i < j; i, j = i+1, j-1 {
		lineage[i], lineage[j] = lineage[j], lineage[i]
	}
	return lineage
}

// DepthSummary holds aggregate metrics for one tree depth level.
type DepthSummary struct {
	Depth              int
	WorkerCount        int
	BestPenalty        int
	AvgBestPenalty     int
	AvgImprovement     int
	AvgDistanceParent  float64
	ProducedGlobalBest int
}

// DepthSummaries computes per-depth aggregate statistics.
func (t *SearchTree) DepthSummaries() []DepthSummary {
	depthMap := make(map[int]*DepthSummary)

	for _, n := range t.Nodes {
		ds, ok := depthMap[n.Depth]
		if !ok {
			ds = &DepthSummary{Depth: n.Depth, BestPenalty: n.BestPenalty}
			depthMap[n.Depth] = ds
		}
		ds.WorkerCount++
		ds.AvgBestPenalty += n.BestPenalty
		ds.AvgImprovement += n.ImprovementFromStart
		ds.AvgDistanceParent += n.HammingDistanceParent
		if n.BestPenalty < ds.BestPenalty {
			ds.BestPenalty = n.BestPenalty
		}
		if n.ProducedGlobalBest {
			ds.ProducedGlobalBest++
		}
	}

	// Compute averages and collect sorted.
	var result []DepthSummary
	maxDepth := 0
	for d := range depthMap {
		if d > maxDepth {
			maxDepth = d
		}
	}
	for d := 0; d <= maxDepth; d++ {
		ds, ok := depthMap[d]
		if !ok {
			continue
		}
		if ds.WorkerCount > 0 {
			ds.AvgBestPenalty /= ds.WorkerCount
			ds.AvgImprovement /= ds.WorkerCount
			ds.AvgDistanceParent /= float64(ds.WorkerCount)
		}
		result = append(result, *ds)
	}
	return result
}

// DiversitySummary holds overall diversity metrics.
type DiversitySummary struct {
	AvgParentChildDistance float64
	MinParentChildDistance float64
	MaxParentChildDistance float64
	AvgDistanceFromRoot   float64
	NearDuplicateCount    int     // branches with hamming < threshold
	NearDuplicateThreshold float64 // default 0.05
}

// ComputeDiversity calculates structural diversity metrics.
func (t *SearchTree) ComputeDiversity(threshold float64) DiversitySummary {
	if threshold <= 0 {
		threshold = 0.05
	}
	ds := DiversitySummary{
		MinParentChildDistance:  1.0,
		NearDuplicateThreshold: threshold,
	}

	count := 0
	for _, n := range t.Nodes {
		if n.ParentWorkerID == -1 {
			continue // root has no parent distance
		}
		ds.AvgParentChildDistance += n.HammingDistanceParent
		ds.AvgDistanceFromRoot += n.HammingDistanceRoot
		if n.HammingDistanceParent < ds.MinParentChildDistance {
			ds.MinParentChildDistance = n.HammingDistanceParent
		}
		if n.HammingDistanceParent > ds.MaxParentChildDistance {
			ds.MaxParentChildDistance = n.HammingDistanceParent
		}
		if n.HammingDistanceParent < threshold {
			ds.NearDuplicateCount++
		}
		count++
	}

	if count > 0 {
		ds.AvgParentChildDistance /= float64(count)
		ds.AvgDistanceFromRoot /= float64(count)
	}
	return ds
}

// --- Roster Distance Functions ---

// RosterHammingDistance computes the fraction of nurse-day cells that differ between two rosters.
// Returns 0.0 for identical, 1.0 for completely different.
func RosterHammingDistance(a, b *Roster) float64 {
	if a == nil || b == nil {
		return 1.0
	}
	if len(a.NurseIDs) == 0 || a.NumDays == 0 {
		return 1.0
	}

	totalCells := len(a.NurseIDs) * a.NumDays
	diffCells := 0

	for ni := 0; ni < len(a.NurseIDs) && ni < len(b.Assignments); ni++ {
		for d := 0; d < a.NumDays && d < len(a.Assignments[ni]) && d < len(b.Assignments[ni]); d++ {
			aa := a.Assignments[ni][d]
			ba := b.Assignments[ni][d]
			if aa.ShiftType != ba.ShiftType || aa.Skill != ba.Skill {
				diffCells++
			}
		}
	}

	if totalCells == 0 {
		return 0.0
	}
	return float64(diffCells) / float64(totalCells)
}

// RosterFingerprint computes a short hash of a roster's assignments.
func RosterFingerprint(r *Roster) string {
	if r == nil {
		return ""
	}
	h := md5.New()
	for ni := range r.Assignments {
		for d := range r.Assignments[ni] {
			a := r.Assignments[ni][d]
			fmt.Fprintf(h, "%d:%d:%s:%s|", ni, d, a.ShiftType, a.Skill)
		}
	}
	return fmt.Sprintf("%x", h.Sum(nil))[:12]
}

// HistoryFingerprint computes a short hash of history state.
func HistoryFingerprint(hist History) string {
	h := md5.New()
	fmt.Fprintf(h, "w%d|", hist.Week)
	for _, nh := range hist.NurseHistory {
		fmt.Fprintf(h, "%s:%d:%d:%d:%s:%d|",
			nh.Nurse, nh.NumberOfAssignments, nh.NumberOfWorkingWeekends,
			nh.NumberOfConsecutiveWorkingDays, nh.LastAssignedShiftType,
			nh.NumberOfConsecutiveDaysOff)
	}
	return fmt.Sprintf("%x", h.Sum(nil))[:12]
}

// --- Tree CSV ---

// TreeCSVHeader returns the CSV header for tree node output.
func TreeCSVHeader() string {
	cols := []string{
		"week", "seed", "worker_id", "parent_worker_id", "depth", "children_count",
		"start_penalty", "best_penalty", "final_penalty",
		"improvement_from_start", "improvement_over_parent",
		"candidates", "attempts", "hard_rejects",
		"accepted_improving", "accepted_worse", "rejected_by_probability",
		"start_temperature", "temperature_at_best", "final_temperature",
		"produced_global_best",
		"roster_fingerprint", "history_fingerprint",
		"hamming_distance_parent", "hamming_distance_root", "hamming_distance_global_at_branch",
	}
	return strings.Join(cols, ",")
}

// TreeCSVRow formats a TreeNode as a CSV row.
func TreeCSVRow(week int, seed int64, n TreeNode) string {
	globalBest := 0
	if n.ProducedGlobalBest {
		globalBest = 1
	}
	return fmt.Sprintf("%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%.6f,%.6f,%.6f,%d,%s,%s,%.4f,%.4f,%.4f",
		week, seed, n.WorkerID, n.ParentWorkerID, n.Depth, n.ChildrenCount,
		n.StartPenalty, n.BestPenalty, n.FinalPenalty,
		n.ImprovementFromStart, n.ImprovementOverParent,
		n.Candidates, n.Attempts, n.HardRejects,
		n.AcceptedImproving, n.AcceptedWorse, n.RejectedByProb,
		n.StartTemperature, n.TempAtBest, n.FinalTemperature,
		globalBest,
		n.RosterFingerprint, n.HistoryFingerprint,
		n.HammingDistanceParent, n.HammingDistanceRoot, n.HammingDistanceGlobalBest,
	)
}

// WriteTreeCSV writes a complete tree CSV file.
func WriteTreeCSV(path string, week int, seed int64, tree SearchTree) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintln(f, TreeCSVHeader())
	for _, n := range tree.Nodes {
		fmt.Fprintln(f, TreeCSVRow(week, seed, n))
	}
	return nil
}

// AppendTreeCSV appends tree nodes to an existing CSV file (no header).
func AppendTreeCSV(path string, week int, seed int64, tree SearchTree) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write header if file is empty.
	info, _ := f.Stat()
	if info != nil && info.Size() == 0 {
		fmt.Fprintln(f, TreeCSVHeader())
	}

	for _, n := range tree.Nodes {
		fmt.Fprintln(f, TreeCSVRow(week, seed, n))
	}
	return nil
}
