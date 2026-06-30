package inrc2

import (
	"fmt"
	"os"
	"strings"
)

// --- Plateau Detection Instrumentation ---
// Pure observation. Does not alter algorithm behaviour, scoring, or acceptance.
// A plateau is recorded when a worker fails to improve its local best for
// PlateauThreshold consecutive scored candidates, measured only after temperature
// drops below 25% of initial (to exclude intentional exploration phase).

// PlateauThreshold is the default number of candidates without local improvement
// before a plateau event is recorded.
const PlateauThreshold = 100000

// PlateauEvent records one observed stagnation period.
type PlateauEvent struct {
	WorkerID         int
	ParentWorkerID   int
	Week             int
	Candidate        int     // candidate number when plateau was detected
	Temperature      float64 // temperature at plateau detection
	CurrentPenalty   int     // penalty at detection
	LocalBest        int     // worker's local best (unchanged for threshold)
	GlobalBest       int     // global best at detection time
	CandsSinceImprove int    // candidates since last local improvement
	Depth            int     // worker depth in branch tree
}

// plateauObserver tracks stagnation within a single worker.
// Pure observation — reads only, never modifies search state.
type plateauObserver struct {
	threshold          int
	initialTemperature float64
	activationTemp     float64 // initialTemp * 0.25
	workerID           int
	parentWorkerID     int
	depth              int
	week               int

	lastImproveCand    int  // candidate number of last local best improvement
	active             bool // true once temp <= activationTemp
	events             []PlateauEvent
}

func newPlateauObserver(workerID, parentWorkerID, depth, week int, initialTemp float64) *plateauObserver {
	return &plateauObserver{
		threshold:          PlateauThreshold,
		initialTemperature: initialTemp,
		activationTemp:     initialTemp * 0.25,
		workerID:           workerID,
		parentWorkerID:     parentWorkerID,
		depth:              depth,
		week:               week,
		lastImproveCand:    0,
	}
}

// observe is called after each scored candidate. It checks for plateau conditions.
// Returns true if a new plateau was just recorded (for testing/debugging only).
func (p *plateauObserver) observe(candidate int, temperature float64, currentPenalty, localBest int, globalBest int64) bool {
	// Activate once temperature drops below threshold.
	if !p.active {
		if temperature <= p.activationTemp {
			p.active = true
			p.lastImproveCand = candidate // start measurement from activation
		}
		return false
	}

	// Check for plateau.
	candsSinceImprove := candidate - p.lastImproveCand
	if candsSinceImprove >= p.threshold {
		p.events = append(p.events, PlateauEvent{
			WorkerID:          p.workerID,
			ParentWorkerID:    p.parentWorkerID,
			Week:              p.week,
			Candidate:         candidate,
			Temperature:       temperature,
			CurrentPenalty:    currentPenalty,
			LocalBest:         localBest,
			GlobalBest:        int(globalBest),
			CandsSinceImprove: candsSinceImprove,
			Depth:             p.depth,
		})
		// Reset to avoid recording every subsequent candidate as a plateau.
		p.lastImproveCand = candidate
		return true
	}
	return false
}

// recordImprovement resets the counter when local best improves.
func (p *plateauObserver) recordImprovement(candidate int) {
	p.lastImproveCand = candidate
}

// --- Plateau CSV Export ---

// PlateauCSVHeader returns the CSV header for plateau events.
func PlateauCSVHeader() string {
	cols := []string{
		"week", "worker_id", "parent_worker_id", "depth",
		"candidate", "temperature", "current_penalty",
		"local_best", "global_best", "cands_since_improve",
	}
	return strings.Join(cols, ",")
}

// PlateauCSVRow formats a PlateauEvent as a CSV line.
func PlateauCSVRow(e PlateauEvent) string {
	return fmt.Sprintf("%d,%d,%d,%d,%d,%.6f,%d,%d,%d,%d",
		e.Week, e.WorkerID, e.ParentWorkerID, e.Depth,
		e.Candidate, e.Temperature, e.CurrentPenalty,
		e.LocalBest, e.GlobalBest, e.CandsSinceImprove)
}

// WritePlateauCSV writes plateau events to a CSV file.
func WritePlateauCSV(path string, events []PlateauEvent) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintln(f, PlateauCSVHeader())
	for _, e := range events {
		fmt.Fprintln(f, PlateauCSVRow(e))
	}
	return nil
}
