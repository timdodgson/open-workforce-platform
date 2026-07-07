package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

// --- Low-level file writers (preserve 0644 and error handling patterns) ---

func writeTelemetryBytes(path string, data []byte) {
	os.WriteFile(path, data, 0644)
}

func writeTelemetryBytesErr(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}

func writeSolutionJSON(outputDir string, data []byte) {
	writeTelemetryBytes(filepath.Join(outputDir, "solution.json"), data)
}

func writeTelemetryFile(outputDir, filename string, data []byte) {
	writeTelemetryBytes(filepath.Join(outputDir, filename), data)
}

// --- CVRP audit CSV (dashboard-compatible results.csv) ---

const cvrpResultsCSVHeader = "instance,seed,mode,iterationsPerWorker,maxTotalWorkers,maxConcurrent,initialTemperature,coolingRate,coolingMode,effectiveCoolingRate,minTemperature,lateAcceptanceLen,week,startPenalty,finalPenalty,improvement,hardViolations,softViolations,candidates,accepted,rejected,acceptanceRate,bestIteration,bestWorkerID,workersStarted,branchesCreated,branchesDropped,maxQueueDepth,maxConcurrentSeen,durationMs,saFinalTemp,saTempAtBest,saAcceptedBetter,saAcceptedWorse,saRejectedByProb,lahcAcceptedByCurrent,lahcAcceptedByLate,lahcRejectedByLate,branchesQueued,branchesStarted2,branchesCompleted,winningBranchDepth,workersImproved,workersProducedBest,rejectedNoop,rejectedSkill,rejectedSuccession,rejectedHistory\n"

type cvrpResultsCSVParams struct {
	Instance     string
	Seed         int64
	WinnerMode   string
	Iterations   int
	Temperature  float64
	SearchResult optimisation.SearchResult
}

func buildCVRPResultsCSV(p cvrpResultsCSVParams) []byte {
	acceptRate := 0.0
	if p.SearchResult.Candidates > 0 {
		acceptRate = float64(p.SearchResult.Accepted) / float64(p.SearchResult.Candidates) * 100
	}
	row := fmt.Sprintf("%s,%d,%s,%d,1,1,%.1f,0,adaptive,0,0.0001,0,1,%d,%d,%d,0,0,%d,%d,%d,%.1f,0,0,1,0,0,0,1,%d,0,0,%d,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0\n",
		p.Instance, p.Seed, p.WinnerMode, p.Iterations,
		p.Temperature,
		p.SearchResult.InitialPenalty, p.SearchResult.BestPenalty, p.SearchResult.InitialPenalty-p.SearchResult.BestPenalty,
		p.SearchResult.Candidates, p.SearchResult.Accepted, p.SearchResult.Rejected, acceptRate,
		p.SearchResult.DurationMs, p.SearchResult.Accepted)
	return []byte(cvrpResultsCSVHeader + row)
}

// --- CVRP NRP-format discoveries.csv ---

const cvrpDiscoveriesCSVHeader = "run_id,instance,seed,beam_width,iterations,temperature,cooling_mode,timestamp,week,worker_id,beam_path,candidate,elapsed_ms,temperature_at_event,current_penalty,previous_best,new_best,improvement,improvement_pct,event_type,branch_depth,seed_used,accepted_worse_count,hard_reject_count,soft_reject_count,discovery_number,cands_since_previous,time_since_previous_ms,improvement_per_10k,improvement_per_second,post_reheat_improved,post_reheat_best_delta,post_reheat_cands_to_improve,post_reheat_spawned_branch,post_reheat_beat_global,post_reheat_on_winning_lineage\n"

type cvrpDiscoveriesCSVParams struct {
	RunLabel    string
	Instance    string
	Seed        int64
	Iterations  int
	Temperature float64
	Discoveries []optimisation.Discovery
}

func buildCVRPDiscoveriesCSV(p cvrpDiscoveriesCSVParams) []byte {
	var b strings.Builder
	b.WriteString(cvrpDiscoveriesCSVHeader)
	prevCandidate := 0
	prevElapsed := int64(0)
	for i, d := range p.Discoveries {
		candsSince := d.Candidate - prevCandidate
		timeSince := d.ElapsedMs - prevElapsed
		impPct := 0.0
		if d.OldBest > 0 {
			impPct = float64(d.Improvement) / float64(d.OldBest) * 100
		}
		impPer10K := 0.0
		if candsSince > 0 {
			impPer10K = float64(d.Improvement) / float64(candsSince) * 10000
		}
		impPerSec := 0.0
		if timeSince > 0 {
			impPerSec = float64(d.Improvement) / (float64(timeSince) / 1000)
		}
		b.WriteString(fmt.Sprintf("%s,%s,%d,1,%d,%.1f,adaptive,,%d,0,0,%d,%d,0,%d,%d,%d,%d,%.2f,GLOBAL_BEST,0,%d,0,0,0,%d,%d,%d,%.2f,%.2f,0,0,0,0,0,0\n",
			p.RunLabel, p.Instance, p.Seed, p.Iterations, p.Temperature,
			1, d.Candidate, d.ElapsedMs,
			d.NewBest, d.OldBest, d.NewBest, d.Improvement, impPct,
			p.Seed, i+1, candsSince, timeSince, impPer10K, impPerSec))
		prevCandidate = d.Candidate
		prevElapsed = d.ElapsedMs
	}
	return []byte(b.String())
}

// --- VRPTW simple discoveries.csv ---

const vrptwDiscoveriesCSVHeader = "elapsed_ms,candidate,old_best,new_best,improvement\n"

func buildVRPTWDiscoveriesCSV(discoveries []optimisation.Discovery) []byte {
	if len(discoveries) == 0 {
		return nil
	}
	var b strings.Builder
	b.WriteString(vrptwDiscoveriesCSVHeader)
	for _, d := range discoveries {
		b.WriteString(fmt.Sprintf("%d,%d,%d,%d,%d\n", d.ElapsedMs, d.Candidate, d.OldBest, d.NewBest, d.Improvement))
	}
	return []byte(b.String())
}

// --- PFRS run.json (hand-formatted to preserve exact field layout) ---

type pfrsBeamRunJSONParams struct {
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
	Portfolio            []string
	RunLabel             string
}

func formatPFRSBeamRunJSON(p pfrsBeamRunJSONParams) string {
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
  "portfolio": %q,
  "runLabel": %q
}`, p.InstanceID, p.Mode, p.IterationsPerWorker,
		p.InitialTemperature, p.CoolingMode,
		p.EffectiveCoolingRate, p.LateAcceptanceLength,
		p.BeamWidth,
		strings.Join(seedParts, ", "),
		p.Seed, runtime.NumCPU(), p.MaxTotalWorkers,
		p.LookaheadWeight, p.FinalWindowWeeks, p.FinalWindowIter, p.BeamStrategy, p.DiversitySlotsPct,
		strings.Join(p.Portfolio, ","), p.RunLabel)
}

func writePFRSBeamRunJSON(outputDir string, p pfrsBeamRunJSONParams) {
	path := filepath.Join(outputDir, "run.json")
	if err := writeTelemetryBytesErr(path, []byte(formatPFRSBeamRunJSON(p))); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing run.json: %v\n", err)
	}
}

type pfrsStandardRunJSONParams struct {
	InstanceName string
	WorkerMode   string
	BestPenalty  int
	RunLabel     string
}

func formatPFRSStandardRunJSON(p pfrsStandardRunJSONParams) string {
	return fmt.Sprintf(`{
  "instance": %q,
  "problemType": "nrp",
  "mode": %q,
  "bestObjective": %d,
  "totalPenalty": %d,
  "runLabel": %q
}`, p.InstanceName, p.WorkerMode, p.BestPenalty, p.BestPenalty, p.RunLabel)
}

func writePFRSStandardRunJSON(outputDir string, p pfrsStandardRunJSONParams) {
	path := filepath.Join(outputDir, "run.json")
	if err := writeTelemetryBytesErr(path, []byte(formatPFRSStandardRunJSON(p))); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing run.json: %v\n", err)
	}
}

// --- PFRS worker intelligence CSVs ---

func emitPFRSWorkerDecisionsCSV(outputDir string, recorder *inrc2.ShadowRecorder) {
	if recorder == nil {
		return
	}
	records := recorder.Records()
	if len(records) == 0 {
		return
	}
	path := filepath.Join(outputDir, "worker_decisions.csv")
	if err := inrc2.WriteWorkerDecisionsCSV(path, records); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing worker_decisions.csv: %v\n", err)
	} else {
		fmt.Fprintf(os.Stderr, "Worker decisions CSV written: %s (%d records)\n", path, len(records))
	}
}

func emitPFRSWorkerAssistCSV(outputDir string, recorder *inrc2.AssistRecorder) {
	if recorder == nil {
		return
	}
	records := recorder.Records()
	if len(records) == 0 {
		return
	}
	path := filepath.Join(outputDir, "worker_assist.csv")
	if err := inrc2.WriteWorkerAssistCSV(path, records); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing worker_assist.csv: %v\n", err)
	} else {
		fmt.Fprintf(os.Stderr, "Worker assist CSV written: %s (%d records)\n", path, len(records))
	}
}

func emitPFRSWorkerIntelligenceCSVs(outputDir string, decisionRecorder *inrc2.ShadowRecorder, assistRecorder *inrc2.AssistRecorder) {
	emitPFRSWorkerDecisionsCSV(outputDir, decisionRecorder)
	emitPFRSWorkerAssistCSV(outputDir, assistRecorder)
}

func writePFRSAuditCSV(path string, rows []inrc2.WeekAuditRow) {
	if err := inrc2.WriteAuditCSV(path, rows); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing audit CSV: %v\n", err)
	} else {
		fmt.Fprintf(os.Stderr, "Audit CSV written: %s (%d rows)\n", path, len(rows))
	}
}

// --- Generic solver run finalisation ---

type genericSolverRunOutput struct {
	OutputDir    string
	RunMeta      map[string]interface{}
	SolutionJSON []byte
	ExtraFiles   map[string][]byte // optional additional telemetry files
	Telemetry    solverTelemetryInput
	Storage      storageConfig
	RunLabel     string
	Algorithm    string
	Penalty      int
}

func finalizeGenericSolverRun(out genericSolverRunOutput) {
	writeRunMetadata(out.OutputDir, out.RunMeta)
	if len(out.SolutionJSON) > 0 {
		writeSolutionJSON(out.OutputDir, out.SolutionJSON)
	}
	for name, data := range out.ExtraFiles {
		if len(data) > 0 {
			writeTelemetryFile(out.OutputDir, name, data)
		}
	}
	emitSolverTelemetry(out.Telemetry)
	uploadRunOutput(out.Storage, out.RunLabel, out.OutputDir, out.Algorithm, out.Penalty)
}

// solverTelemetryInput bundles optional post-run CSV emission for generic solvers.
type solverTelemetryInput struct {
	OutputDir         string
	ProblemType       string
	Instance          string
	Algorithm         string
	Seed              int64
	Temperature       float64
	Iterations        int
	Result            optimisation.SearchResult
	PortfolioRecorder *optimisation.PortfolioAssistRecorder
}

// emitSolverTelemetry writes worker_learning, portfolio_assist, and generic_search_assist CSVs.
func emitSolverTelemetry(in solverTelemetryInput) {
	inrc2.EmitSingleWorkerLearning(in.OutputDir, inrc2.SingleWorkerConfig{
		ProblemType: in.ProblemType,
		Instance:    in.Instance,
		Algorithm:   in.Algorithm,
		Seed:        in.Seed,
		Temperature: in.Temperature,
		Iterations:  in.Iterations,
	}, in.Result)

	if in.PortfolioRecorder != nil {
		records := in.PortfolioRecorder.Records()
		if len(records) > 0 {
			optimisation.WritePortfolioAssistCSV(filepath.Join(in.OutputDir, "portfolio_assist.csv"), records)
		}
	}

	if len(in.Result.AssistRecords) > 0 {
		optimisation.WriteSearchAssistCSV(filepath.Join(in.OutputDir, "generic_search_assist.csv"), in.Result.AssistRecords)
	}
}
