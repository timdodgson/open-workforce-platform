package inrc2

import "testing"

func TestRankAlgorithmBenchmarkResults(t *testing.T) {
	results := map[string]*AlgorithmBenchmarkResult{
		"sa":   {Algorithm: "sa", TotalPenalty: 200, TotalHard: 0},
		"tabu": {Algorithm: "tabu", TotalPenalty: 150, TotalHard: 0},
		"bad":  {Algorithm: "bad", TotalPenalty: 100, TotalHard: 2},
	}
	algs := []string{"sa", "tabu", "bad"}

	valid, invalid := RankAlgorithmBenchmarkResults(algs, results)
	if len(valid) != 2 {
		t.Fatalf("valid count = %d, want 2", len(valid))
	}
	if valid[0].Algorithm != "tabu" || valid[1].Algorithm != "sa" {
		t.Fatalf("valid order = %s, %s; want tabu, sa", valid[0].Algorithm, valid[1].Algorithm)
	}
	if len(invalid) != 1 || invalid[0].Algorithm != "bad" {
		t.Fatalf("invalid = %+v", invalid)
	}
}
