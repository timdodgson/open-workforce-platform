// Package loader provides dataset loading for the platform.
//
// It reads JSON files from disk and converts them into domain objects.
package loader

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/event"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/resource"
)

// Dataset holds the loaded business events, resources, travel data, and optional NRP context.
type Dataset struct {
	Events       []event.BusinessEvent
	Resources    []resource.Resource
	TravelMatrix []TravelEntry
	NRPContext   *NRPContext
}

// TravelEntry represents travel time between two locations in the dataset.
type TravelEntry struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Minutes int    `json:"minutes"`
}

// NRPContext carries INRC-II constraint data loaded from a dataset.
type NRPContext struct {
	Contracts            []Contract           `json:"contracts"`
	ShiftTypes           []ShiftTypeEntry     `json:"shiftTypes"`
	ForbiddenSuccessions []ForbiddenSuccession `json:"forbiddenSuccessions"`
	Requests             []Request            `json:"requests"`
	CoverageRequirements []CoverageEntry      `json:"coverageRequirements"`
	WeekendStart         int                  `json:"weekendStart"`
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

// ShiftTypeEntry captures shift type info.
type ShiftTypeEntry struct {
	ID                        string `json:"id"`
	StartMinute               int    `json:"startMinute"`
	EndMinute                 int    `json:"endMinute"`
	MinConsecutiveAssignments int    `json:"minConsecutiveAssignments"`
	MaxConsecutiveAssignments int    `json:"maxConsecutiveAssignments"`
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
	Type      string `json:"type"`
	Weight    int    `json:"weight"`
}

// CoverageEntry captures coverage requirements.
type CoverageEntry struct {
	Day       int    `json:"day"`
	ShiftType string `json:"shiftType"`
	Skill     string `json:"skill"`
	Minimum   int    `json:"minimum"`
	Optimal   int    `json:"optimal"`
}

// LoadDataset reads a JSON file containing business events, resources, and optional travel/NRP data.
func LoadDataset(path string) (Dataset, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Dataset{}, fmt.Errorf("failed to read dataset: %w", err)
	}

	var raw struct {
		BusinessEvents []event.BusinessEvent `json:"businessEvents"`
		Resources      []resource.Resource   `json:"resources"`
		TravelMatrix   []TravelEntry         `json:"travelMatrix"`
		NRPContext     *NRPContext            `json:"nrpContext"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return Dataset{}, fmt.Errorf("failed to parse dataset: %w", err)
	}

	if len(raw.BusinessEvents) == 0 {
		return Dataset{}, fmt.Errorf("dataset contains no business events")
	}

	if len(raw.Resources) == 0 {
		return Dataset{}, fmt.Errorf("dataset contains no resources")
	}

	return Dataset{
		Events:       raw.BusinessEvents,
		Resources:    raw.Resources,
		TravelMatrix: raw.TravelMatrix,
		NRPContext:   raw.NRPContext,
	}, nil
}
