package inrc2

import (
	"path/filepath"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2/legacysearch"
)

// TuneSweepRunParams configures a standard (non-beam) PFRS tuning sweep.
type TuneSweepRunParams struct {
	Scenario       Scenario
	WeekFiles      []string
	NumWeeks         int
	History        History
	Options        TuneOptions
	Grid           []TuningGridEntry
	WorkerSI       WorkerIntelligenceWire
	Hooks          TuningSweepHooks
	OnPFRSProgress func(p PFRSProgress, seed int64, week, numWeeks int)
}

// RunTuneSweep executes a multi-grid, multi-seed PFRS tuning sweep from domain options.
func RunTuneSweep(p TuneSweepRunParams) TuningSweepResult {
	algProfile, _ := legacysearch.GetProfile("research")
	progressMs := int64(0)
	if p.Options.ProgressEnabled {
		progressMs = int64(p.Options.ProgressIntervalSec) * 1000
	}

	return RunTuningSweep(TuningSweepParams{
		Scenario:   p.Scenario,
		WeekFiles:  p.WeekFiles,
		NumWeeks:   p.NumWeeks,
		History:    p.History,
		Grid:       p.Grid,
		Seeds:      p.Options.Seeds,
		AlgProfile: algProfile,
		BuildConfig: func(entry TuningGridEntry, seed int64, currentWeek *int) PFRSConfig {
			var progressCb ProgressFunc
			if p.Options.ProgressEnabled && p.OnPFRSProgress != nil {
				week := *currentWeek
				progressCb = func(prog PFRSProgress) {
					p.OnPFRSProgress(prog, seed, week+1, p.NumWeeks)
				}
			}
			return BuildTunePFRSConfig(TuneConfigParams{
				Entry:              entry,
				Mode:               p.Options.WorkerMode,
				Portfolio:          p.Options.Portfolio,
				MaxConcurrent:      p.Options.MaxConcurrent,
				CoolingMode:        p.Options.CoolingMode,
				Seed:               seed,
				LAHCBufferLength:   p.Options.LAHCBufferLength,
				DecisionEngine:     p.WorkerSI.Engine,
				DecisionRecorder:   p.WorkerSI.DecisionRecorder,
				AssistMode:         p.WorkerSI.AssistMode,
				AssistRecorder:     p.WorkerSI.AssistRecorder,
				OnProgress:         progressCb,
				ProgressIntervalMs: progressMs,
			})
		},
		Hooks: p.Hooks,
	})
}

// TuneSweepFinalizeResult holds artifact finalization output for a sweep run.
type TuneSweepFinalizeResult struct {
	AuditRowCount      int
	LearningRecordCount int
}

// FinalizeTuneSweep writes standard sweep artifacts when an audit path is configured.
func FinalizeTuneSweep(opts TuneOptions, sweep TuningSweepResult, bestPenalty int) (TuneSweepFinalizeResult, error) {
	var out TuneSweepFinalizeResult
	if opts.AuditCSVPath == "" || len(sweep.AuditRows) == 0 {
		return out, nil
	}

	outputDir := filepath.Dir(opts.AuditCSVPath)

	totalRecords, err := FinalizeStandardArtifacts(StandardArtifactsParams{
		OutputDir:    outputDir,
		AuditCSVPath: opts.AuditCSVPath,
		AuditRows:    sweep.AuditRows,
		RunJSON: PFRSStandardRunJSONParams{
			InstanceName: opts.InstanceName,
			WorkerMode:   opts.WorkerMode,
			BestPenalty:  bestPenalty,
			RunLabel:     opts.RunLabel,
		},
		LearningCfg: NRPLearningConfig{
			Instance:            opts.InstanceName,
			RunSeed:             opts.Seeds[0],
			Temperature:         opts.OverrideTemp,
			LAHCLength:          opts.LAHCBufferLength,
			IterationsPerWorker: opts.OverrideIter,
		},
		Bundles: sweep.Bundles,
	})
	if err != nil {
		return out, err
	}
	out.AuditRowCount = len(sweep.AuditRows)
	out.LearningRecordCount = totalRecords
	return out, nil
}
