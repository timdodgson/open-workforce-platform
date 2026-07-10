package policy

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFeatureSchemaVersion(t *testing.T) {
	if FeatureSchemaVersion == "" {
		t.Fatal("FeatureSchemaVersion must not be empty")
	}
}

func TestFeatureStore_Disabled(t *testing.T) {
	fs := NewFeatureStore("")
	err := fs.Record(FeatureRecord{})
	if err != nil {
		t.Errorf("disabled store should not error, got: %v", err)
	}
	if fs.Count() != 0 {
		t.Errorf("disabled store count should be 0, got %d", fs.Count())
	}
}

func TestFeatureStore_WriteAndRead(t *testing.T) {
	dir := t.TempDir()
	fs := NewFeatureStore(dir)
	defer fs.Close()

	fv := FeatureVector{
		SchemaVersion: FeatureSchemaVersion,
		Problem:       "cvrp",
		Instance:      "A-n32-k5",
		Algorithm:     "sa",
	}

	record := FeatureRecord{
		Features:     fv,
		Action:       "early_stop",
		Confidence:   0.82,
		PolicySource: "learned",
		Outcome: FeatureOutcome{
			Improved:       false,
			FinalObjective: 784,
			ComputeUsed:    100000,
			RuntimeMs:      78,
		},
	}

	if err := fs.Record(record); err != nil {
		t.Fatalf("Record failed: %v", err)
	}
	if fs.Count() != 1 {
		t.Errorf("Count = %d, want 1", fs.Count())
	}

	fs.Close()
	data, err := os.ReadFile(filepath.Join(dir, "features.jsonl"))
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}

	var decoded FeatureRecord
	if err := json.Unmarshal([]byte(lines[0]), &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.Features.SchemaVersion != FeatureSchemaVersion {
		t.Errorf("decoded SchemaVersion = %q", decoded.Features.SchemaVersion)
	}
	if decoded.Action != "early_stop" {
		t.Errorf("decoded Action = %q, want early_stop", decoded.Action)
	}
	if decoded.Confidence != 0.82 {
		t.Errorf("decoded Confidence = %f, want 0.82", decoded.Confidence)
	}
	if decoded.Features.Problem != "cvrp" {
		t.Errorf("decoded Problem = %q, want cvrp", decoded.Features.Problem)
	}
}
