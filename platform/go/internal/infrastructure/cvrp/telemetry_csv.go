package cvrp

import (
	"fmt"
	"strings"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

// ResultsCSVHeader is the canonical header for CVRP dashboard results.csv.
const ResultsCSVHeader = "instance,seed,mode,iterationsPerWorker,maxTotalWorkers,maxConcurrent,initialTemperature,coolingRate,coolingMode,effectiveCoolingRate,minTemperature,lateAcceptanceLen,week,startPenalty,finalPenalty,improvement,hardViolations,softViolations,candidates,accepted,rejected,acceptanceRate,bestIteration,bestWorkerID,workersStarted,branchesCreated,branchesDropped,maxQueueDepth,maxConcurrentSeen,durationMs,saFinalTemp,saTempAtBest,saAcceptedBetter,saAcceptedWorse,saRejectedByProb,lahcAcceptedByCurrent,lahcAcceptedByLate,lahcRejectedByLate,branchesQueued,branchesStarted2,branchesCompleted,winningBranchDepth,workersImproved,workersProducedBest,rejectedNoop,rejectedSkill,rejectedSuccession,rejectedHistory\n"

// DiscoveriesCSVHeader is the canonical header for CVRP NRP-format discoveries.csv.
const DiscoveriesCSVHeader = "run_id,instance,seed,beam_width,iterations,temperature,cooling_mode,timestamp,week,worker_id,beam_path,candidate,elapsed_ms,temperature_at_event,current_penalty,previous_best,new_best,improvement,improvement_pct,event_type,branch_depth,seed_used,accepted_worse_count,hard_reject_count,soft_reject_count,discovery_number,cands_since_previous,time_since_previous_ms,improvement_per_10k,improvement_per_second,post_reheat_improved,post_reheat_best_delta,post_reheat_cands_to_improve,post_reheat_spawned_branch,post_reheat_beat_global,post_reheat_on_winning_lineage\n"

// ResultsCSVParams holds inputs for BuildResultsCSV.
type ResultsCSVParams struct {
	Instance     string
	Seed         int64
	WinnerMode   string
	Iterations   int
	Temperature  float64
	SearchResult optimisation.SearchResult
}

// BuildResultsCSV returns dashboard-compatible results.csv bytes.
func BuildResultsCSV(p ResultsCSVParams) []byte {
	acceptRate := 0.0
	if p.SearchResult.Candidates > 0 {
		acceptRate = float64(p.SearchResult.Accepted) / float64(p.SearchResult.Candidates) * 100
	}
	row := fmt.Sprintf("%s,%d,%s,%d,1,1,%.1f,0,adaptive,0,0.0001,0,1,%d,%d,%d,0,0,%d,%d,%d,%.1f,0,0,1,0,0,0,1,%d,0,0,%d,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0\n",
		p.Instance, p.Seed, p.WinnerMode, p.Iterations,
		p.Temperature,
		p.SearchResult.InitialPenalty, p.SearchResult.BestPenalty, p.SearchResult.InitialPenalty-p.SearchResult.BestPenalty,
		p.SearchResult.Candidates, p.SearchResult.Accepted, p.SearchResult.Rejected, acceptRate,
		p.SearchResult.DurationMs, p.SearchResult.Accepted)
	return []byte(ResultsCSVHeader + row)
}

// DiscoveriesCSVParams holds inputs for BuildDiscoveriesCSV.
type DiscoveriesCSVParams struct {
	RunLabel    string
	Instance    string
	Seed        int64
	Iterations  int
	Temperature float64
	Discoveries []optimisation.Discovery
}

// BuildDiscoveriesCSV returns NRP-format discoveries.csv bytes for CVRP runs.
func BuildDiscoveriesCSV(p DiscoveriesCSVParams) []byte {
	var b strings.Builder
	b.WriteString(DiscoveriesCSVHeader)
	prevCandidate := 0
	prevElapsed := int64(0)
	for i, d := range p.Discoveries {
		candsSince := d.Candidate - prevCandidate
		timeSince := d.ElapsedMs - prevElapsed
		impPct := 0.0
		if d.OldBest > 0 {
			impPct = float64(d.Improvement) / float64(d.OldBest) * 100
		}
		impPer10K := 0.0
		if candsSince > 0 {
			impPer10K = float64(d.Improvement) / float64(candsSince) * 10000
		}
		impPerSec := 0.0
		if timeSince > 0 {
			impPerSec = float64(d.Improvement) / (float64(timeSince) / 1000)
		}
		b.WriteString(fmt.Sprintf("%s,%s,%d,1,%d,%.1f,adaptive,,%d,0,0,%d,%d,0,%d,%d,%d,%d,%.2f,GLOBAL_BEST,0,%d,0,0,0,%d,%d,%d,%.2f,%.2f,0,0,0,0,0,0\n",
			p.RunLabel, p.Instance, p.Seed, p.Iterations, p.Temperature,
			1, d.Candidate, d.ElapsedMs,
			d.NewBest, d.OldBest, d.NewBest, d.Improvement, impPct,
			p.Seed, i+1, candsSince, timeSince, impPer10K, impPerSec))
		prevCandidate = d.Candidate
		prevElapsed = d.ElapsedMs
	}
	return []byte(b.String())
}
