package optimisation

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureSITelemetryContract(t *testing.T) {
	dir := t.TempDir()
	EnsureSITelemetryContract(dir, nil)

	for _, name := range SITelemetryCSV {
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected %s to exist: %v", name, err)
		}
	}
}

func TestCanPromoteLearnedPolicy(t *testing.T) {
	if !CanPromoteLearnedPolicy(0.85) {
		t.Fatal("expected promotion at 85%")
	}
	if CanPromoteLearnedPolicy(0.50) {
		t.Fatal("expected block at 50%")
	}
}
