package inrc2

import (
	"fmt"
	"path/filepath"
	"time"
)

// TuneBeamHooks provides optional progress and status callbacks during beam tuning.
type TuneBeamHooks struct {
	OnPFRSProgress         func(p PFRSProgress)
	OnBeamWeek             func(week int, path BeamPath)
	OnRefinementBefore     func(penalty, violations int)
	OnRefinementConfig     func(mode string, iter int, temp float64)
	OnRefinementAfter      func(prePenalty, postPenalty, preViolations, postViolations int, summary RefinementSummary)
	OnFinalValidationStart func()
	OnFinalWeekLine        func(week, penalty, softCount, hardViolations int)
	OnFinalPenalty         func(penalty, totalViolations int)
	OnTelemetrySummary     func(dir string, s BeamWinningPathTelemetrySummary)
	OnArtifactMessage      func(msg string)
	OnError                func(msg string)
}

// TuneBeamParams configures a single-config PFRS beam tuning run.
type TuneBeamParams struct {
	Options   TuneOptions
	Scenario  Scenario
	WeekFiles []string
	NumWeeks  int
	History   History
	Grid      []TuningGridEntry
	WorkerSI  WorkerIntelligenceWire
	Hooks     TuneBeamHooks
}

// TuneBeamResult holds beam tuning output for CLI display and artifact upload.
type TuneBeamResult struct {
	BeamResult         BeamResult
	BaseConfig         PFRSConfig
	EffectiveBeamSeeds []int64
	TelemetrySummary   BeamWinningPathTelemetrySummary
	TelemetryDir       string
	OutputDir          string
}

// RunTuneBeam executes beam search, optional refinement, validation, and artifact finalization.
func RunTuneBeam(p TuneBeamParams) (TuneBeamResult, error) {
	entry := p.Grid[0]
	effectiveBeamSeeds := p.Options.BeamSeeds
	if len(effectiveBeamSeeds) == 0 {
		effectiveBeamSeeds = p.Options.Seeds
	}

	weekSlice := p.WeekFiles[:p.NumWeeks]

	baseConfig := BuildTunePFRSConfig(TuneConfigParams{
		Entry:             entry,
		Mode:              p.Options.WorkerMode,
		Portfolio:         p.Options.Portfolio,
		MaxConcurrent:     p.Options.MaxConcurrent,
		CoolingMode:       p.Options.CoolingMode,
		LAHCBufferLength:  p.Options.LAHCBufferLength,
		ReheatThreshold:   p.Options.ReheatThreshold,
		ReheatFactor:      p.Options.ReheatFactor,
		ReheatMinFraction: p.Options.ReheatMinFraction,
		NoReheat:          p.Options.NoReheat,
		Beam:              true,
		DecisionEngine:    p.WorkerSI.Engine,
		DecisionRecorder:  p.WorkerSI.DecisionRecorder,
		AssistMode:        p.WorkerSI.AssistMode,
		AssistRecorder:    p.WorkerSI.AssistRecorder,
	})

	if p.Options.ProgressEnabled {
		baseConfig.ProgressIntervalMs = int64(p.Options.ProgressIntervalSec) * 1000
		if p.Hooks.OnPFRSProgress != nil {
			baseConfig.OnProgress = p.Hooks.OnPFRSProgress
		}
	}

	beam := BeamConfig{
		BeamWidth:                p.Options.BeamWidth,
		Seeds:                    effectiveBeamSeeds,
		FinalWindowWeeks:         p.Options.FinalWindowWeeks,
		FinalWindowIter:          p.Options.FinalWindowIter,
		LookaheadWeight:          p.Options.LookaheadWeight,
		DiversitySlotsPct:        p.Options.DiversitySlotsPct,
		BeamStrategy:             p.Options.BeamStrategy,
		MidHorizonWeek:           p.Options.MidHorizonWeek,
		MidHorizonWeight:         p.Options.MidHorizonWeight,
		MidHorizonSecondHalfIter: p.Options.MidHorizonSecondHalfIter,
	}

	onProgress := p.Hooks.OnBeamWeek
	beamResult, err := RunBeamSearch(p.Scenario, weekSlice, p.History, baseConfig, beam, onProgress)
	if err != nil {
		if p.Hooks.OnError != nil {
			p.Hooks.OnError(fmt.Sprintf("Beam search failed: %v", err))
		}
		return TuneBeamResult{}, err
	}

	if p.Options.TreeCSVPath != "" {
		if err := WriteBeamTreeCSV(p.Options.TreeCSVPath, beamResult); err != nil {
			if p.Hooks.OnError != nil {
				p.Hooks.OnError(fmt.Sprintf("Error writing tree CSV: %v", err))
			}
		} else if p.Hooks.OnArtifactMessage != nil {
			p.Hooks.OnArtifactMessage(fmt.Sprintf("Tree CSV written: %s (%d paths)", p.Options.TreeCSVPath, len(beamResult.AllPaths)))
		}
	}

	runCtx := RunContext{
		RunID:       fmt.Sprintf("%s-%d", p.Scenario.ID, baseConfig.Seed),
		Instance:    p.Scenario.ID,
		Seed:        baseConfig.Seed,
		BeamWidth:   p.Options.BeamWidth,
		Iterations:  baseConfig.IterationsPerWorker,
		Temperature: baseConfig.InitialTemperature,
		CoolingMode: baseConfig.CoolingMode,
		Timestamp:   time.Now().Format(time.RFC3339),
	}

	telemetryDir := filepath.Dir(p.Options.AuditCSVPath)
	telSummary, err := WriteBeamWinningPathTelemetry(BeamWinningPathTelemetryParams{
		TelemetryDir: telemetryDir,
		RunCtx:       runCtx,
		Config:       baseConfig,
		WinningPath:  beamResult.WinningPath,
		Scenario:     p.Scenario,
		BeamResult:   beamResult,
	})
	if err != nil {
		if p.Hooks.OnError != nil {
			p.Hooks.OnError(fmt.Sprintf("Error writing beam telemetry CSVs: %v", err))
		}
	} else if p.Hooks.OnTelemetrySummary != nil {
		p.Hooks.OnTelemetrySummary(telemetryDir, telSummary)
	}

	if len(beamResult.WinningPath) > 0 {
		if p.Options.RefinementMode != "none" {
			preRefPenalty, preRefViolations := OfficialValidateBeamPath(p.Scenario, weekSlice, beamResult.WinningPath, p.History)
			if p.Hooks.OnRefinementBefore != nil {
				p.Hooks.OnRefinementBefore(preRefPenalty, preRefViolations)
			}
			if p.Hooks.OnRefinementConfig != nil {
				p.Hooks.OnRefinementConfig(p.Options.RefinementMode, p.Options.RefinementIter, p.Options.RefinementTemp)
			}

			refinedPath, refSummary := Refine(p.Scenario, weekSlice, beamResult.WinningPath, RefinementConfig{
				Mode:               p.Options.RefinementMode,
				Iterations:         p.Options.RefinementIter,
				Seed:               baseConfig.Seed,
				InitialTemperature: p.Options.RefinementTemp,
			}, p.History)

			postRefPenalty, postRefViolations := OfficialValidateBeamPath(p.Scenario, weekSlice, refinedPath, p.History)
			if p.Hooks.OnRefinementAfter != nil {
				p.Hooks.OnRefinementAfter(preRefPenalty, postRefPenalty, preRefViolations, postRefViolations, refSummary)
			}

			beamResult.WinningPath = refinedPath
			if len(refinedPath) > 0 {
				beamResult.TotalPenalty = refinedPath[len(refinedPath)-1].CumulativePenalty
			}
		}

		updatedPath, finalPenalty, totalViolations := OfficialRevalidateBeamPath(p.Scenario, weekSlice, beamResult.WinningPath, p.History)
		beamResult.WinningPath = updatedPath
		beamResult.TotalPenalty = finalPenalty

		if p.Hooks.OnFinalValidationStart != nil {
			p.Hooks.OnFinalValidationStart()
		}
		for _, wp := range beamResult.WinningPath {
			if p.Hooks.OnFinalWeekLine != nil {
				p.Hooks.OnFinalWeekLine(wp.Week, wp.WeekPenalty, len(wp.ScoreResult.SoftDetails), wp.ScoreResult.HardViolations)
			}
		}
		if p.Hooks.OnFinalPenalty != nil {
			p.Hooks.OnFinalPenalty(finalPenalty, totalViolations)
		}

		rosterPath := filepath.Join(telemetryDir, "roster.json")
		if err := WriteRosterJSON(rosterPath, p.Scenario, beamResult.WinningPath); err != nil {
			if p.Hooks.OnError != nil {
				p.Hooks.OnError(fmt.Sprintf("Error writing roster JSON: %v", err))
			}
		} else if p.Hooks.OnArtifactMessage != nil {
			p.Hooks.OnArtifactMessage(fmt.Sprintf("Roster JSON written: %s", rosterPath))
		}
	}

	outputDir := filepath.Dir(p.Options.AuditCSVPath)
	if err := FinalizeBeamArtifacts(BeamArtifactsParams{
		OutputDir:    outputDir,
		AuditCSVPath: p.Options.AuditCSVPath,
		ScenarioID:   p.Scenario.ID,
		Config:       baseConfig,
		WinningPath:  beamResult.WinningPath,
		RunJSON: PFRSBeamRunJSONParams{
			InstanceID: p.Scenario.ID, Mode: baseConfig.Mode,
			IterationsPerWorker: baseConfig.IterationsPerWorker,
			InitialTemperature:  baseConfig.InitialTemperature, CoolingMode: baseConfig.CoolingMode,
			EffectiveCoolingRate: baseConfig.EffectiveCoolingRate(),
			LateAcceptanceLength: baseConfig.LateAcceptanceLength,
			BeamWidth:            p.Options.BeamWidth, BeamSeeds: effectiveBeamSeeds, Seed: baseConfig.Seed,
			MaxTotalWorkers: baseConfig.MaxTotalWorkers,
			LookaheadWeight: p.Options.LookaheadWeight, FinalWindowWeeks: p.Options.FinalWindowWeeks,
			FinalWindowIter: p.Options.FinalWindowIter, BeamStrategy: p.Options.BeamStrategy,
			DiversitySlotsPct: p.Options.DiversitySlotsPct,
			MidHorizonWeek:    p.Options.MidHorizonWeek, MidHorizonWeight: p.Options.MidHorizonWeight,
			Portfolio: p.Options.Portfolio, RunLabel: p.Options.RunLabel,
		},
		LearningCfg: NRPLearningConfig{
			Instance:            p.Scenario.ID,
			RunSeed:             baseConfig.Seed,
			Temperature:         baseConfig.InitialTemperature,
			LAHCLength:          baseConfig.LateAcceptanceLength,
			TabuTenure:          0,
			IterationsPerWorker: baseConfig.IterationsPerWorker,
		},
	}); err != nil {
		if p.Hooks.OnError != nil {
			p.Hooks.OnError(fmt.Sprintf("Error finalizing beam artifacts: %v", err))
		}
	}

	return TuneBeamResult{
		BeamResult:         beamResult,
		BaseConfig:         baseConfig,
		EffectiveBeamSeeds: effectiveBeamSeeds,
		TelemetrySummary:   telSummary,
		TelemetryDir:       telemetryDir,
		OutputDir:          outputDir,
	}, nil
}
