package main

import (
	"encoding/json"
	"os"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/runoutput"
)

// storageConfig holds S3 upload settings parsed from CLI flags.
type storageConfig struct {
	Mode   string
	Bucket string
	Region string
}

// parseStorageConfig reads storage flags. When pfrsPrefix is true, prefers --pfrs-storage etc.
// Both --storage and --pfrs-storage are accepted on all commands.
func parseStorageConfig(args []string, pfrsPrefix bool) storageConfig {
	primary := ""
	alt := "pfrs-"
	if pfrsPrefix {
		primary = "pfrs-"
		alt = ""
	}
	cfg := storageConfig{
		Mode: firstNonEmpty(
			parseStringFlag(args, "--"+primary+"storage"),
			parseStringFlag(args, "--"+alt+"storage"),
		),
	}
	ro := runoutput.StorageConfig{
		Bucket: firstNonEmpty(
			parseStringFlag(args, "--"+primary+"s3-bucket"),
			parseStringFlag(args, "--"+alt+"s3-bucket"),
		),
		Region: firstNonEmpty(
			parseStringFlag(args, "--"+primary+"s3-region"),
			parseStringFlag(args, "--"+alt+"s3-region"),
		),
	}.WithDefaults()
	cfg.Bucket = ro.Bucket
	cfg.Region = ro.Region
	return cfg
}

func runoutputStorage(cfg storageConfig) runoutput.StorageConfig {
	return runoutput.StorageConfig{
		Mode:   cfg.Mode,
		Bucket: cfg.Bucket,
		Region: cfg.Region,
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// parseRunLabelFlag reads --run-label or --pfrs-run-label.
func parseRunLabelFlag(args []string, pfrsPrefix bool) string {
	if pfrsPrefix {
		return parseStringFlag(args, "--pfrs-run-label")
	}
	return parseStringFlag(args, "--run-label")
}

// ensureRunOutputDir creates platform/web/pfrs-lab/data/runs/<label>/ and returns the path.
func ensureRunOutputDir(runLabel string) string {
	return runoutput.EnsureDir(runLabel)
}

// writeJSONFile marshals v with indentation and writes to path (0644). Ignores write errors (matches prior behaviour).
func writeJSONFile(path string, v interface{}) {
	data, _ := json.MarshalIndent(v, "", "  ")
	os.WriteFile(path, data, 0644)
}

// writeRunMetadata writes run.json under outputDir.
func writeRunMetadata(outputDir string, meta map[string]interface{}) {
	_ = runoutput.WriteRunMetadata(outputDir, meta)
}

// uploadRunOutput uploads a completed run directory to S3 when configured.
func uploadRunOutput(cfg storageConfig, runLabel, outputDir, algorithm string, penalty int) {
	_ = runoutput.Upload(runoutputStorage(cfg), runLabel, outputDir, algorithm, penalty)
}
