package inrc2

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFinalizeBeamArtifacts(t *testing.T) {
	dir := t.TempDir()
	auditPath := filepath.Join(dir, "results.csv")

	err := FinalizeBeamArtifacts(BeamArtifactsParams{
		OutputDir:    dir,
		AuditCSVPath: auditPath,
		ScenarioID:   "test-instance",
		Config:       PFRSConfig{Mode: "single", IterationsPerWorker: 1000},
		WinningPath:  nil,
		RunJSON: PFRSBeamRunJSONParams{
			InstanceID: "test-instance",
			Mode:       "single",
			RunLabel:   "test-run",
		},
	})
	if err != nil {
		t.Fatalf("FinalizeBeamArtifacts: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "run.json")); err != nil {
		t.Fatalf("run.json not written: %v", err)
	}
}
