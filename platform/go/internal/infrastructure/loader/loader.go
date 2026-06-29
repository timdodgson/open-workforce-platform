// Package loader provides dataset loading for the platform.
//
// It reads JSON files from disk and converts them into domain objects.
package loader

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/event"
)

// LoadEvents reads a JSON file and returns validated BusinessEvent domain objects.
//
// The file must contain a JSON array of business events.
func LoadEvents(path string) ([]event.BusinessEvent, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read dataset: %w", err)
	}

	var events []event.BusinessEvent
	if err := json.Unmarshal(data, &events); err != nil {
		return nil, fmt.Errorf("failed to parse dataset: %w", err)
	}

	if len(events) == 0 {
		return nil, fmt.Errorf("dataset contains no events")
	}

	return events, nil
}
