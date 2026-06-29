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

// Dataset holds the loaded business events, resources, and travel data from a dataset file.
type Dataset struct {
	Events       []event.BusinessEvent
	Resources    []resource.Resource
	TravelMatrix []TravelEntry
}

// TravelEntry represents travel time between two locations in the dataset.
type TravelEntry struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Minutes int    `json:"minutes"`
}

// LoadDataset reads a JSON file containing business events, resources, and optional travel data.
func LoadDataset(path string) (Dataset, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Dataset{}, fmt.Errorf("failed to read dataset: %w", err)
	}

	var raw struct {
		BusinessEvents []event.BusinessEvent `json:"businessEvents"`
		Resources      []resource.Resource   `json:"resources"`
		TravelMatrix   []TravelEntry         `json:"travelMatrix"`
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
	}, nil
}
