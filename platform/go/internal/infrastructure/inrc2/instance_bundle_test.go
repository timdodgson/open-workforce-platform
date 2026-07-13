package inrc2_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2"
)

func TestLoadInstanceBundle_n005w4(t *testing.T) {
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })
	if err := os.Chdir("../../../"); err != nil {
		t.Fatalf("chdir platform/go: %v", err)
	}

	bundle, err := inrc2.LoadInstanceBundle("n005w4")
	if err != nil {
		t.Fatalf("LoadInstanceBundle: %v", err)
	}
	if bundle.Scenario.ID == "" {
		t.Fatal("expected scenario ID")
	}
	if len(bundle.WeekFiles) == 0 {
		t.Fatal("expected week files")
	}
	if len(bundle.HistFiles) == 0 {
		t.Fatal("expected history files")
	}
}

func TestSelectHistoryAndWeeks_n030w4(t *testing.T) {
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })
	if err := os.Chdir("../../../"); err != nil {
		t.Fatalf("chdir platform/go: %v", err)
	}

	bundle, err := inrc2.LoadInstanceBundle("n030w4")
	if err != nil {
		t.Skipf("n030w4 not available: %v", err)
	}
	if err := bundle.SelectHistory(1); err != nil {
		t.Fatalf("SelectHistory(1): %v", err)
	}
	weeks, err := inrc2.ParseWeekSequence("6-2-9-1")
	if err != nil {
		t.Fatalf("ParseWeekSequence: %v", err)
	}
	if err := bundle.SelectWeeks(weeks); err != nil {
		t.Fatalf("SelectWeeks: %v", err)
	}
	if len(bundle.WeekFiles) != 4 {
		t.Fatalf("expected 4 week files, got %d", len(bundle.WeekFiles))
	}
	want := []string{"WD-n030w4-6.json", "WD-n030w4-2.json", "WD-n030w4-9.json", "WD-n030w4-1.json"}
	for i, path := range bundle.WeekFiles {
		base := filepath.Base(path)
		if base != want[i] {
			t.Errorf("week %d: got %s want %s", i, base, want[i])
		}
	}
}

func TestWireWorkerIntelligence_shadow(t *testing.T) {
	wire, lines := inrc2.WireWorkerIntelligence("shadow", "", "")
	if wire.DecisionRecorder == nil {
		t.Fatal("expected decision recorder")
	}
	if wire.AssistRecorder != nil {
		t.Fatal("shadow mode should not create assist recorder")
	}
	if len(lines) == 0 {
		t.Fatal("expected log lines")
	}
}
