package tsp

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAndEvaluate(t *testing.T) {
	path := filepath.Join("..", "instances", "tsp-5city.json")
	ds, err := LoadDataset(path)
	if err != nil {
		t.Fatal(err)
	}
	p := NewProblem(ds)
	sol, err := p.CreateInitialSolution()
	if err != nil {
		t.Fatal(err)
	}
	if got := p.Evaluate(sol); got <= 0 {
		t.Fatalf("expected positive tour length, got %d", got)
	}
}

func TestLoadDataset_missingFile(t *testing.T) {
	if _, err := LoadDataset("nonexistent.json"); err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadDataset_tooFewCities(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "tsp-*.json")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = f.WriteString(`{"name":"x","cities":[{"x":0,"y":0}]}`)
	f.Close()
	if _, err := LoadDataset(f.Name()); err == nil {
		t.Fatal("expected error for single city")
	}
}
