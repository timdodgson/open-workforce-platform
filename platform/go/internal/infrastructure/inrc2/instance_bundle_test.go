package inrc2_test

import (
	"os"
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
