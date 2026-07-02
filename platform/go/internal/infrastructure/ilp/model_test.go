package ilp_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/ilp"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2"
)

const testDataDir = "../../../../../examples/inrc2/testdatasets_json/n005w4/"

func loadTestInstance(t *testing.T) (inrc2.Scenario, []string, inrc2.History) {
	t.Helper()
	sc, err := inrc2.LoadScenario(testDataDir + "Sc-n005w4.json")
	if err != nil {
		t.Fatalf("failed to load scenario: %v", err)
	}
	hist, err := inrc2.LoadHistory(testDataDir + "H0-n005w4-0.json")
	if err != nil {
		t.Fatalf("failed to load history: %v", err)
	}
	weekFiles := []string{
		testDataDir + "WD-n005w4-0.json",
		testDataDir + "WD-n005w4-1.json",
		testDataDir + "WD-n005w4-2.json",
		testDataDir + "WD-n005w4-3.json",
	}
	return sc, weekFiles, hist
}

func TestBuildModel_ProducesValidLPFile(t *testing.T) {
	sc, weekFiles, hist := loadTestInstance(t)

	tmpDir := t.TempDir()
	modelPath := filepath.Join(tmpDir, "test.lp")

	info, err := ilp.BuildModel(sc, weekFiles, hist, 1, modelPath)
	if err != nil {
		t.Fatalf("BuildModel failed: %v", err)
	}

	// Check model info.
	if info.NumNurses != 5 {
		t.Errorf("expected 5 nurses, got %d", info.NumNurses)
	}
	if info.NumDays != 7 {
		t.Errorf("expected 7 days for 1 week, got %d", info.NumDays)
	}
	if info.NumShiftTypes != 3 {
		t.Errorf("expected 3 shift types, got %d", info.NumShiftTypes)
	}
	if info.NumVariables <= 0 {
		t.Error("expected positive variable count")
	}
	if info.NumConstraints <= 0 {
		t.Error("expected positive constraint count")
	}

	// Read the file and verify structure.
	data, err := os.ReadFile(modelPath)
	if err != nil {
		t.Fatalf("failed to read model: %v", err)
	}
	content := string(data)

	// Must have LP format sections.
	if !strings.Contains(content, "Minimize") {
		t.Error("missing Minimize section")
	}
	if !strings.Contains(content, "Subject To") {
		t.Error("missing Subject To section")
	}
	if !strings.Contains(content, "Binary") {
		t.Error("missing Binary section")
	}
	if !strings.Contains(content, "End") {
		t.Error("missing End marker")
	}

	// Should contain x_ variables with skill dimension.
	if !strings.Contains(content, "x_0_0_Early_") {
		t.Error("missing expected variable x_0_0_Early_<skill>")
	}
}

func TestBuildModel_MultiWeek(t *testing.T) {
	sc, weekFiles, hist := loadTestInstance(t)

	tmpDir := t.TempDir()
	modelPath := filepath.Join(tmpDir, "multi.lp")

	info, err := ilp.BuildModel(sc, weekFiles, hist, 2, modelPath)
	if err != nil {
		t.Fatalf("BuildModel failed: %v", err)
	}

	if info.NumDays != 14 {
		t.Errorf("expected 14 days for 2 weeks, got %d", info.NumDays)
	}

	// 2-week model should have more constraints than 1-week.
	tmpDir2 := t.TempDir()
	modelPath2 := filepath.Join(tmpDir2, "single.lp")
	info1, _ := ilp.BuildModel(sc, weekFiles, hist, 1, modelPath2)
	if info.NumConstraints <= info1.NumConstraints {
		t.Errorf("2-week model should have more constraints than 1-week: %d vs %d",
			info.NumConstraints, info1.NumConstraints)
	}
}

func TestBuildModel_ForbiddenSuccession(t *testing.T) {
	sc, weekFiles, hist := loadTestInstance(t)

	tmpDir := t.TempDir()
	modelPath := filepath.Join(tmpDir, "forbidden.lp")

	_, err := ilp.BuildModel(sc, weekFiles, hist, 1, modelPath)
	if err != nil {
		t.Fatalf("BuildModel failed: %v", err)
	}

	data, err := os.ReadFile(modelPath)
	if err != nil {
		t.Fatalf("failed to read model: %v", err)
	}
	content := string(data)

	// n005w4 has forbidden: Late->Early, Night->Early, Night->Late.
	// Should have h4_ constraints.
	if !strings.Contains(content, "h4_") {
		t.Error("missing forbidden succession constraints (h4_)")
	}
}

func TestBuildModel_MaxConsecutiveWorkingDays(t *testing.T) {
	sc, weekFiles, hist := loadTestInstance(t)

	tmpDir := t.TempDir()
	modelPath := filepath.Join(tmpDir, "maxwork.lp")

	info, err := ilp.BuildModel(sc, weekFiles, hist, 1, modelPath)
	if err != nil {
		t.Fatalf("BuildModel failed: %v", err)
	}

	// All constraints should be supported.
	if len(info.UnsupportedConstraints) != 0 {
		t.Errorf("expected no unsupported constraints, got %v", info.UnsupportedConstraints)
	}
	// S2 should be in supported list.
	found := false
	for _, c := range info.SupportedConstraints {
		if c == "S2" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected S2 in supported constraints")
	}
}

func TestBuildModel_HistoryForbiddenSuccession(t *testing.T) {
	sc, weekFiles, hist := loadTestInstance(t)

	// Patrick has lastAssignedShiftType="Night" and consecutiveDaysOff=0.
	// Night->Early and Night->Late are forbidden.
	// Model should prevent Patrick from working Early or Late on day 0.

	tmpDir := t.TempDir()
	modelPath := filepath.Join(tmpDir, "histforbid.lp")

	_, err := ilp.BuildModel(sc, weekFiles, hist, 1, modelPath)
	if err != nil {
		t.Fatalf("BuildModel failed: %v", err)
	}

	data, err := os.ReadFile(modelPath)
	if err != nil {
		t.Fatalf("failed to read model: %v", err)
	}
	content := string(data)

	// Should have h4h_ constraint for Patrick (nurse 0) forbidding Early on day 0.
	if !strings.Contains(content, "h4h_0_") {
		t.Error("missing history-based forbidden succession constraint for Patrick")
	}
}

func TestBuildModel_ShiftOffRequests(t *testing.T) {
	sc, weekFiles, hist := loadTestInstance(t)

	tmpDir := t.TempDir()
	modelPath := filepath.Join(tmpDir, "shiftoff.lp")

	_, err := ilp.BuildModel(sc, weekFiles, hist, 1, modelPath)
	if err != nil {
		t.Fatalf("BuildModel failed: %v", err)
	}

	data, err := os.ReadFile(modelPath)
	if err != nil {
		t.Fatalf("failed to read model: %v", err)
	}
	content := string(data)

	// Week 0 has 3 shift-off requests: Sara-Any-Thursday, Sara-Night-Saturday, Stefaan-Late-Saturday.
	// Should have s5_ penalty variables in objective.
	if !strings.Contains(content, "s5_0_0") {
		t.Error("missing shift-off request variable s5_0_0")
	}
	if !strings.Contains(content, "s5_0_1") {
		t.Error("missing shift-off request variable s5_0_1")
	}
}

func TestBuildModel_CompleteWeekends(t *testing.T) {
	sc, weekFiles, hist := loadTestInstance(t)

	tmpDir := t.TempDir()
	modelPath := filepath.Join(tmpDir, "weekends.lp")

	_, err := ilp.BuildModel(sc, weekFiles, hist, 1, modelPath)
	if err != nil {
		t.Fatalf("BuildModel failed: %v", err)
	}

	data, err := os.ReadFile(modelPath)
	if err != nil {
		t.Fatalf("failed to read model: %v", err)
	}
	content := string(data)

	// Both contracts have completeWeekends=1.
	// Should have s6_ constraints.
	if !strings.Contains(content, "s6a_") {
		t.Error("missing complete weekend constraint s6a_")
	}
	if !strings.Contains(content, "s6b_") {
		t.Error("missing complete weekend constraint s6b_")
	}
}
