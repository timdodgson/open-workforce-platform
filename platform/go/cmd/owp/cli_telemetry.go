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
	_ = writeTelemetryBytesErr(path, data)
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

// logTelemetryFileWrite logs stderr success/error for telemetry file writes.
func logTelemetryFileWrite(err error, errLabel, successMsg string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", errLabel, err)
	} else if successMsg != "" {
		fmt.Fprintf(os.Stderr, "%s\n", successMsg)
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
	err := inrc2.WriteWorkerDecisionsCSV(path, records)
	logTelemetryFileWrite(err, "worker_decisions.csv", fmt.Sprintf("Worker decisions CSV written: %s (%d records)", path, len(records)))
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
	err := inrc2.WriteWorkerAssistCSV(path, records)
	logTelemetryFileWrite(err, "worker_assist.csv", fmt.Sprintf("Worker assist CSV written: %s (%d records)", path, len(records)))
}

func emitPFRSWorkerIntelligenceCSVs(outputDir string, decisionRecorder *inrc2.ShadowRecorder, assistRecorder *inrc2.AssistRecorder) {
	emitPFRSWorkerDecisionsCSV(outputDir, decisionRecorder)
	emitPFRSWorkerAssistCSV(outputDir, assistRecorder)
}

func writePFRSAuditCSV(path string, rows []inrc2.WeekAuditRow) {
	err := inrc2.WriteAuditCSV(path, rows)
	logTelemetryFileWrite(err, "audit CSV", fmt.Sprintf("Audit CSV written: %s (%d rows)", path, len(rows)))
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
	runPostRunPolicyPipeline(out.Telemetry)
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
	PolicyMode        string
	PolicyDir         string
}

// emitSolverTelemetry writes Search Intelligence CSVs for generic solvers:
// worker_learning.csv, portfolio_assist.csv (portfolio mode), generic_search_assist.csv (search-level),
// policy_decisions.csv (SI 2.0 when --policy-mode is set).
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

	if len(in.Result.PolicyDecisions) > 0 {
		optimisation.WritePolicyDecisionsCSV(filepath.Join(in.OutputDir, "policy_decisions.csv"), in.Result.PolicyDecisions)
	}
}

func runPostRunPolicyPipeline(in solverTelemetryInput) {
	if in.PolicyMode == "" || in.OutputDir == "" {
		return
	}
	assistCount := len(in.Result.AssistRecords)
	if in.PortfolioRecorder != nil {
		assistCount += len(in.PortfolioRecorder.Records())
	}
	report := optimisation.RunPostRunPolicyPipeline(optimisation.PostRunPolicyConfig{
		PolicyMode:          in.PolicyMode,
		PolicyDir:           in.PolicyDir,
		OutputDir:           in.OutputDir,
		Domain:              in.ProblemType,
		PolicyDecisionCount: len(in.Result.PolicyDecisions),
		AssistRecordCount:   assistCount,
	})
	if summary := optimisation.FormatPostRunSummary(report); summary != "" {
		fmt.Fprintf(os.Stderr, "%s\n", summary)
	}
}
