package ilp

import (
	"encoding/json"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2"
)

// ConstraintBreakdownRow is one INRC-II soft constraint aggregate for the dashboard.
type ConstraintBreakdownRow struct {
	ID         string `json:"id"`
	Penalty    int    `json:"penalty"`
	Violations int    `json:"violations"`
}

// ConstraintBreakdownFile is written as constraint-breakdown.json for NRP runs.
type ConstraintBreakdownFile struct {
	TotalPenalty   int                      `json:"totalPenalty"`
	NumWeeks       int                      `json:"numWeeks"`
	HardViolations int                      `json:"hardViolations"`
	Constraints    []ConstraintBreakdownRow `json:"constraints"`
}

var scorerConstraintToID = map[string]string{
	"S1_OptimalCoverage":        "S1",
	"S2_ConsecutiveWorkingDays": "S2",
	"S3_ConsecutiveDaysOff":     "S3",
	"S4_ConsecutiveShiftType":   "S4",
	"S5_ShiftOffRequest":        "S5",
	"S6_CompleteWeekend":        "S6",
	"S7_TotalAssignments":       "S7",
	"S8_TotalWorkingWeekends":   "S8",
}

// BuildConstraintBreakdown aggregates official scorer SoftDetails across weeks.
func BuildConstraintBreakdown(perWeek []inrc2.ScoreResult) ConstraintBreakdownFile {
	penaltyByID := make(map[string]int)
	violationsByID := make(map[string]int)
	totalPenalty := 0
	hardViolations := 0

	for _, wr := range perWeek {
		totalPenalty += wr.SoftPenalty
		hardViolations += wr.HardViolations
		for _, d := range wr.SoftDetails {
			id := scorerConstraintToID[d.Constraint]
			if id == "" {
				continue
			}
			penaltyByID[id] += d.Penalty
			violationsByID[id]++
		}
	}

	order := []string{"S1", "S2", "S3", "S4", "S5", "S6", "S7", "S8"}
	rows := make([]ConstraintBreakdownRow, 0, len(order))
	for _, id := range order {
		rows = append(rows, ConstraintBreakdownRow{
			ID:         id,
			Penalty:    penaltyByID[id],
			Violations: violationsByID[id],
		})
	}

	return ConstraintBreakdownFile{
		TotalPenalty:   totalPenalty,
		NumWeeks:       len(perWeek),
		HardViolations: hardViolations,
		Constraints:    rows,
	}
}

// MarshalConstraintBreakdown returns indented JSON for constraint-breakdown.json.
func MarshalConstraintBreakdown(perWeek []inrc2.ScoreResult) ([]byte, error) {
	return json.MarshalIndent(BuildConstraintBreakdown(perWeek), "", "  ")
}
