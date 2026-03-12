package objectstore

import (
	"fmt"
	"net/url"
	"path"
	"strings"
	"time"
)

// Config controls the object-store client wiring.
type Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	Region    string
}

type Client struct {
	Config Config
}

func New(cfg Config) (*Client, error) {
	if cfg.Endpoint == "" || cfg.Bucket == "" {
		return nil, fmt.Errorf("object store endpoint and bucket are required")
	}
	return &Client{Config: cfg}, nil
}

func (c *Client) Health() error {
	if c == nil || c.Config.Endpoint == "" {
		return fmt.Errorf("object store not configured")
	}
	return nil
}

func (c *Client) BlobRefForArtifact(tenantID, workspaceID, artifactID, fileName string) string {
	name := sanitizeFileName(fileName)
	if name == "" {
		name = artifactID
	}
	return fmt.Sprintf("s3://%s/%s/%s/%s_%s", c.Config.Bucket, tenantID, workspaceID, artifactID, name)
}

func (c *Client) IsNamespacedBlobRef(blobRef, tenantID, workspaceID string) bool {
	prefix := fmt.Sprintf("s3://%s/%s/%s/", c.Config.Bucket, tenantID, workspaceID)
	return strings.HasPrefix(blobRef, prefix)
}

func (c *Client) UploadURL(blobRef string) (string, time.Time, error) {
	if err := c.Health(); err != nil {
		return "", time.Time{}, err
	}
	if blobRef == "" {
		return "", time.Time{}, fmt.Errorf("blob_ref is required")
	}
	expiresAt := time.Now().UTC().Add(15 * time.Minute)
	return fmt.Sprintf("%s/upload?blob_ref=%s&expires_at=%d", strings.TrimRight(c.Config.Endpoint, "/"), url.QueryEscape(blobRef), expiresAt.Unix()), expiresAt, nil
}

func (c *Client) DownloadURL(blobRef string) (string, time.Time, error) {
	if err := c.Health(); err != nil {
		return "", time.Time{}, err
	}
	if blobRef == "" {
		return "", time.Time{}, fmt.Errorf("blob_ref is required")
	}
	expiresAt := time.Now().UTC().Add(15 * time.Minute)
	return fmt.Sprintf("%s/download?blob_ref=%s&expires_at=%d", strings.TrimRight(c.Config.Endpoint, "/"), url.QueryEscape(blobRef), expiresAt.Unix()), expiresAt, nil
}

func sanitizeFileName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	name = path.Base(name)
	name = strings.ReplaceAll(name, "..", "")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	return name
}
