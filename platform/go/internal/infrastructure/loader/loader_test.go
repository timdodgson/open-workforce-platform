package loader_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/loader"
)

func TestLoadEvents_ValidFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "events.json")

	data := `[
		{"id":"EVT-001","type":"test.event","occurredAt":"2026-06-15T08:00:00Z","details":{"key":"value"}},
		{"id":"EVT-002","type":"test.event","occurredAt":"2026-06-15T09:00:00Z","details":{"key":"other"}}
	]`
	os.WriteFile(path, []byte(data), 0644)

	events, err := loader.LoadEvents(path)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(events) != 2 {
		t.Errorf("expected 2 events, got %d", len(events))
	}
	if events[0].ID() != "EVT-001" {
		t.Errorf("expected first event id EVT-001, got %s", events[0].ID())
	}
}

func TestLoadEvents_FileNotFound(t *testing.T) {
	_, err := loader.LoadEvents("/nonexistent/path.json")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadEvents_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	os.WriteFile(path, []byte(`not json`), 0644)

	_, err := loader.LoadEvents(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestLoadEvents_EmptyArray(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.json")
	os.WriteFile(path, []byte(`[]`), 0644)

	_, err := loader.LoadEvents(path)
	if err == nil {
		t.Fatal("expected error for empty dataset")
	}
}

func TestLoadEvents_InvalidEvent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "invalid-event.json")

	// Missing required id field
	data := `[{"id":"","type":"test.event","occurredAt":"2026-06-15T08:00:00Z","details":{"key":"value"}}]`
	os.WriteFile(path, []byte(data), 0644)

	_, err := loader.LoadEvents(path)
	if err == nil {
		t.Fatal("expected error for invalid event in dataset")
	}
}
