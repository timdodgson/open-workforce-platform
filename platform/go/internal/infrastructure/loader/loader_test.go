package loader_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/loader"
)

const validDataset = `{
	"businessEvents": [
		{"id":"EVT-001","type":"test.event","occurredAt":"2026-06-15T08:00:00Z","details":{"key":"value"}},
		{"id":"EVT-002","type":"test.event","occurredAt":"2026-06-15T09:00:00Z","details":{"key":"other"}}
	],
	"resources": [
		{"id":"RES-001","type":"person","details":{"name":"Alice"}},
		{"id":"RES-002","type":"vehicle","details":{"plate":"AB12 CDE"}}
	]
}`

func writeFile(t *testing.T, name, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	os.WriteFile(path, []byte(content), 0644)
	return path
}

func TestLoadDataset_Valid(t *testing.T) {
	path := writeFile(t, "dataset.json", validDataset)

	ds, err := loader.LoadDataset(path)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(ds.Events) != 2 {
		t.Errorf("expected 2 events, got %d", len(ds.Events))
	}
	if len(ds.Resources) != 2 {
		t.Errorf("expected 2 resources, got %d", len(ds.Resources))
	}
	if ds.Events[0].ID() != "EVT-001" {
		t.Errorf("expected first event id EVT-001, got %s", ds.Events[0].ID())
	}
	if ds.Resources[0].ID() != "RES-001" {
		t.Errorf("expected first resource id RES-001, got %s", ds.Resources[0].ID())
	}
}

func TestLoadDataset_FileNotFound(t *testing.T) {
	_, err := loader.LoadDataset("/nonexistent/path.json")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadDataset_InvalidJSON(t *testing.T) {
	path := writeFile(t, "bad.json", `not json`)

	_, err := loader.LoadDataset(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestLoadDataset_NoEvents(t *testing.T) {
	data := `{"businessEvents":[],"resources":[{"id":"RES-001","type":"person","details":{"name":"Alice"}}]}`
	path := writeFile(t, "no-events.json", data)

	_, err := loader.LoadDataset(path)
	if err == nil {
		t.Fatal("expected error for empty events")
	}
}

func TestLoadDataset_NoResources(t *testing.T) {
	data := `{"businessEvents":[{"id":"EVT-001","type":"test.event","occurredAt":"2026-06-15T08:00:00Z","details":{"key":"value"}}],"resources":[]}`
	path := writeFile(t, "no-resources.json", data)

	_, err := loader.LoadDataset(path)
	if err == nil {
		t.Fatal("expected error for empty resources")
	}
}

func TestLoadDataset_InvalidEvent(t *testing.T) {
	data := `{"businessEvents":[{"id":"","type":"test","occurredAt":"2026-06-15T08:00:00Z","details":{"k":"v"}}],"resources":[{"id":"RES-001","type":"person","details":{"name":"Alice"}}]}`
	path := writeFile(t, "invalid-event.json", data)

	_, err := loader.LoadDataset(path)
	if err == nil {
		t.Fatal("expected error for invalid event")
	}
}

func TestLoadDataset_InvalidResource(t *testing.T) {
	data := `{"businessEvents":[{"id":"EVT-001","type":"test","occurredAt":"2026-06-15T08:00:00Z","details":{"k":"v"}}],"resources":[{"id":"","type":"person","details":{"name":"Alice"}}]}`
	path := writeFile(t, "invalid-resource.json", data)

	_, err := loader.LoadDataset(path)
	if err == nil {
		t.Fatal("expected error for invalid resource")
	}
}
