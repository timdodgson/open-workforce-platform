package inrc2

// BuildBeamWeekAuditBundles builds worker-learning bundles from a beam winning path.
func BuildBeamWeekAuditBundles(winningPath []BeamPath) []WeekAuditBundle {
	bundles := make([]WeekAuditBundle, 0, len(winningPath))
	for weekIdx, wp := range winningPath {
		if len(wp.Audit.Workers) == 0 {
			continue
		}
		bundles = append(bundles, WeekAuditBundle{
			Week:                weekIdx + 1,
			GlobalBestAtSpawn:   Worker0StartPenalty(wp.Audit),
			TotalWorkersStarted: len(wp.Audit.Workers),
			ActiveFamilies:      1,
			Workers:             wp.Audit.Workers,
		})
	}
	return bundles
}

// BeamArtifactsParams configures PFRS beam run artifact output.
type BeamArtifactsParams struct {
	OutputDir    string
	AuditCSVPath string
	ScenarioID   string
	Config       PFRSConfig
	WinningPath  []BeamPath
	RunJSON      PFRSBeamRunJSONParams
	LearningCfg  NRPLearningConfig
}

// FinalizeBeamArtifacts writes run.json, audit CSV, and worker_learning.csv for a beam run.
func FinalizeBeamArtifacts(p BeamArtifactsParams) error {
	if err := WritePFRSBeamRunJSON(p.OutputDir, p.RunJSON); err != nil {
		return err
	}

	if p.AuditCSVPath != "" && len(p.WinningPath) > 0 {
		rows := make([]WeekAuditRow, 0, len(p.WinningPath))
		for _, wp := range p.WinningPath {
			row := BuildWeekAuditRow(p.ScenarioID, p.Config, wp.Week, Worker0StartPenalty(wp.Audit), wp.Stats, wp.ScoreResult, wp.Audit)
			row.Seed = wp.Seed
			rows = append(rows, row)
		}
		if err := WriteAuditCSV(p.AuditCSVPath, rows); err != nil {
			return err
		}
	}

	bundles := BuildBeamWeekAuditBundles(p.WinningPath)
	if len(bundles) > 0 {
		if err := EmitNRPWorkerLearning(p.OutputDir, p.LearningCfg, bundles); err != nil {
			return err
		}
	}
	return nil
}
