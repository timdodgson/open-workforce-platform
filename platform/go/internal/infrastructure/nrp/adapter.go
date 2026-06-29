// Package nrp provides an adapter to convert Nurse Rostering Problem input
// into Open Workforce Platform dataset format.
package nrp

import (
	"encoding/json"
	"fmt"
	"os"
)

// NRPInput represents a simplified NRP problem definition.
type NRPInput struct {
	Nurses  []Nurse  `json:"nurses"`
	Shifts  []Shift  `json:"shifts"`
	Demands []Demand `json:"demands"`
}

// Nurse represents a nurse in the NRP.
type Nurse struct {
	ID        string   `json:"id"`
	Skills    []string `json:"skills"`
	Available bool     `json:"available"`
}

// Shift represents a shift definition.
type Shift struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	StartMinute int    `json:"startMinute"`
	EndMinute   int    `json:"endMinute"`
}

// Demand represents the staffing requirement for a shift.
type Demand struct {
	ShiftID       string `json:"shiftId"`
	RequiredSkill string `json:"requiredSkill"`
	Count         int    `json:"count"`
}

// OWPDataset represents the Open Workforce Platform dataset format.
type OWPDataset struct {
	BusinessEvents []OWPEvent    `json:"businessEvents"`
	Resources      []OWPResource `json:"resources"`
	TravelMatrix   []interface{} `json:"travelMatrix"`
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

	// Convert nurses to resources.
	resources := make([]OWPResource, 0, len(input.Nurses))
	for _, nurse := range input.Nurses {
		// Each nurse is available for one full shift (use earliest start to latest end).
		shiftStart := 0
		shiftEnd := 1440
		if len(input.Shifts) > 0 {
			shiftStart = input.Shifts[0].StartMinute
			shiftEnd = input.Shifts[len(input.Shifts)-1].EndMinute
		}
		capacity := shiftEnd - shiftStart

		details, _ := json.Marshal(map[string]interface{}{
			"name":       nurse.ID,
			"skills":     nurse.Skills,
			"capacity":   capacity,
			"available":  nurse.Available,
			"shiftStart": shiftStart,
			"shiftEnd":   shiftEnd,
		})

		resources = append(resources, OWPResource{
			ID:      nurse.ID,
			Type:    "nurse",
			Details: details,
		})
	}

	// Convert demands to business events.
	var events []OWPEvent
	eventIdx := 0

	for _, demand := range input.Demands {
		shift, ok := shiftOf[demand.ShiftID]
		if !ok {
			continue
		}

		duration := shift.EndMinute - shift.StartMinute

		for i := 0; i < demand.Count; i++ {
			eventIdx++
			eventID := fmt.Sprintf("EVT-%03d", eventIdx)

			details, _ := json.Marshal(map[string]interface{}{
				"shift":         shift.Name,
				"requiredSkill": demand.RequiredSkill,
				"duration":      duration,
				"earliestStart": shift.StartMinute,
				"latestFinish":  shift.EndMinute,
				"priority":      100,
			})

			events = append(events, OWPEvent{
				ID:         eventID,
				Type:       "shift.demand",
				OccurredAt: "2026-06-15T00:00:00Z",
				Details:    details,
			})
		}
	}

	return OWPDataset{
		BusinessEvents: events,
		Resources:      resources,
		TravelMatrix:   []interface{}{},
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
