package optimisation

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveBudgetPolicyPath(t *testing.T) {
	dir := t.TempDir()
	model := filepath.Join(dir, BudgetPolicyFilename)
	if err := os.WriteFile(model, []byte(`{"version":"1","trained_on":1,"entries":[]}`), 0644); err != nil {
		t.Fatal(err)
	}

	got := ResolveBudgetPolicyPath(dir, "/legacy/model.json", "hybrid")
	if got != model {
		t.Errorf("got %q, want %q", got, model)
	}

	got = ResolveBudgetPolicyPath(dir, "/legacy/model.json", "rules")
	if got != "/legacy/model.json" {
		t.Errorf("rules mode should use legacy path, got %q", got)
	}
}
