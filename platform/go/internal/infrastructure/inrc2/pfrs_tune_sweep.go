package inrc2

import (
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2/legacysearch"
)

// TuningSweepHooks provides optional callbacks for CLI progress reporting.
type TuningSweepHooks struct {
	OnSeedStart func(entry TuningGridEntry, seed int64)
	OnWeekLine  func(week int, penalty, workers, branches, candidates int, durationMs int64)
	OnSeedDone  func(result TuningResult, seed int64)
}

// TuningSweepParams configures a multi-grid, multi-seed PFRS tuning sweep.
type TuningSweepParams struct {
	Scenario    Scenario
	WeekFiles   []string
	NumWeeks    int
	History     History
	Grid        []TuningGridEntry
	Seeds       []int64
	AlgProfile  legacysearch.AlgorithmProfile
	BuildConfig func(entry TuningGridEntry, seed int64, currentWeek *int) PFRSConfig
	Hooks       TuningSweepHooks
}

// TuningSweepResult holds aggregated sweep output for artifact finalization.
type TuningSweepResult struct {
	MultiResults []MultiSeedResult
	AuditRows    []WeekAuditRow
	Bundles      []WeekAuditBundle
}

// RunTuningSweep executes each grid entry across all seeds and weeks.
func RunTuningSweep(p TuningSweepParams) TuningSweepResult {
	var out TuningSweepResult

	for _, entry := range p.Grid {
		var seedResults []TuningResult

		for _, seed := range p.Seeds {
			if p.Hooks.OnSeedStart != nil {
				p.Hooks.OnSeedStart(entry, seed)
			}

			currentWeek := 0
			config := p.BuildConfig(entry, seed, &currentWeek)

			result := TuningResult{Entry: entry, Seed: seed}
			currentHist := p.History
			var weekAuditBundles []WeekAuditBundle
			weekData := make([]WeekData, 0, p.NumWeeks)
			weekSols := make([]Solution, 0, p.NumWeeks)

			for w := 0; w < p.NumWeeks; w++ {
				currentWeek = w
				wd, err := LoadWeekData(p.WeekFiles[w])
				if err != nil {
					continue
				}

				var weekAudit PFRSAudit
				config.OnAudit = func(a PFRSAudit) { weekAudit = a }

				sol, stats, scoreResult, err := SolveWeekPFRS(p.Scenario, wd, currentHist, config)
				if err != nil {
					result.TotalHard++
					cSol, _, _ := SolveWeek(p.Scenario, wd, currentHist, "constructive", p.AlgProfile)
					currentHist = UpdateHistory(p.Scenario, currentHist, cSol)
					continue
				}

				// Progress lines stay week-local; official total is MultiStage below.
				result.TotalHard += scoreResult.HardViolations
				result.TotalAssign += len(sol.Assignments)
				result.TotalMs += stats.DurationMs
				result.TotalCands += stats.CandidatesEvaluated
				weekData = append(weekData, wd)
				weekSols = append(weekSols, sol)

				if p.Hooks.OnWeekLine != nil {
					p.Hooks.OnWeekLine(w+1, scoreResult.SoftPenalty,
						stats.WorkersStarted, stats.BranchesCreated,
						stats.CandidatesEvaluated, stats.DurationMs)
				}

				startPenalty := Worker0StartPenalty(weekAudit)
				row := BuildWeekAuditRow(p.Scenario.ID, config, w+1, startPenalty, stats, scoreResult, weekAudit)
				out.AuditRows = append(out.AuditRows, row)

				if len(weekAudit.Workers) > 0 {
					weekAuditBundles = append(weekAuditBundles, WeekAuditBundle{
						Week:                w + 1,
						GlobalBestAtSpawn:   startPenalty,
						TotalWorkersStarted: stats.WorkersStarted,
						ActiveFamilies:      1,
						Workers:             weekAudit.Workers,
					})
				}

				currentHist = UpdateHistory(p.Scenario, currentHist, sol)
			}

			if len(weekSols) == p.NumWeeks {
				ms := ScoreMultiStage(p.Scenario, weekData, p.History, weekSols)
				result.TotalPenalty = ms.TotalObjective
				result.TotalSoft = len(ms.SoftDetails)
				result.TotalHard = ms.HardViolations
			}
			result.Valid = result.TotalHard == 0
			seedResults = append(seedResults, result)
			out.Bundles = append(out.Bundles, weekAuditBundles...)

			if p.Hooks.OnSeedDone != nil {
				p.Hooks.OnSeedDone(result, seed)
			}
		}

		out.MultiResults = append(out.MultiResults, AggregateSeeds(entry, seedResults))
	}

	return out
}
