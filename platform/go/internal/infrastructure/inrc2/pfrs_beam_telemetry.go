package inrc2

import (
	"path/filepath"
)

// BeamWinningPathTelemetryParams configures CSV telemetry export for a beam winning path.
type BeamWinningPathTelemetryParams struct {
	TelemetryDir string
	RunCtx       RunContext
	Config       PFRSConfig
	WinningPath  []BeamPath
	Scenario     Scenario
	BeamResult   BeamResult
}

// BeamWinningPathTelemetrySummary counts rows written to each telemetry CSV.
type BeamWinningPathTelemetrySummary struct {
	PlateauEvents       int
	WorkerRows          int
	ImprovementRows     int
	BranchRows          int
	DiversityRows       int
	DiscoveryRows       int
	MidHorizonRows      int
	MidHorizonNurseRows int
}

// WriteBeamWinningPathTelemetry writes plateau, worker, improvement, branch, diversity,
// and discovery CSVs for the beam winning path.
func WriteBeamWinningPathTelemetry(p BeamWinningPathTelemetryParams) (BeamWinningPathTelemetrySummary, error) {
	var summary BeamWinningPathTelemetrySummary
	if len(p.WinningPath) == 0 {
		return summary, nil
	}

	dir := p.TelemetryDir
	if dir == "" {
		return summary, nil
	}

	var allPlateaus []PlateauEvent
	for weekIdx, wp := range p.WinningPath {
		for i := range wp.Audit.Plateaus {
			wp.Audit.Plateaus[i].Week = weekIdx + 1
		}
		allPlateaus = append(allPlateaus, wp.Audit.Plateaus...)
	}
	if len(allPlateaus) > 0 {
		plateauPath := filepath.Join(dir, "plateaus.csv")
		durationMs := p.WinningPath[0].Stats.DurationMs
		if err := WritePlateauCSV(plateauPath, p.RunCtx, allPlateaus, p.Config.IterationsPerWorker, durationMs); err != nil {
			return summary, err
		}
		summary.PlateauEvents = len(allPlateaus)
	}

	var allWorkerRows []WorkerLifecycleRow
	var allImprovementRows []ImprovementRow
	var allBranchRows []BranchRow
	var allDiscoveryRows []DiscoveryRow

	for weekIdx, wp := range p.WinningPath {
		branchCounts := make(map[int]int)
		for _, bu := range wp.Audit.BestUpdates {
			branchCounts[bu.WorkerID]++
		}

		depthMap := workerDepthMap(wp.Audit.Workers)

		rows := BuildWorkerLifecycleRows(p.RunCtx, wp.Audit.Workers, weekIdx+1, wp.Seed,
			p.Config.InitialTemperature, branchCounts, depthMap)
		allWorkerRows = append(allWorkerRows, rows...)

		impRows := BuildImprovementRows(p.RunCtx, weekIdx+1, wp.Audit.BestUpdates, p.Config.EffectiveCoolingRate())
		allImprovementRows = append(allImprovementRows, impRows...)

		parentMap := make(map[int]int)
		for _, w := range wp.Audit.Workers {
			parentMap[w.WorkerID] = w.ParentWorkerID
		}
		branchRows := BuildBranchRows(p.RunCtx, weekIdx+1, wp.Audit.BestUpdates,
			p.Config.EffectiveCoolingRate(), depthMap, parentMap)
		allBranchRows = append(allBranchRows, branchRows...)

		discRows := BuildDiscoveryRows(p.RunCtx, weekIdx+1, wp.ID, wp.Seed,
			wp.Audit.Discoveries, depthMap, wp.Audit.WinningWorkerID)
		allDiscoveryRows = append(allDiscoveryRows, discRows...)
	}

	if len(allWorkerRows) > 0 {
		if err := WriteWorkerLifecycleCSV(filepath.Join(dir, "workers.csv"), allWorkerRows); err != nil {
			return summary, err
		}
		summary.WorkerRows = len(allWorkerRows)
	}
	if len(allImprovementRows) > 0 {
		if err := WriteImprovementsCSV(filepath.Join(dir, "improvements.csv"), allImprovementRows); err != nil {
			return summary, err
		}
		summary.ImprovementRows = len(allImprovementRows)
	}
	if len(allBranchRows) > 0 {
		if err := WriteBranchCSV(filepath.Join(dir, "branches.csv"), allBranchRows); err != nil {
			return summary, err
		}
		summary.BranchRows = len(allBranchRows)
	}

	diversityRows := BuildDiversityRows(p.RunCtx, p.BeamResult, p.Scenario)
	if len(diversityRows) > 0 {
		if err := WriteDiversityCSV(filepath.Join(dir, "diversity.csv"), diversityRows); err != nil {
			return summary, err
		}
		summary.DiversityRows = len(diversityRows)
	}
	if len(allDiscoveryRows) > 0 {
		if err := WriteDiscoveriesCSV(filepath.Join(dir, "discoveries.csv"), allDiscoveryRows); err != nil {
			return summary, err
		}
		summary.DiscoveryRows = len(allDiscoveryRows)
	}

	if len(p.BeamResult.MidHorizonSnapshots) > 0 {
		if err := WriteMidHorizonCSV(filepath.Join(dir, "mid_horizon.csv"), p.BeamResult.MidHorizonSnapshots); err != nil {
			return summary, err
		}
		summary.MidHorizonRows = len(p.BeamResult.MidHorizonSnapshots)
		if err := WriteMidHorizonNurseCSV(filepath.Join(dir, "mid_horizon_nurses.csv"), p.BeamResult.MidHorizonSnapshots); err != nil {
			return summary, err
		}
		nurseRows := 0
		for _, s := range p.BeamResult.MidHorizonSnapshots {
			nurseRows += len(s.Nurses)
		}
		summary.MidHorizonNurseRows = nurseRows
	}

	return summary, nil
}

func workerDepthMap(workers []WorkerAudit) map[int]int {
	depthMap := make(map[int]int)
	for _, w := range workers {
		depth := 0
		pid := w.ParentWorkerID
		for pid >= 0 {
			depth++
			found := false
			for _, w2 := range workers {
				if w2.WorkerID == pid {
					pid = w2.ParentWorkerID
					found = true
					break
				}
			}
			if !found {
				break
			}
		}
		depthMap[w.WorkerID] = depth
	}
	return depthMap
}
