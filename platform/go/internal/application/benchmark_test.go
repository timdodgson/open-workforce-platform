package application_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/application"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/loader"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

var algorithms = []string{"constructive", "hill-climbing", "simulated-annealing"}

var datasets = []string{
	"constructive-baseline.json",
	"skill-trap.json",
	"travel-trap.json",
	"time-window-trap.json",
	"capacity-trap.json",
	"preference-trap.json",
}

func datasetsDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..", "..", "..", "examples", "datasets")
}

func convertTravel(entries []loader.TravelEntry) []optimisation.TravelEntry {
	result := make([]optimisation.TravelEntry, len(entries))
	for i, e := range entries {
		result[i] = optimisation.TravelEntry{From: e.From, To: e.To, Minutes: e.Minutes}
	}
	return result
}

func TestBenchmark_AllAlgorithmsAllDatasets(t *testing.T) {
	dir := datasetsDir()

	for _, ds := range datasets {
		for _, alg := range algorithms {
			name := ds + "/" + alg
			t.Run(name, func(t *testing.T) {
				path := filepath.Join(dir, ds)
				dataset, err := loader.LoadDataset(path)
				if err != nil {
					t.Fatalf("failed to load dataset %s: %v", ds, err)
				}

				result, err := application.Optimise(dataset.Events, dataset.Resources, convertTravel(dataset.TravelMatrix), alg)
				if err != nil {
					t.Fatalf("algorithm %s failed on %s: %v", alg, ds, err)
				}

				if result.Score() < 0 || result.Score() > 100 {
					t.Errorf("invalid assignment score: %d", result.Score())
				}

				if result.ObjectiveScore() == 0 && result.Size() > 0 {
					t.Error("expected non-zero objective score with assignments")
				}

				stats := result.Statistics()
				if stats.Algorithm != alg {
					t.Errorf("expected algorithm %s in stats, got %s", alg, stats.Algorithm)
				}
				if stats.Iterations < 1 {
					t.Errorf("expected at least 1 iteration, got %d", stats.Iterations)
				}
			})
		}
	}
}

func TestBenchmark_Deterministic(t *testing.T) {
	dir := datasetsDir()

	for _, ds := range datasets {
		for _, alg := range algorithms {
			name := ds + "/" + alg + "/deterministic"
			t.Run(name, func(t *testing.T) {
				path := filepath.Join(dir, ds)
				dataset, err := loader.LoadDataset(path)
				if err != nil {
					t.Fatalf("failed to load: %v", err)
				}

				travel := convertTravel(dataset.TravelMatrix)
				r1, _ := application.Optimise(dataset.Events, dataset.Resources, travel, alg)
				r2, _ := application.Optimise(dataset.Events, dataset.Resources, travel, alg)

				if r1.ObjectiveScore() != r2.ObjectiveScore() {
					t.Errorf("non-deterministic: %d vs %d", r1.ObjectiveScore(), r2.ObjectiveScore())
				}
				if r1.Size() != r2.Size() {
					t.Errorf("non-deterministic size: %d vs %d", r1.Size(), r2.Size())
				}
			})
		}
	}
}

func TestBenchmark_SkillTrap_SearchImproves(t *testing.T) {
	dir := datasetsDir()
	path := filepath.Join(dir, "skill-trap.json")
	dataset, err := loader.LoadDataset(path)
	if err != nil {
		t.Fatalf("failed to load: %v", err)
	}
	travel := convertTravel(dataset.TravelMatrix)

	constructive, _ := application.Optimise(dataset.Events, dataset.Resources, travel, "constructive")
	hillClimb, _ := application.Optimise(dataset.Events, dataset.Resources, travel, "hill-climbing")

	if hillClimb.Size() < constructive.Size() {
		t.Errorf("hill-climbing should assign at least as many items: constructive=%d, hill-climbing=%d",
			constructive.Size(), hillClimb.Size())
	}
}
