package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2/siadapter"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/telemetry/workerlearning"
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

// logTelemetryFileWrite logs stderr success/error for telemetry file writes.
func logTelemetryFileWrite(err error, errLabel, successMsg string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", errLabel, err)
	} else if successMsg != "" {
		fmt.Fprintf(os.Stderr, "%s\n", successMsg)
	}
}

// --- PFRS worker intelligence CSVs (delegates to siadapter) ---

func emitPFRSWorkerIntelligenceCSVs(outputDir string, decisionRecorder *inrc2.ShadowRecorder, assistRecorder *inrc2.AssistRecorder) {
	siadapter.EmitPFRSWorkerIntelligenceCSVs(outputDir, decisionRecorder, assistRecorder)
}

func emitPFRSTelemetry(in siadapter.PFRSTelemetryInput) {
	siadapter.EmitPFRSTelemetry(in)
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
	AssistMode        string
	Result            optimisation.SearchResult
	PortfolioRecorder *optimisation.PortfolioAssistRecorder
	PolicyMode        string
	PolicyDir         string
}

// emitSolverTelemetry writes Search Intelligence CSVs for generic solvers:
// worker_learning.csv, portfolio_assist.csv (portfolio mode), generic_search_assist.csv (search-level),
// policy_decisions.csv (SI 2.0 when --policy-mode is set).
func emitSolverTelemetry(in solverTelemetryInput) {
	written := map[string]bool{}

	workerlearning.EmitSingleWorkerLearning(in.OutputDir, workerlearning.SingleWorkerConfig{
		ProblemType: in.ProblemType,
		Instance:    in.Instance,
		Algorithm:   in.Algorithm,
		Seed:        in.Seed,
		Temperature: in.Temperature,
		Iterations:  in.Iterations,
	}, in.Result)
	written["worker_learning.csv"] = true

	assistMode := in.AssistMode
	if assistMode == "" {
		assistMode = "off"
	}

	if in.PortfolioRecorder != nil {
		records := in.PortfolioRecorder.Records()
		if len(records) > 0 {
			optimisation.WritePortfolioAssistCSV(filepath.Join(in.OutputDir, "portfolio_assist.csv"), records)
			written["portfolio_assist.csv"] = true
		}
	}

	if len(in.Result.AssistRecords) > 0 {
		optimisation.WriteSearchAssistCSV(filepath.Join(in.OutputDir, "generic_search_assist.csv"), in.Result.AssistRecords)
		written["generic_search_assist.csv"] = true
		siadapter.EmitAdaptedWorkerCSVs(in.OutputDir, assistMode, in.Result.AssistRecords)
		if assistMode == "shadow" {
			written["worker_decisions.csv"] = true
		}
		if assistMode == "assist" || assistMode == "adaptive" {
			written["worker_assist.csv"] = true
		}
	}

	if len(in.Result.PolicyDecisions) > 0 {
		optimisation.WritePolicyDecisionsCSV(filepath.Join(in.OutputDir, "policy_decisions.csv"), in.Result.PolicyDecisions)
		written["policy_decisions.csv"] = true
		evalInput := optimisation.PolicyEvaluationInput{
			RunID:          filepath.Base(in.OutputDir),
			Domain:         in.ProblemType,
			Instance:       in.Instance,
			Algorithm:      in.Algorithm,
			InitialPenalty: in.Result.InitialPenalty,
			BestPenalty:    in.Result.BestPenalty,
			Decisions:      in.Result.PolicyDecisions,
		}
		_ = optimisation.WritePolicyEvaluationCSV(in.OutputDir, evalInput)
		written["policy_evaluation.csv"] = true
		if err := optimisation.WriteCounterfactualFromPolicyDecisions(in.OutputDir, optimisation.CounterfactualEmitInput{
			RunID:          evalInput.RunID,
			Domain:         evalInput.Domain,
			Instance:       evalInput.Instance,
			Algorithm:      evalInput.Algorithm,
			InitialPenalty: evalInput.InitialPenalty,
			BestPenalty:    evalInput.BestPenalty,
			Decisions:      evalInput.Decisions,
		}); err == nil {
			written["counterfactual_learning.csv"] = true
		}
	}

	optimisation.EnsureSITelemetryContract(in.OutputDir, written)
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
