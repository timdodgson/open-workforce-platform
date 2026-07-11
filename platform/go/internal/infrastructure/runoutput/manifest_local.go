package runoutput

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/s3upload"
)

// localManifestPath is platform/web/pfrs-lab/data/manifest.json relative to platform/go.
func localManifestPath() string {
	return filepath.Join(filepath.Dir(DefaultRunsRoot), "manifest.json")
}

// UpdateLocalManifest appends or updates a run entry in the lab manifest (local dev / no S3).
func UpdateLocalManifest(runLabel, algorithm string, penalty int) error {
	if runLabel == "" {
		return nil
	}
	path := localManifestPath()
	var manifest s3upload.Manifest
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &manifest)
	}
	if manifest.Version == "" {
		manifest.Version = "1.0"
	}
	entry := s3upload.ManifestEntry{
		RunID:          runLabel,
		Label:          runLabel,
		Algorithm:      algorithm,
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
		TotalPenalty:   penalty,
		StorageVersion: "local",
	}
	found := false
	for i, existing := range manifest.Runs {
		if existing.RunID == runLabel {
			manifest.Runs[i] = entry
			found = true
			break
		}
	}
	if !found {
		manifest.Runs = append(manifest.Runs, entry)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
