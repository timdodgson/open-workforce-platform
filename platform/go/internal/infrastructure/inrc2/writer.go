package inrc2

import (
	"encoding/json"
	"fmt"
	"os"
)

// WriteSolution writes an INRC-II solution to a JSON file in official format.
func WriteSolution(sol Solution, path string) error {
	data, err := json.MarshalIndent(sol, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal solution: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

// WriteHistory writes an INRC-II history state to a JSON file in official format.
func WriteHistory(h History, path string) error {
	data, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}
