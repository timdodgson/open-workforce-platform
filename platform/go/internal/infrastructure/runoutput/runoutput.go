// Package runoutput writes benchmark/solver run artifacts to the local runs directory
// and optionally uploads them to S3.
package runoutput

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/s3upload"
)

const (
	// DefaultRunsRoot is the local path for run output (relative to platform/go).
	DefaultRunsRoot = "../web/pfrs-lab/data/runs"

	defaultS3Bucket = "pfrs-research-lab-data"
	defaultS3Region = "eu-west-1"
)

// StorageConfig holds S3 upload settings from CLI flags.
type StorageConfig struct {
	Mode   string
	Bucket string
	Region string
}

// WithDefaults fills empty bucket/region with project defaults.
func (c StorageConfig) WithDefaults() StorageConfig {
	if c.Bucket == "" {
		c.Bucket = defaultS3Bucket
	}
	if c.Region == "" {
		c.Region = defaultS3Region
	}
	return c
}

// EnsureDir creates platform/web/pfrs-lab/data/runs/<label>/ and returns the path.
func EnsureDir(runLabel string) string {
	outputDir := filepath.Join(DefaultRunsRoot, runLabel)
	os.MkdirAll(outputDir, 0755)
	return outputDir
}

// WriteRunMetadata writes run.json under outputDir.
func WriteRunMetadata(outputDir string, meta map[string]interface{}) error {
	return writeJSON(filepath.Join(outputDir, "run.json"), meta)
}

// WriteFile writes a named artifact under outputDir (0644).
func WriteFile(outputDir, filename string, data []byte) error {
	if len(data) == 0 {
		return nil
	}
	return os.WriteFile(filepath.Join(outputDir, filename), data, 0644)
}

// WriteILPBenchmark writes run.json, ilp-benchmark.json, and optional solution.json.
func WriteILPBenchmark(outputDir string, meta map[string]interface{}, benchResult interface{}, solutionJSON []byte) error {
	if err := WriteRunMetadata(outputDir, meta); err != nil {
		return err
	}
	benchJSON, err := json.MarshalIndent(benchResult, "", "  ")
	if err != nil {
		return err
	}
	if err := WriteFile(outputDir, "ilp-benchmark.json", benchJSON); err != nil {
		return err
	}
	return WriteFile(outputDir, "solution.json", solutionJSON)
}

// Upload uploads a completed run directory to S3 when storage mode is s3.
func Upload(cfg StorageConfig, runLabel, outputDir, algorithm string, penalty int) error {
	cfg = cfg.WithDefaults()
	return s3upload.UploadRun(cfg.Mode, s3upload.UploadRunConfig{
		RunLabel:  runLabel,
		RunDir:    outputDir,
		Algorithm: algorithm,
		Penalty:   penalty,
		Bucket:    cfg.Bucket,
		Region:    cfg.Region,
	})
}

func writeJSON(path string, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
