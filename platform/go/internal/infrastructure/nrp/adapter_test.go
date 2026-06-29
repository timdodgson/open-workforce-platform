package nrp_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/nrp"
)

func sampleInput() nrp.NRPInput {
	return nrp.NRPInput{
		Nurses: []nrp.Nurse{
			{ID: "N-001", Skills: []string{"general", "medication"}, Available: true},
			{ID: "N-002", Skills: []string{"general"}, Available: true},
			{ID: "N-003", Skills: []string{"general"}, Available: false},
		},
		Shifts: []nrp.Shift{
			{ID: "early", Name: "Early", StartMinute: 360, EndMinute: 840},
		},
		Demands: []nrp.Demand{
			{Day: 1, ShiftID: "early", RequiredSkill: "general", Minimum: 2, Optimal: 2},
			{Day: 1, ShiftID: "early", RequiredSkill: "medication", Minimum: 1, Optimal: 1},
		},
	}
}

func TestConvert_NursesToResources(t *testing.T) {
	dataset := nrp.Convert(sampleInput())

	if len(dataset.Resources) != 3 {
		t.Fatalf("expected 3 resources, got %d", len(dataset.Resources))
	}
	if dataset.Resources[0].ID != "N-001" {
		t.Errorf("expected first resource N-001, got %s", dataset.Resources[0].ID)
	}
}

func TestConvert_DemandsToEvents(t *testing.T) {
	dataset := nrp.Convert(sampleInput())

	// 2 general + 1 medication = 3 events
	if len(dataset.BusinessEvents) != 3 {
		t.Fatalf("expected 3 events, got %d", len(dataset.BusinessEvents))
	}
}

func TestConvert_SkillsPreserved(t *testing.T) {
	dataset := nrp.Convert(sampleInput())

	var details struct {
		Skills []string `json:"skills"`
	}
	json.Unmarshal(dataset.Resources[0].Details, &details)

	if len(details.Skills) != 2 || details.Skills[0] != "general" {
		t.Errorf("expected skills preserved, got %v", details.Skills)
	}
}

func TestConvert_EventHasRequiredSkill(t *testing.T) {
	dataset := nrp.Convert(sampleInput())

	var details struct {
		RequiredSkill string `json:"requiredSkill"`
		Duration      int    `json:"duration"`
		EarliestStart int    `json:"earliestStart"`
		LatestFinish  int    `json:"latestFinish"`
	}
	json.Unmarshal(dataset.BusinessEvents[0].Details, &details)

	if details.RequiredSkill != "general" {
		t.Errorf("expected requiredSkill general, got %s", details.RequiredSkill)
	}
	if details.Duration != 480 {
		t.Errorf("expected duration 480, got %d", details.Duration)
	}
	if details.EarliestStart != 360 {
		t.Errorf("expected earliestStart 360, got %d", details.EarliestStart)
	}
}

func TestConvert_Deterministic(t *testing.T) {
	d1 := nrp.Convert(sampleInput())
	d2 := nrp.Convert(sampleInput())

	data1, _ := json.Marshal(d1)
	data2, _ := json.Marshal(d2)

	if string(data1) != string(data2) {
		t.Error("conversion is not deterministic")
	}
}

func TestLoadNRP(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")

	input := sampleInput()
	data, _ := json.Marshal(input)
	os.WriteFile(path, data, 0644)

	loaded, err := nrp.LoadNRP(path)
	if err != nil {
		t.Fatalf("failed to load: %v", err)
	}
	if len(loaded.Nurses) != 3 {
		t.Errorf("expected 3 nurses, got %d", len(loaded.Nurses))
	}
}

func TestWriteAndOptimise(t *testing.T) {
	dataset := nrp.Convert(sampleInput())

	dir := t.TempDir()
	path := filepath.Join(dir, "output.json")
	if err := nrp.WriteDataset(dataset, path); err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	// Verify the output file is valid JSON.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if _, ok := raw["businessEvents"]; !ok {
		t.Error("output missing businessEvents")
	}
	if _, ok := raw["resources"]; !ok {
		t.Error("output missing resources")
	}
}
