package siadapter

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

// PFRSTelemetryInput bundles PFRS run metadata for SI CSV emission.
type PFRSTelemetryInput struct {
	OutputDir          string
	Instance           string
	WorkerMode         string
	Portfolio          []string
	Seed               int64
	Iterations         int
	BestPenalty        int
	DecisionRecorder   *inrc2.ShadowRecorder
	AssistRecorder     *inrc2.AssistRecorder
	PolicyMode         string
	PolicyDir          string
	WorkerDecisionMode string
}

// EmitPFRSWorkerIntelligenceCSVs writes worker decision/assist CSVs only.
func EmitPFRSWorkerIntelligenceCSVs(outputDir string, decisionRecorder *inrc2.ShadowRecorder, assistRecorder *inrc2.AssistRecorder) {
	EmitPFRSTelemetry(PFRSTelemetryInput{
		OutputDir:        outputDir,
		DecisionRecorder: decisionRecorder,
		AssistRecorder:   assistRecorder,
	})
}

// EmitPFRSTelemetry writes PFRS SI CSVs with cross-layer adapters and contract stubs.
func EmitPFRSTelemetry(in PFRSTelemetryInput) {
	if in.OutputDir == "" {
		return
	}
	written := map[string]bool{}
	if _, err := os.Stat(filepath.Join(in.OutputDir, "worker_learning.csv")); err == nil {
		written["worker_learning.csv"] = true
	}

	emitPFRSWorkerDecisionsCSV(in.OutputDir, in.DecisionRecorder)
	if in.DecisionRecorder != nil && len(in.DecisionRecorder.Records()) > 0 {
		written["worker_decisions.csv"] = true
	}
	emitPFRSWorkerAssistCSV(in.OutputDir, in.AssistRecorder)
	if in.AssistRecorder != nil && len(in.AssistRecorder.Records()) > 0 {
		written["worker_assist.csv"] = true
	}

	if EmitAdaptedSearchCSV(in.OutputDir, in.Iterations, in.DecisionRecorder, in.AssistRecorder) {
		written["generic_search_assist.csv"] = true
	}

	if in.WorkerMode == "portfolio" && len(in.Portfolio) > 0 {
		records := BuildNRPPortfolioAssistRecords(in.Instance, in.Seed, in.Portfolio, in.Iterations, in.BestPenalty)
		if len(records) > 0 {
			path := filepath.Join(in.OutputDir, "portfolio_assist.csv")
			if err := optimisation.WritePortfolioAssistCSV(path, records); err == nil {
				written["portfolio_assist.csv"] = true
			}
		}
	}

	EmitNRPPolicyCSVs(in.OutputDir, NRPPolicyEmitInput{
		PolicyMode:       in.PolicyMode,
		Instance:         in.Instance,
		WorkerMode:       in.WorkerMode,
		BestPenalty:      in.BestPenalty,
		DecisionRecorder: in.DecisionRecorder,
		AssistRecorder:   in.AssistRecorder,
	}, written)
	optimisation.EnsureSITelemetryContract(in.OutputDir, written)

	if in.PolicyMode != "" {
		assistCount := 0
		policyCount := 0
		if in.DecisionRecorder != nil {
			assistCount += len(in.DecisionRecorder.Records())
		}
		if in.AssistRecorder != nil {
			assistCount += len(in.AssistRecorder.Records())
		}
		if written["policy_decisions.csv"] {
			if in.DecisionRecorder != nil {
				policyCount += len(in.DecisionRecorder.Records())
			}
			if in.AssistRecorder != nil {
				policyCount += len(in.AssistRecorder.Records())
			}
		}
		report := optimisation.RunPostRunPolicyPipeline(optimisation.PostRunPolicyConfig{
			PolicyMode:          in.PolicyMode,
			PolicyDir:           in.PolicyDir,
			OutputDir:           in.OutputDir,
			Domain:              "nrp",
			PolicyDecisionCount: policyCount,
			AssistRecordCount:   assistCount,
		})
		if summary := optimisation.FormatPostRunSummary(report); summary != "" {
			fmt.Fprintf(os.Stderr, "%s\n", summary)
		}
	}
}

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

func logTelemetryFileWrite(err error, errLabel, successMsg string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", errLabel, err)
	} else if successMsg != "" {
		fmt.Fprintf(os.Stderr, "%s\n", successMsg)
	}
}
