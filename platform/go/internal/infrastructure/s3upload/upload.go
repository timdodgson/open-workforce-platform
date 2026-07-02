package s3upload

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// ManifestEntry holds lightweight metadata for the run manifest.
type ManifestEntry struct {
	RunID        string  `json:"runId"`
	Label        string  `json:"label"`
	Algorithm    string  `json:"algorithm"`
	Timestamp    string  `json:"timestamp"`
	TotalPenalty int     `json:"totalPenalty"`
	BeamHealth   int     `json:"beamHealth"`
	StorageVersion string `json:"storageVersion"`
}

// Manifest holds the full manifest.json content.
type Manifest struct {
	Version string          `json:"version"`
	Runs    []ManifestEntry `json:"runs"`
}

// Client wraps the S3 client for uploading run data.
type Client struct {
	s3     *s3.Client
	bucket string
}

// NewClient creates an S3 upload client using the default AWS credential chain.
func NewClient(bucket, region string) (*Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &Client{
		s3:     s3.NewFromConfig(cfg),
		bucket: bucket,
	}, nil
}

// UploadFile uploads a single file to the run's S3 prefix.
func (c *Client) UploadFile(runId, filename, content string) error {
	key := fmt.Sprintf("runs/%s/%s", runId, filename)
	contentType := "text/plain"
	if strings.HasSuffix(filename, ".json") {
		contentType = "application/json"
	} else if strings.HasSuffix(filename, ".csv") {
		contentType = "text/csv"
	}

	_, err := c.s3.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      &c.bucket,
		Key:         &key,
		Body:        strings.NewReader(content),
		ContentType: &contentType,
	})
	return err
}

// UploadLocalFile reads a local file and uploads it to S3.
func (c *Client) UploadLocalFile(runId, filename, localPath string) error {
	data, err := os.ReadFile(localPath)
	if err != nil {
		return err
	}
	return c.UploadFile(runId, filename, string(data))
}

// UpdateManifest reads the existing manifest, adds/updates the entry, and writes it back.
func (c *Client) UpdateManifest(entry ManifestEntry) error {
	manifest, err := c.loadManifest()
	if err != nil {
		manifest = &Manifest{Version: "1.0"}
	}

	// Update or append.
	found := false
	for i, existing := range manifest.Runs {
		if existing.RunID == entry.RunID {
			manifest.Runs[i] = entry
			found = true
			break
		}
	}
	if !found {
		manifest.Runs = append(manifest.Runs, entry)
	}

	// Write back.
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}

	key := "manifest.json"
	contentType := "application/json"
	_, err = c.s3.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      &c.bucket,
		Key:         &key,
		Body:        strings.NewReader(string(data)),
		ContentType: &contentType,
	})
	return err
}

func (c *Client) loadManifest() (*Manifest, error) {
	key := "manifest.json"
	resp, err := c.s3.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: &c.bucket,
		Key:    &key,
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var manifest Manifest
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return nil, err
	}
	return &manifest, nil
}

// Timestamp returns a formatted timestamp for manifest entries.
func Timestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}
