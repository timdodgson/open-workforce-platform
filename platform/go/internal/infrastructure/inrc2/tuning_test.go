package inrc2_test

import (
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2"
)

func TestGenerateGrid_Size(t *testing.T) {
	iterations := []int{30000, 60000}
	workers := []int{16, 32}
	temps := []float64{1.0, 2.0}
	rates := []float64{0.0009, 0.0005}

	grid := inrc2.GenerateGrid(iterations, workers, temps, rates)

	// 2 * 2 * 2 * 2 = 16 combinations.
	expected := 2 * 2 * 2 * 2
	if len(grid) != expected {
		t.Errorf("expected %d grid entries, got %d", expected, len(grid))
	}
}

func TestGenerateGrid_DeterministicOrder(t *testing.T) {
	iterations := []int{30000, 60000}
	workers := []int{16}
	temps := []float64{1.0, 5.0}
	rates := []float64{0.0009}

	grid := inrc2.GenerateGrid(iterations, workers, temps, rates)

	// Should be: (30000,16,1.0,0.0009), (30000,16,5.0,0.0009), (60000,16,1.0,0.0009), (60000,16,5.0,0.0009)
	if len(grid) != 4 {
		t.Fatalf("expected 4, got %d", len(grid))
	}
	if grid[0].IterationsPerWorker != 30000 || grid[0].InitialTemperature != 1.0 {
		t.Errorf("first entry wrong: %+v", grid[0])
	}
	if grid[1].IterationsPerWorker != 30000 || grid[1].InitialTemperature != 5.0 {
		t.Errorf("second entry wrong: %+v", grid[1])
	}
	if grid[2].IterationsPerWorker != 60000 || grid[2].InitialTemperature != 1.0 {
		t.Errorf("third entry wrong: %+v", grid[2])
	}
	if grid[3].IterationsPerWorker != 60000 || grid[3].InitialTemperature != 5.0 {
		t.Errorf("fourth entry wrong: %+v", grid[3])
	}
}

func TestRankResults_ValidFirst(t *testing.T) {
	results := []inrc2.TuningResult{
		{Entry: inrc2.TuningGridEntry{IterationsPerWorker: 30000}, TotalPenalty: 500, TotalHard: 2, Valid: false},
		{Entry: inrc2.TuningGridEntry{IterationsPerWorker: 60000}, TotalPenalty: 300, TotalHard: 0, Valid: true},
		{Entry: inrc2.TuningGridEntry{IterationsPerWorker: 100000}, TotalPenalty: 200, TotalHard: 0, Valid: true},
	}

	valid, invalid := inrc2.RankResults(results)

	if len(valid) != 2 {
		t.Fatalf("expected 2 valid, got %d", len(valid))
	}
	if len(invalid) != 1 {
		t.Fatalf("expected 1 invalid, got %d", len(invalid))
	}

	// Valid ranked by penalty ascending.
	if valid[0].TotalPenalty != 200 {
		t.Errorf("expected best penalty 200, got %d", valid[0].TotalPenalty)
	}
	if valid[1].TotalPenalty != 300 {
		t.Errorf("expected second penalty 300, got %d", valid[1].TotalPenalty)
	}
}

func TestRankResults_PenaltyTiebreak(t *testing.T) {
	results := []inrc2.TuningResult{
		{Entry: inrc2.TuningGridEntry{IterationsPerWorker: 60000}, TotalPenalty: 200, TotalHard: 0, TotalMs: 500, Valid: true},
		{Entry: inrc2.TuningGridEntry{IterationsPerWorker: 30000}, TotalPenalty: 200, TotalHard: 0, TotalMs: 100, Valid: true},
	}

	valid, _ := inrc2.RankResults(results)

	// Same penalty — faster runtime first.
	if valid[0].TotalMs != 100 {
		t.Errorf("expected faster result first (100ms), got %dms", valid[0].TotalMs)
	}
}

func TestGenerateGrid_FullSpec(t *testing.T) {
	// The actual grid from the spec.
	iterations := []int{30000, 60000, 100000}
	workers := []int{16, 32}
	temps := []float64{1.0, 2.0, 5.0}
	rates := []float64{0.0009, 0.0005, 0.0001}

	grid := inrc2.GenerateGrid(iterations, workers, temps, rates)

	// 3 * 2 * 3 * 3 = 54 combinations.
	expected := 3 * 2 * 3 * 3
	if len(grid) != expected {
		t.Errorf("expected %d grid entries, got %d", expected, len(grid))
	}
}

func TestAggregateSeeds(t *testing.T) {
	entry := inrc2.TuningGridEntry{IterationsPerWorker: 100000, MaxTotalWorkers: 32, InitialTemperature: 1.0, CoolingRate: 0.0009}
	results := []inrc2.TuningResult{
		{Entry: entry, Seed: 42, TotalPenalty: 200, TotalSoft: 10, TotalMs: 100, TotalCands: 5000, Valid: true},
		{Entry: entry, Seed: 101, TotalPenalty: 250, TotalSoft: 12, TotalMs: 120, TotalCands: 5200, Valid: true},
		{Entry: entry, Seed: 202, TotalPenalty: 180, TotalSoft: 9, TotalMs: 90, TotalCands: 4800, Valid: true},
	}

	ms := inrc2.AggregateSeeds(entry, results)

	if ms.Seeds != 3 {
		t.Errorf("expected 3 seeds, got %d", ms.Seeds)
	}
	if ms.BestPen != 180 {
		t.Errorf("expected best 180, got %d", ms.BestPen)
	}
	if ms.BestSeed != 202 {
		t.Errorf("expected best seed 202, got %d", ms.BestSeed)
	}
	if ms.WorstPen != 250 {
		t.Errorf("expected worst 250, got %d", ms.WorstPen)
	}
	// Average: (200+250+180)/3 = 210
	if ms.AvgPen != 210 {
		t.Errorf("expected avg 210, got %d", ms.AvgPen)
	}
	if !ms.AllValid {
		t.Error("expected all valid")
	}
	if ms.TotalCands != 15000 {
		t.Errorf("expected 15000 candidates, got %d", ms.TotalCands)
	}
}

func TestRankMultiSeedResults_ByAvgPenalty(t *testing.T) {
	results := []inrc2.MultiSeedResult{
		{Entry: inrc2.TuningGridEntry{IterationsPerWorker: 30000}, AvgPen: 300, AllValid: true, AvgMs: 50},
		{Entry: inrc2.TuningGridEntry{IterationsPerWorker: 60000}, AvgPen: 200, AllValid: true, AvgMs: 100},
		{Entry: inrc2.TuningGridEntry{IterationsPerWorker: 100000}, AvgPen: 150, AllValid: false, ValidCount: 2},
	}

	valid, invalid := inrc2.RankMultiSeedResults(results)

	if len(valid) != 2 {
		t.Fatalf("expected 2 valid, got %d", len(valid))
	}
	if len(invalid) != 1 {
		t.Fatalf("expected 1 invalid, got %d", len(invalid))
	}

	// Ranked by avg penalty ascending.
	if valid[0].AvgPen != 200 {
		t.Errorf("expected first avg 200, got %d", valid[0].AvgPen)
	}
	if valid[1].AvgPen != 300 {
		t.Errorf("expected second avg 300, got %d", valid[1].AvgPen)
	}
}

