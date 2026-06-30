package inrc2

// TuningGridEntry represents one parameter combination to evaluate.
type TuningGridEntry struct {
	IterationsPerWorker int
	MaxTotalWorkers     int
	InitialTemperature  float64
	CoolingRate         float64
}

// TuningResult holds the outcome of running one grid entry across all weeks for one seed.
type TuningResult struct {
	Entry        TuningGridEntry
	Seed         int64
	TotalPenalty int
	TotalHard    int
	TotalSoft    int
	TotalAssign  int
	TotalMs      int64
	TotalCands   int
	Valid        bool // Hard == 0
}

// MultiSeedResult aggregates results for one grid entry across multiple seeds.
type MultiSeedResult struct {
	Entry      TuningGridEntry
	Seeds      int
	BestPen    int
	AvgPen     int
	WorstPen   int
	BestSeed   int64
	AvgSoft    int
	AvgMs      int64
	TotalCands int
	AllValid   bool // all seeds produced Hard == 0
	ValidCount int  // how many seeds were valid
}

// AggregateSeeds takes per-seed results for a single grid entry and produces a MultiSeedResult.
func AggregateSeeds(entry TuningGridEntry, results []TuningResult) MultiSeedResult {
	ms := MultiSeedResult{
		Entry:    entry,
		Seeds:    len(results),
		BestPen:  int(^uint(0) >> 1), // max int
		WorstPen: 0,
		AllValid: true,
	}

	totalPen := 0
	totalSoft := 0
	totalMs := int64(0)

	for _, r := range results {
		if !r.Valid {
			ms.AllValid = false
		} else {
			ms.ValidCount++
		}

		totalPen += r.TotalPenalty
		totalSoft += r.TotalSoft
		totalMs += r.TotalMs
		ms.TotalCands += r.TotalCands

		if r.TotalPenalty < ms.BestPen {
			ms.BestPen = r.TotalPenalty
			ms.BestSeed = r.Seed
		}
		if r.TotalPenalty > ms.WorstPen {
			ms.WorstPen = r.TotalPenalty
		}
	}

	if len(results) > 0 {
		ms.AvgPen = totalPen / len(results)
		ms.AvgSoft = totalSoft / len(results)
		ms.AvgMs = totalMs / int64(len(results))
	}

	return ms
}

// GenerateGrid produces the cartesian product of parameter values in deterministic order.
func GenerateGrid(iterations []int, workers []int, temperatures []float64, coolingRates []float64) []TuningGridEntry {
	var grid []TuningGridEntry
	for _, iter := range iterations {
		for _, w := range workers {
			for _, temp := range temperatures {
				for _, cr := range coolingRates {
					grid = append(grid, TuningGridEntry{
						IterationsPerWorker: iter,
						MaxTotalWorkers:     w,
						InitialTemperature:  temp,
						CoolingRate:         cr,
					})
				}
			}
		}
	}
	return grid
}

// RankResults sorts results: valid first (by penalty ascending), then invalid (by hard ascending).
func RankResults(results []TuningResult) (valid []TuningResult, invalid []TuningResult) {
	for _, r := range results {
		if r.Valid {
			valid = append(valid, r)
		} else {
			invalid = append(invalid, r)
		}
	}

	// Sort valid by penalty ascending, then runtime ascending for tie-breaking.
	for i := 0; i < len(valid)-1; i++ {
		for j := i + 1; j < len(valid); j++ {
			if valid[j].TotalPenalty < valid[i].TotalPenalty ||
				(valid[j].TotalPenalty == valid[i].TotalPenalty && valid[j].TotalMs < valid[i].TotalMs) {
				valid[i], valid[j] = valid[j], valid[i]
			}
		}
	}

	// Sort invalid by hard ascending, then penalty ascending.
	for i := 0; i < len(invalid)-1; i++ {
		for j := i + 1; j < len(invalid); j++ {
			if invalid[j].TotalHard < invalid[i].TotalHard ||
				(invalid[j].TotalHard == invalid[i].TotalHard && invalid[j].TotalPenalty < invalid[i].TotalPenalty) {
				invalid[i], invalid[j] = invalid[j], invalid[i]
			}
		}
	}

	return valid, invalid
}

// RankMultiSeedResults ranks by average penalty ascending, then average runtime for tie-breaking.
// Only configurations where all seeds are valid are ranked. Others go to invalid.
func RankMultiSeedResults(results []MultiSeedResult) (valid []MultiSeedResult, invalid []MultiSeedResult) {
	for _, r := range results {
		if r.AllValid {
			valid = append(valid, r)
		} else {
			invalid = append(invalid, r)
		}
	}

	// Sort valid by average penalty ascending, then average runtime ascending.
	for i := 0; i < len(valid)-1; i++ {
		for j := i + 1; j < len(valid); j++ {
			if valid[j].AvgPen < valid[i].AvgPen ||
				(valid[j].AvgPen == valid[i].AvgPen && valid[j].AvgMs < valid[i].AvgMs) {
				valid[i], valid[j] = valid[j], valid[i]
			}
		}
	}

	// Sort invalid by valid count descending (most valid first), then average penalty.
	for i := 0; i < len(invalid)-1; i++ {
		for j := i + 1; j < len(invalid); j++ {
			if invalid[j].ValidCount > invalid[i].ValidCount ||
				(invalid[j].ValidCount == invalid[i].ValidCount && invalid[j].AvgPen < invalid[i].AvgPen) {
				invalid[i], invalid[j] = invalid[j], invalid[i]
			}
		}
	}

	return valid, invalid
}
