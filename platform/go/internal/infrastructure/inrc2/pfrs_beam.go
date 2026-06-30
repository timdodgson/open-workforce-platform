package inrc2

import "sort"

// --- PFRS Multi-History Beam Search ---
// Pure orchestration over week histories. Does not change PFRS worker behaviour or scoring.

// BeamPath represents one candidate history path through the planning horizon.
type BeamPath struct {
	ID               int
	ParentID         int // -1 for root
	Week             int // 1-indexed (last completed week)
	CumulativePenalty int
	WeekPenalty      int
	CumulativeSoft   int
	WeekSoft         int
	History          History  // output history after this week
	Solution         Solution // solution for this week
	Seed             int64    // seed used for this week's PFRS run
	Valid            bool     // Hard == 0 for this week
	Stats            PFRSStats   // PFRS execution stats for this week's run
	ScoreResult      ScoreResult // official score result
	Audit            PFRSAudit   // audit data from this week's run

	// Diversity metrics — computed after beam pruning.
	Fingerprint      string  // 12-char MD5 hash of roster assignments
	HammingToBest    float64 // Hamming distance to best path this week (0.0-1.0)
}

// BeamResult holds the output of a full beam search across all weeks.
type BeamResult struct {
	WinningPath    []BeamPath // one entry per week for the best full-horizon path
	AllPaths       []BeamPath // all candidate paths generated (for audit)
	TotalPenalty   int
	TotalSoft      int
	AllValid       bool
	WeekSummaries  []BeamWeekSummary
}

// BeamWeekSummary captures per-week beam search statistics.
type BeamWeekSummary struct {
	Week           int
	Candidates     int // paths generated this week
	Retained       int // paths kept after pruning
	BestCumulative int // best cumulative penalty among retained paths
}

// BeamConfig holds beam search parameters.
type BeamConfig struct {
	BeamWidth    int     // max paths retained per week
	Seeds        []int64 // seeds to expand each path with
}

// RunBeamSearch executes PFRS with multi-history beam search across all weeks.
// For each week, expands each retained path with each seed, keeps top beamWidth by cumulative penalty.
// beamWidth=1 with a single seed reproduces existing single-path behaviour.
func RunBeamSearch(sc Scenario, weekFiles []string, initialHist History,
	baseConfig PFRSConfig, beam BeamConfig, onWeekProgress func(week int, path BeamPath)) (BeamResult, error) {

	numWeeks := sc.NumberOfWeeks
	if numWeeks > len(weekFiles) {
		numWeeks = len(weekFiles)
	}

	// Start with a single root path.
	currentPaths := []BeamPath{{
		ID:       0,
		ParentID: -1,
		Week:     0,
		History:  initialHist,
		Valid:    true,
	}}

	nextID := 1
	var weekSummaries []BeamWeekSummary
	var allPaths []BeamPath

	for w := 0; w < numWeeks; w++ {
		wd, err := LoadWeekData(weekFiles[w])
		if err != nil {
			return BeamResult{}, err
		}

		var candidates []BeamPath

		// Expand each retained path with each seed.
		for _, path := range currentPaths {
			for _, seed := range beam.Seeds {
				config := baseConfig
				config.Seed = seed

				// Set up audit capture for this run.
				var runAudit PFRSAudit
				config.OnAudit = func(a PFRSAudit) {
					runAudit = a
				}

				sol, stats, scoreResult, err := SolveWeekPFRS(sc, wd, path.History, config)
				if err != nil {
					// PFRS failed — skip this candidate.
					continue
				}

				// Only keep hard-valid paths.
				if scoreResult.HardViolations != 0 {
					continue
				}

				newHist := UpdateHistory(sc, path.History, sol)
				candidate := BeamPath{
					ID:                nextID,
					ParentID:          path.ID,
					Week:              w + 1,
					CumulativePenalty: path.CumulativePenalty + scoreResult.SoftPenalty,
					WeekPenalty:       scoreResult.SoftPenalty,
					CumulativeSoft:    path.CumulativeSoft + len(scoreResult.SoftDetails),
					WeekSoft:          len(scoreResult.SoftDetails),
					History:           newHist,
					Solution:          sol,
					Seed:              seed,
					Valid:             true,
					Stats:             stats,
					ScoreResult:       scoreResult,
					Audit:             runAudit,
				}
				nextID++
				candidates = append(candidates, candidate)

				if onWeekProgress != nil {
					onWeekProgress(w+1, candidate)
				}
			}
		}

		if len(candidates) == 0 {
			// No valid paths — beam search failed for this week.
			return BeamResult{AllValid: false}, nil
		}

		// Rank by cumulative penalty ascending, keep top beamWidth.
		sort.SliceStable(candidates, func(i, j int) bool {
			return candidates[i].CumulativePenalty < candidates[j].CumulativePenalty
		})

		retained := beam.BeamWidth
		if retained > len(candidates) {
			retained = len(candidates)
		}
		currentPaths = candidates[:retained]

		// Compute diversity metrics for all candidates this week.
		// Reconstruct rosters from solutions to compute fingerprints and Hamming distances.
		rosters := make([]*Roster, len(candidates))
		for i := range candidates {
			rosters[i] = SolutionToRoster(candidates[i].Solution, sc)
			candidates[i].Fingerprint = RosterFingerprint(rosters[i])
		}
		// Best path is the first retained (lowest cumulative penalty).
		bestRoster := rosters[0]
		for i := range candidates {
			candidates[i].HammingToBest = RosterHammingDistance(rosters[i], bestRoster)
		}

		// Track all candidates for audit.
		allPaths = append(allPaths, candidates...)

		weekSummaries = append(weekSummaries, BeamWeekSummary{
			Week:           w + 1,
			Candidates:     len(candidates),
			Retained:       retained,
			BestCumulative: currentPaths[0].CumulativePenalty,
		})
	}

	// Best path is the first in the final retained set.
	best := currentPaths[0]

	// Reconstruct the winning lineage by walking parent IDs.
	pathIndex := make(map[int]BeamPath, len(allPaths))
	for _, p := range allPaths {
		pathIndex[p.ID] = p
	}

	var winningLineage []BeamPath
	current := best
	for {
		winningLineage = append([]BeamPath{current}, winningLineage...)
		if current.ParentID <= 0 {
			break
		}
		parent, ok := pathIndex[current.ParentID]
		if !ok {
			break
		}
		current = parent
	}

	result := BeamResult{
		WinningPath:   winningLineage,
		AllPaths:      allPaths,
		TotalPenalty:  best.CumulativePenalty,
		TotalSoft:     best.CumulativeSoft,
		AllValid:      true,
		WeekSummaries: weekSummaries,
	}

	return result, nil
}
