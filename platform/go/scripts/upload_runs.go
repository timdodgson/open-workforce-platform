//go:build ignore

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/s3upload"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: go run scripts/upload_runs.go <run-id> [run-id...]")
		os.Exit(1)
	}
	base := filepath.Join("..", "web", "pfrs-lab", "data", "runs")
	for _, id := range os.Args[1:] {
		dir := filepath.Join(base, id)
		if err := s3upload.UploadRun("s3", s3upload.UploadRunConfig{
			RunLabel:  id,
			RunDir:    dir,
			Algorithm: "upload",
			Penalty:   0,
		}); err != nil {
			fmt.Fprintf(os.Stderr, "FAIL %s: %v\n", id, err)
			os.Exit(1)
		}
	}
}
