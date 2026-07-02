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
	BeamWidth         int     // max paths retained per week
	Seeds             []int64 // seeds to expand each path with
	FinalWindowWeeks  int     // number of final weeks to optimise as a coupled block (default 1 = normal)
	FinalWindowIter   int     // iteration override for final window workers (0 = use base config)
	LookaheadWeight   float64 // weight for amortized global constraint look-ahead (0 = disabled)
	DiversitySlotsPct int     // % of beam width reserved for diversity picks (0 = disabled, e.g. 30 = 30%)
	BeamStrategy      string  // "none" (default), "lookahead", or "budget"
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

	// Determine final window boundaries.
	finalWindowWeeks := beam.FinalWindowWeeks
	if finalWindowWeeks <= 0 {
		finalWindowWeeks = 1 // default: no coupling
	}
	// How many weeks to run in normal beam mode before the final window.
	normalWeeks := numWeeks - finalWindowWeeks
	if normalWeeks < 0 {
		normalWeeks = 0
		finalWindowWeeks = numWeeks
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

	// --- Phase 1: Normal beam search for weeks 1..normalWeeks ---
	for w := 0; w < normalWeeks; w++ {
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

		// Rank by cumulative penalty + strategy bias, keep top beamWidth.
		// Official CumulativePenalty in the BeamPath is NOT modified.
		sort.SliceStable(candidates, func(i, j int) bool {
			var iBias, jBias int
			switch beam.BeamStrategy {
			case "lookahead":
				iBias = LookaheadPenalty(sc, candidates[i].History, beam.LookaheadWeight)
				jBias = LookaheadPenalty(sc, candidates[j].History, beam.LookaheadWeight)
			case "budget":
				iBias = BudgetPenalty(sc, candidates[i].History, beam.LookaheadWeight)
				jBias = BudgetPenalty(sc, candidates[j].History, beam.LookaheadWeight)
			default: // "none"
				// Pure cumulative penalty, no bias.
			}
			return (candidates[i].CumulativePenalty + iBias) < (candidates[j].CumulativePenalty + jBias)
		})

		retained := beam.BeamWidth
		if retained > len(candidates) {
			retained = len(candidates)
		}

		// Diversity-aware selection: reserve a percentage of slots for underrepresented families.
		if beam.DiversitySlotsPct > 0 && retained > 1 {
			diversitySlots := (retained * beam.DiversitySlotsPct) / 100
			if diversitySlots < 1 {
				diversitySlots = 1
			}
			greedySlots := retained - diversitySlots

			// Take top N greedy (already sorted by penalty + lookahead).
			greedy := candidates[:greedySlots]

			// Track which families are already represented in greedy picks.
			representedFamilies := make(map[int]bool)
			for _, p := range greedy {
				representedFamilies[p.ParentID] = true // Use parentID as proxy for family lineage
			}

			// From remaining candidates, pick best from each unrepresented parent lineage.
			var diversityPicks []BeamPath
			used := make(map[int]bool)
			for i := greedySlots; i < len(candidates) && len(diversityPicks) < diversitySlots; i++ {
				parentFamily := candidates[i].ParentID
				if !representedFamilies[parentFamily] && !used[parentFamily] {
					diversityPicks = append(diversityPicks, candidates[i])
					used[parentFamily] = true
					representedFamilies[parentFamily] = true
				}
			}

			// If we couldn't fill all diversity slots with unique families, fill with next best.
			for i := greedySlots; i < len(candidates) && len(diversityPicks) < diversitySlots; i++ {
				alreadyPicked := false
				for _, dp := range diversityPicks {
					if dp.ID == candidates[i].ID {
						alreadyPicked = true
						break
					}
				}
				if !alreadyPicked {
					diversityPicks = append(diversityPicks, candidates[i])
				}
			}

			// Combine greedy + diversity picks.
			currentPaths = append(greedy, diversityPicks...)
		} else {
			currentPaths = candidates[:retained]
		}

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

	// --- Phase 2: Final window (coupled weeks) ---
	// If finalWindowWeeks > 1, we run the remaining weeks as a coupled block.
	// Each retained path is expanded through ALL final weeks sequentially,
	// but pruning only happens after seeing the combined outcome.
	if finalWindowWeeks > 1 {
		// For each retained path from the normal phase, run all final weeks in sequence
		// and collect coupled candidates ranked by total cumulative penalty.
		var coupledCandidates []BeamPath

		for _, basePath := range currentPaths {
			for _, seed := range beam.Seeds {
				// Run each final week sequentially from this base path.
				chainPath := basePath
				chainValid := true
				var chainWeekPaths []BeamPath

				for fw := 0; fw < finalWindowWeeks; fw++ {
					weekIdx := normalWeeks + fw
					if weekIdx >= numWeeks {
						break
					}

					wd, err := LoadWeekData(weekFiles[weekIdx])
					if err != nil {
						return BeamResult{}, err
					}

					config := baseConfig
					config.Seed = seed
					// Use final window iteration override if set.
					if beam.FinalWindowIter > 0 {
						config.IterationsPerWorker = beam.FinalWindowIter
					}

					var runAudit PFRSAudit
					config.OnAudit = func(a PFRSAudit) {
						runAudit = a
					}

					sol, stats, scoreResult, err := SolveWeekPFRS(sc, wd, chainPath.History, config)
					if err != nil {
						chainValid = false
						break
					}
					if scoreResult.HardViolations != 0 {
						chainValid = false
						break
					}

					newHist := UpdateHistory(sc, chainPath.History, sol)
					weekPath := BeamPath{
						ID:                nextID,
						ParentID:          chainPath.ID,
						Week:              weekIdx + 1,
						CumulativePenalty: chainPath.CumulativePenalty + scoreResult.SoftPenalty,
						WeekPenalty:       scoreResult.SoftPenalty,
						CumulativeSoft:    chainPath.CumulativeSoft + len(scoreResult.SoftDetails),
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
					chainWeekPaths = append(chainWeekPaths, weekPath)
					allPaths = append(allPaths, weekPath)

					if onWeekProgress != nil {
						onWeekProgress(weekIdx+1, weekPath)
					}

					// Chain forward: next week starts from this week's result.
					chainPath = weekPath
				}

				if chainValid && len(chainWeekPaths) == finalWindowWeeks {
					// The final path in the chain represents the full coupled outcome.
					finalPath := chainWeekPaths[len(chainWeekPaths)-1]
					coupledCandidates = append(coupledCandidates, finalPath)
				}
			}
		}

		if len(coupledCandidates) == 0 {
			return BeamResult{AllValid: false}, nil
		}

		// Rank coupled candidates by cumulative penalty, keep best.
		sort.SliceStable(coupledCandidates, func(i, j int) bool {
			return coupledCandidates[i].CumulativePenalty < coupledCandidates[j].CumulativePenalty
		})

		retained := beam.BeamWidth
		if retained > len(coupledCandidates) {
			retained = len(coupledCandidates)
		}
		currentPaths = coupledCandidates[:retained]

		// Add week summaries for each final window week.
		for fw := 0; fw < finalWindowWeeks; fw++ {
			weekIdx := normalWeeks + fw
			weekNum := weekIdx + 1
			// Collect all paths for this week from allPaths.
			var weekCands int
			bestCum := currentPaths[0].CumulativePenalty
			for _, p := range allPaths {
				if p.Week == weekNum {
					weekCands++
				}
			}
			weekSummaries = append(weekSummaries, BeamWeekSummary{
				Week:           weekNum,
				Candidates:     weekCands,
				Retained:       retained,
				BestCumulative: bestCum,
			})
		}
	} else {
		// finalWindowWeeks == 1: the normal loop already processed all weeks.
		// Nothing additional needed.
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
