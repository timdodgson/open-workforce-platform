package inrc2

import "sort"

// --- PFRS Multi-History Beam Search ---
// Pure orchestration over week histories. Does not change PFRS worker behaviour or scoring.

// BeamPath represents one candidate history path through the planning horizon.
type BeamPath struct {
	ID                int
	ParentID          int // -1 for root
	Week              int // 1-indexed (last completed week)
	CumulativePenalty int
	WeekPenalty       int
	CumulativeSoft    int
	WeekSoft          int
	History           History     // output history after this week
	Solution          Solution    // solution for this week
	Seed              int64       // seed used for this week's PFRS run
	Valid             bool        // Hard == 0 for this week
	Stats             PFRSStats   // PFRS execution stats for this week's run
	ScoreResult       ScoreResult // official score result
	Audit             PFRSAudit   // audit data from this week's run

	// Diversity metrics — computed after beam pruning.
	Fingerprint   string  // 12-char MD5 hash of roster assignments
	HammingToBest float64 // Hamming distance to best path this week (0.0-1.0)
}

// BeamResult holds the output of a full beam search across all weeks.
type BeamResult struct {
	WinningPath         []BeamPath // one entry per week for the best full-horizon path
	AllPaths            []BeamPath // all candidate paths generated (for audit)
	TotalPenalty        int
	TotalSoft           int
	AllValid            bool
	WeekSummaries       []BeamWeekSummary
	MidHorizonWeek      int                      // 1-indexed checkpoint week used (0 = none)
	MidHorizonSnapshots []MidHorizonPathSnapshot // Phase 0 telemetry at checkpoint
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
	// MidHorizonWeek is the 1-indexed checkpoint for S7/S8 exposure telemetry / selection (0 = auto when weight set).
	MidHorizonWeek int
	// MidHorizonWeight is λ for selection_score = official + λ×projected(S7+S8) at the checkpoint week only.
	// When > 0, replaces BeamStrategy bias at that week. Telemetry is emitted whenever the week resolves.
	MidHorizonWeight float64
	// MidHorizonSecondHalfIter, when > 0, overrides IterationsPerWorker for weeks after the checkpoint
	// if the best retained path's projected S7+S8 exposure is > 0 (Phase 2 adaptive budget).
	MidHorizonSecondHalfIter int
}

// RunBeamSearch executes PFRS with multi-history beam search across all weeks.
// Path CumulativePenalty is always ScoreMultiStage on the path prefix (validator-aligned).
// Within-week workers still use week-local soft for move acceptance; beam selection does not.
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
	// When finalWindowWeeks=1 (no coupling), run ALL weeks normally.
	normalWeeks := numWeeks
	if finalWindowWeeks > 1 {
		normalWeeks = numWeeks - finalWindowWeeks
		if normalWeeks < 0 {
			normalWeeks = 0
			finalWindowWeeks = numWeeks
		}
	}

	// Start with a single root path.
	root := BeamPath{
		ID:       0,
		ParentID: -1,
		Week:     0,
		History:  initialHist,
		Valid:    true,
	}
	currentPaths := []BeamPath{root}

	nextID := 1
	var weekSummaries []BeamWeekSummary
	var allPaths []BeamPath
	midWeek := ResolveMidHorizonWeek(numWeeks, beam.MidHorizonWeek, beam.MidHorizonWeight)
	var midSnapshots []MidHorizonPathSnapshot
	secondHalfBoost := false
	pathByID := map[int]BeamPath{0: root}
	loadedWeeks := make([]WeekData, 0, numWeeks)

	// --- Phase 1: Normal beam search for weeks 1..normalWeeks ---
	for w := 0; w < normalWeeks; w++ {
		wd, err := LoadWeekData(weekFiles[w])
		if err != nil {
			return BeamResult{}, err
		}
		loadedWeeks = append(loadedWeeks, wd)
		for _, p := range currentPaths {
			pathByID[p.ID] = p
		}

		var candidates []BeamPath

		// Expand each retained path with each seed.
		for _, path := range currentPaths {
			for _, seed := range beam.Seeds {
				config := baseConfig
				config.Seed = seed
				weekNum := w + 1
				if secondHalfBoost && beam.MidHorizonSecondHalfIter > 0 && weekNum > midWeek {
					config.IterationsPerWorker = beam.MidHorizonSecondHalfIter
				}

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
					ID:       nextID,
					ParentID: path.ID,
					Week:     w + 1,
					History:  newHist,
					Solution: sol,
					Seed:     seed,
					Valid:    true,
					Stats:    stats,
					Audit:    runAudit,
				}
				applyOfficialBeamPathScore(&candidate, path, pathByID, sc, loadedWeeks, initialHist, scoreResult)
				nextID++
				pathByID[candidate.ID] = candidate
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

		// Rank by official MultiStage cumulative + strategy bias, keep top beamWidth.
		weekNum := w + 1
		sort.SliceStable(candidates, func(i, j int) bool {
			iBias := beamPathRankBias(sc, candidates[i].History, beam, weekNum, midWeek)
			jBias := beamPathRankBias(sc, candidates[j].History, beam, weekNum, midWeek)
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
			Week:           weekNum,
			Candidates:     len(candidates),
			Retained:       retained,
			BestCumulative: currentPaths[0].CumulativePenalty,
		})

		// Phase 0: capture mid-horizon exposure snapshots after prune.
		if midWeek > 0 && weekNum == midWeek {
			retainedIDs := make(map[int]bool, len(currentPaths))
			for _, p := range currentPaths {
				retainedIDs[p.ID] = true
			}
			for _, c := range candidates {
				midSnapshots = append(midSnapshots, BuildMidHorizonPathSnapshot(
					c, sc, beam.MidHorizonWeight, retainedIDs[c.ID], false))
			}
			// Phase 2: enable second-half iter boost when best retained path has projected exposure.
			if beam.MidHorizonSecondHalfIter > 0 && len(currentPaths) > 0 {
				bestExp := EvaluateMidHorizon(sc, currentPaths[0].History)
				if bestExp.ProjectedTotal() > 0 {
					secondHalfBoost = true
				}
			}
		}
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
			pathByID[basePath.ID] = basePath
			for _, seed := range beam.Seeds {
				// Run each final week sequentially from this base path.
				chainPath := basePath
				chainValid := true
				var chainWeekPaths []BeamPath
				fwWeeks := append([]WeekData(nil), loadedWeeks...)

				for fw := 0; fw < finalWindowWeeks; fw++ {
					weekIdx := normalWeeks + fw
					if weekIdx >= numWeeks {
						break
					}

					wd, err := LoadWeekData(weekFiles[weekIdx])
					if err != nil {
						return BeamResult{}, err
					}
					fwWeeks = append(fwWeeks, wd)

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
						ID:       nextID,
						ParentID: chainPath.ID,
						Week:     weekIdx + 1,
						History:  newHist,
						Solution: sol,
						Seed:     seed,
						Valid:    true,
						Stats:    stats,
						Audit:    runAudit,
					}
					applyOfficialBeamPathScore(&weekPath, chainPath, pathByID, sc, fwWeeks, initialHist, scoreResult)
					nextID++
					pathByID[weekPath.ID] = weekPath
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

		// Rank coupled candidates by official MultiStage cumulative, keep best.
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

	// Mark mid-horizon snapshots that lie on the winning lineage.
	winningIDs := make(map[int]bool, len(winningLineage))
	for _, wp := range winningLineage {
		winningIDs[wp.ID] = true
	}
	for i := range midSnapshots {
		if winningIDs[midSnapshots[i].PathID] {
			midSnapshots[i].Winning = true
		}
	}

	result := BeamResult{
		WinningPath:         winningLineage,
		AllPaths:            allPaths,
		TotalPenalty:        best.CumulativePenalty,
		TotalSoft:           best.CumulativeSoft,
		AllValid:            true,
		WeekSummaries:       weekSummaries,
		MidHorizonWeek:      midWeek,
		MidHorizonSnapshots: midSnapshots,
	}

	return result, nil
}

// lineageSolutions returns solutions from root→parent for a beam node (excludes leaf).
func lineageSolutions(leaf BeamPath, byID map[int]BeamPath) []Solution {
	var chain []BeamPath
	cur := leaf
	for cur.Week > 0 {
		chain = append(chain, cur)
		if cur.ParentID < 0 {
			break
		}
		parent, ok := byID[cur.ParentID]
		if !ok {
			break
		}
		cur = parent
	}
	sols := make([]Solution, 0, len(chain))
	for i := len(chain) - 1; i >= 0; i-- {
		sols = append(sols, chain[i].Solution)
	}
	return sols
}

// applyOfficialBeamPathScore sets CumulativePenalty from ScoreMultiStage on the path prefix.
func applyOfficialBeamPathScore(cand *BeamPath, parent BeamPath, byID map[int]BeamPath, sc Scenario, weeks []WeekData, initialHist History, weekScore ScoreResult) {
	sols := append(lineageSolutions(parent, byID), cand.Solution)
	ms := ScoreMultiStage(sc, weeks, initialHist, sols)
	cand.CumulativePenalty = ms.TotalObjective
	cand.WeekPenalty = ms.TotalObjective - parent.CumulativePenalty
	if cand.WeekPenalty < 0 {
		cand.WeekPenalty = 0
	}
	cand.CumulativeSoft = len(ms.SoftDetails)
	cand.WeekSoft = len(weekScore.SoftDetails)
	cand.ScoreResult = weekScore
}

// beamPathRankBias returns an optional ranking bias on top of official MultiStage cumulative.
// At the mid-horizon checkpoint with MidHorizonWeight > 0, projected S7/S8 replaces
// the normal BeamStrategy bias for that week only.
func beamPathRankBias(sc Scenario, hist History, beam BeamConfig, week, midWeek int) int {
	if midWeek > 0 && week == midWeek && beam.MidHorizonWeight > 0 {
		return MidHorizonSelectionBias(sc, hist, week, midWeek, beam.MidHorizonWeight)
	}
	switch beam.BeamStrategy {
	case "lookahead":
		return LookaheadPenalty(sc, hist, beam.LookaheadWeight)
	case "budget":
		return BudgetPenalty(sc, hist, beam.LookaheadWeight)
	default:
		return 0
	}
}
