package inrc2

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// PFRSBeamRunJSONParams are the dashboard-facing fields for a PFRS beam run.json.
// This is hand-formatted to preserve the exact field layout expected by the UI.
type PFRSBeamRunJSONParams struct {
	InstanceID           string
	Mode                 string
	IterationsPerWorker  int
	InitialTemperature   float64
	CoolingMode          string
	EffectiveCoolingRate float64
	LateAcceptanceLength int
	BeamWidth            int
	BeamSeeds            []int64
	Seed                 int64
	MaxTotalWorkers      int
	LookaheadWeight      float64
	FinalWindowWeeks     int
	FinalWindowIter      int
	BeamStrategy         string
	DiversitySlotsPct    int
	MidHorizonWeek       int
	MidHorizonWeight     float64
	Portfolio            []string
	RunLabel             string
}

func formatPFRSBeamRunJSON(p PFRSBeamRunJSONParams) string {
	seedParts := make([]string, len(p.BeamSeeds))
	for i, s := range p.BeamSeeds {
		seedParts[i] = fmt.Sprintf("%d", s)
	}
	return fmt.Sprintf(`{
  "instance": %q,
  "algorithm": "parallel-feasible-roster-search",
  "mode": %q,
  "iterationsPerWorker": %d,
  "initialTemperature": %.1f,
  "coolingMode": %q,
  "effectiveCoolingRate": %.10f,
  "lateAcceptanceLength": %d,
  "beamWidth": %d,
  "beamSeeds": [%s],
  "seed": %d,
  "cpus": %d,
  "maxTotalWorkers": %d,
  "lookaheadWeight": %.2f,
  "finalWindowWeeks": %d,
  "finalWindowIterations": %d,
  "beamStrategy": %q,
  "diversitySlotsPct": %d,
  "midHorizonWeek": %d,
  "midHorizonWeight": %.2f,
  "portfolio": %q,
  "runLabel": %q
}`, p.InstanceID, p.Mode, p.IterationsPerWorker,
		p.InitialTemperature, p.CoolingMode,
		p.EffectiveCoolingRate, p.LateAcceptanceLength,
		p.BeamWidth,
		strings.Join(seedParts, ", "),
		p.Seed, runtime.NumCPU(), p.MaxTotalWorkers,
		p.LookaheadWeight, p.FinalWindowWeeks, p.FinalWindowIter, p.BeamStrategy, p.DiversitySlotsPct,
		p.MidHorizonWeek, p.MidHorizonWeight,
		strings.Join(p.Portfolio, ","), p.RunLabel)
}

// WritePFRSBeamRunJSON writes run.json for PFRS beam runs.
func WritePFRSBeamRunJSON(outputDir string, p PFRSBeamRunJSONParams) error {
	path := filepath.Join(outputDir, "run.json")
	return os.WriteFile(path, []byte(formatPFRSBeamRunJSON(p)), 0644)
}

// PFRSStandardRunJSONParams are the dashboard-facing fields for standard NRP runs.
// This is hand-formatted to preserve the exact field layout expected by the UI.
type PFRSStandardRunJSONParams struct {
	InstanceName string
	WorkerMode   string
	BestPenalty  int
	RunLabel     string
}

func formatPFRSStandardRunJSON(p PFRSStandardRunJSONParams) string {
	return fmt.Sprintf(`{
  "instance": %q,
  "problemType": "nrp",
  "mode": %q,
  "bestObjective": %d,
  "totalPenalty": %d,
  "runLabel": %q
}`, p.InstanceName, p.WorkerMode, p.BestPenalty, p.BestPenalty, p.RunLabel)
}

// WritePFRSStandardRunJSON writes run.json for standard (non-beam) runs.
func WritePFRSStandardRunJSON(outputDir string, p PFRSStandardRunJSONParams) error {
	path := filepath.Join(outputDir, "run.json")
	return os.WriteFile(path, []byte(formatPFRSStandardRunJSON(p)), 0644)
}
