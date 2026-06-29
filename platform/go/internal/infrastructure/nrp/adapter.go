// Package nrp provides an adapter to convert Nurse Rostering Problem input
// into Open Workforce Platform dataset format.
package nrp

import (
	"encoding/json"
	"fmt"
	"os"
)

// NRPInput represents an INRC-II-style NRP problem definition.
type NRPInput struct {
	Nurses               []Nurse              `json:"nurses"`
	Shifts               []Shift              `json:"shifts"`
	Demands              []Demand             `json:"demands"`
	Contracts            []Contract           `json:"contracts"`
	ForbiddenSuccessions []ForbiddenSuccession `json:"forbiddenSuccessions"`
	Requests             []Request            `json:"requests"`
	WeekendStart         int                  `json:"weekendStart"` // day number where weekend starts (default 6)
}

// Nurse represents a nurse in the NRP.
type Nurse struct {
	ID         string   `json:"id"`
	Skills     []string `json:"skills"`
	Available  bool     `json:"available"`
	ContractID string   `json:"contractId"`
}

// Shift represents a shift definition.
type Shift struct {
	ID                       string `json:"id"`
	Name                     string `json:"name"`
	StartMinute              int    `json:"startMinute"`
	EndMinute                int    `json:"endMinute"`
	MinConsecutiveAssignments int    `json:"minConsecutiveAssignments"`
	MaxConsecutiveAssignments int    `json:"maxConsecutiveAssignments"`
}

// Demand represents the staffing requirement for a shift on a day.
type Demand struct {
	Day           int    `json:"day"`
	ShiftID       string `json:"shiftId"`
	RequiredSkill string `json:"requiredSkill"`
	Minimum       int    `json:"minimum"`
	Optimal       int    `json:"optimal"`
}

// Contract defines nurse employment constraints.
type Contract struct {
	ID                        string `json:"id"`
	MinAssignments            int    `json:"minAssignments"`
	MaxAssignments            int    `json:"maxAssignments"`
	MinConsecutiveWorkingDays int    `json:"minConsecutiveWorkingDays"`
	MaxConsecutiveWorkingDays int    `json:"maxConsecutiveWorkingDays"`
	MinConsecutiveDaysOff     int    `json:"minConsecutiveDaysOff"`
	MaxConsecutiveDaysOff     int    `json:"maxConsecutiveDaysOff"`
	MaxWorkingWeekends        int    `json:"maxWorkingWeekends"`
	CompleteWeekend           bool   `json:"completeWeekend"`
}

// ForbiddenSuccession defines an illegal shift type transition.
type ForbiddenSuccession struct {
	PrecedingShift string `json:"precedingShift"`
	SuccessorShift string `json:"successorShift"`
}

// Request represents a nurse scheduling preference.
type Request struct {
	NurseID   string `json:"nurseId"`
	Day       int    `json:"day"`
	ShiftType string `json:"shiftType"`
	Type      string `json:"type"` // "shiftOn", "shiftOff", "dayOn", "dayOff"
	Weight    int    `json:"weight"`
}

// OWPDataset represents the Open Workforce Platform dataset format.
type OWPDataset struct {
	BusinessEvents []OWPEvent    `json:"businessEvents"`
	Resources      []OWPResource `json:"resources"`
	TravelMatrix   []interface{} `json:"travelMatrix"`
	NRPContext     *NRPContext   `json:"nrpContext,omitempty"`
}

// NRPContext carries INRC-II constraint data through the dataset.
type NRPContext struct {
	Contracts            []Contract           `json:"contracts"`
	ShiftTypes           []ShiftTypeEntry     `json:"shiftTypes"`
	ForbiddenSuccessions []ForbiddenSuccession `json:"forbiddenSuccessions"`
	Requests             []Request            `json:"requests"`
	CoverageRequirements []CoverageEntry      `json:"coverageRequirements"`
	WeekendStart         int                  `json:"weekendStart"`
}

// ShiftTypeEntry captures shift type info for the OWP context.
type ShiftTypeEntry struct {
	ID                        string `json:"id"`
	StartMinute               int    `json:"startMinute"`
	EndMinute                 int    `json:"endMinute"`
	MinConsecutiveAssignments int    `json:"minConsecutiveAssignments"`
	MaxConsecutiveAssignments int    `json:"maxConsecutiveAssignments"`
}

// CoverageEntry captures coverage requirements.
type CoverageEntry struct {
	Day       int    `json:"day"`
	ShiftType string `json:"shiftType"`
	Skill     string `json:"skill"`
	Minimum   int    `json:"minimum"`
	Optimal   int    `json:"optimal"`
}

// OWPEvent represents a business event in OWP format.
type OWPEvent struct {
	ID         string          `json:"id"`
	Type       string          `json:"type"`
	OccurredAt string          `json:"occurredAt"`
	Details    json.RawMessage `json:"details"`
}

// OWPResource represents a resource in OWP format.
type OWPResource struct {
	ID      string          `json:"id"`
	Type    string          `json:"type"`
	Details json.RawMessage `json:"details"`
}

// LoadNRP reads an NRP JSON file.
func LoadNRP(path string) (NRPInput, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return NRPInput{}, fmt.Errorf("failed to read NRP file: %w", err)
	}

	var input NRPInput
	if err := json.Unmarshal(data, &input); err != nil {
		return NRPInput{}, fmt.Errorf("failed to parse NRP file: %w", err)
	}

	return input, nil
}

// Convert transforms an NRP input into an OWP dataset.
func Convert(input NRPInput) OWPDataset {
	// Build shift lookup.
	shiftOf := make(map[string]Shift, len(input.Shifts))
	for _, s := range input.Shifts {
		shiftOf[s.ID] = s
	}

	// Determine resource shift window: earliest start to latest end across all shifts.
	shiftStart := 0
	shiftEnd := 1440
	if len(input.Shifts) > 0 {
		shiftStart = input.Shifts[0].StartMinute
		shiftEnd = input.Shifts[len(input.Shifts)-1].EndMinute
	}

	// Convert nurses to resources.
	resources := make([]OWPResource, 0, len(input.Nurses))
	for _, nurse := range input.Nurses {
		capacity := shiftEnd - shiftStart

		details, _ := json.Marshal(map[string]interface{}{
			"name":       nurse.ID,
			"skills":     nurse.Skills,
			"capacity":   capacity,
			"available":  nurse.Available,
			"shiftStart": shiftStart,
			"shiftEnd":   shiftEnd,
			"contractId": nurse.ContractID,
		})

		resources = append(resources, OWPResource{
			ID:      nurse.ID,
			Type:    "nurse",
			Details: details,
		})
	}

	// Convert demands to business events.
	// Each minimum unit is a mandatory demand work item.
	// Each optimal-above-minimum unit is an optional demand work item.
	var events []OWPEvent
	var coverageReqs []CoverageEntry
	eventIdx := 0

	for _, demand := range input.Demands {
		shift, ok := shiftOf[demand.ShiftID]
		if !ok {
			continue
		}

		duration := shift.EndMinute - shift.StartMinute
		day := demand.Day
		if day == 0 {
			day = 1 // backward compat: old format without explicit day
		}

		demandGroup := fmt.Sprintf("day%d-%s-%s", day, demand.ShiftID, demand.RequiredSkill)

		minimum := demand.Minimum
		if minimum == 0 {
			// Backward compat: old format used "count" which mapped to minimum.
			minimum = demand.Optimal
		}
		optimal := demand.Optimal
		if optimal < minimum {
			optimal = minimum
		}

		coverageReqs = append(coverageReqs, CoverageEntry{
			Day:       day,
			ShiftType: demand.ShiftID,
			Skill:     demand.RequiredSkill,
			Minimum:   minimum,
			Optimal:   optimal,
		})

		// Generate mandatory work items for minimum coverage.
		for i := 0; i < minimum; i++ {
			eventIdx++
			eventID := fmt.Sprintf("EVT-%03d", eventIdx)

			details, _ := json.Marshal(map[string]interface{}{
				"shift":         shift.Name,
				"shiftType":     demand.ShiftID,
				"requiredSkill": demand.RequiredSkill,
				"duration":      duration,
				"earliestStart": shift.StartMinute,
				"latestFinish":  shift.EndMinute,
				"priority":      100,
				"day":           day,
				"mandatory":     true,
				"demandGroup":   demandGroup,
			})

			events = append(events, OWPEvent{
				ID:         eventID,
				Type:       "shift.demand",
				OccurredAt: "2026-06-15T00:00:00Z",
				Details:    details,
			})
		}

		// Generate optional work items for optimal coverage above minimum.
		for i := minimum; i < optimal; i++ {
			eventIdx++
			eventID := fmt.Sprintf("EVT-%03d", eventIdx)

			details, _ := json.Marshal(map[string]interface{}{
				"shift":         shift.Name,
				"shiftType":     demand.ShiftID,
				"requiredSkill": demand.RequiredSkill,
				"duration":      duration,
				"earliestStart": shift.StartMinute,
				"latestFinish":  shift.EndMinute,
				"priority":      50,
				"day":           day,
				"mandatory":     false,
				"demandGroup":   demandGroup,
			})

			events = append(events, OWPEvent{
				ID:         eventID,
				Type:       "shift.demand",
				OccurredAt: "2026-06-15T00:00:00Z",
				Details:    details,
			})
		}
	}

	// Build NRP context.
	weekendStart := input.WeekendStart
	if weekendStart == 0 {
		weekendStart = 6 // default: Saturday is day 6 in a Mon-start week
	}

	var shiftTypeEntries []ShiftTypeEntry
	for _, s := range input.Shifts {
		shiftTypeEntries = append(shiftTypeEntries, ShiftTypeEntry{
			ID:                        s.ID,
			StartMinute:               s.StartMinute,
			EndMinute:                 s.EndMinute,
			MinConsecutiveAssignments: s.MinConsecutiveAssignments,
			MaxConsecutiveAssignments: s.MaxConsecutiveAssignments,
		})
	}

	var nrpCtx *NRPContext
	if len(input.Contracts) > 0 || len(input.ForbiddenSuccessions) > 0 || len(input.Requests) > 0 || len(coverageReqs) > 0 {
		nrpCtx = &NRPContext{
			Contracts:            input.Contracts,
			ShiftTypes:           shiftTypeEntries,
			ForbiddenSuccessions: input.ForbiddenSuccessions,
			Requests:             input.Requests,
			CoverageRequirements: coverageReqs,
			WeekendStart:         weekendStart,
		}
	}

	return OWPDataset{
		BusinessEvents: events,
		Resources:      resources,
		TravelMatrix:   []interface{}{},
		NRPContext:     nrpCtx,
	}
}

// WriteDataset writes an OWP dataset to a JSON file.
func WriteDataset(dataset OWPDataset, path string) error {
	data, err := json.MarshalIndent(dataset, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal dataset: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}
