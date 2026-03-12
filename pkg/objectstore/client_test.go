package objectstore

import "testing"

func TestNewAndHealth(t *testing.T) {
	if _, err := New(Config{}); err == nil {
		t.Fatalf("expected validation error")
	}

	c, err := New(Config{Endpoint: "http://minio:9000", Bucket: "archonhq"})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	if err := c.Health(); err != nil {
		t.Fatalf("health: %v", err)
	}

	blobRef := c.BlobRefForArtifact("ten_01", "ws_01", "art_01", "report.json")
	if blobRef == "" {
		t.Fatalf("expected blob ref")
	}
	if !c.IsNamespacedBlobRef(blobRef, "ten_01", "ws_01") {
		t.Fatalf("expected namespaced blob ref to validate")
	}
	if c.IsNamespacedBlobRef(blobRef, "ten_99", "ws_01") {
		t.Fatalf("expected cross-tenant blob ref to fail")
	}

	uploadURL, expiresAt, err := c.UploadURL(blobRef)
	if err != nil {
		t.Fatalf("upload url: %v", err)
	}
	if uploadURL == "" || expiresAt.IsZero() {
		t.Fatalf("expected upload url and expiry")
	}

	downloadURL, dlexp, err := c.DownloadURL(blobRef)
	if err != nil {
		t.Fatalf("download url: %v", err)
	}
	if downloadURL == "" || dlexp.IsZero() {
		t.Fatalf("expected download url and expiry")
	}

	if _, _, err := c.UploadURL(""); err == nil {
		t.Fatalf("expected upload validation error")
	}
	if _, _, err := c.DownloadURL(""); err == nil {
		t.Fatalf("expected download validation error")
	}

	var nilClient *Client
	if _, _, err := nilClient.UploadURL(blobRef); err == nil {
		t.Fatalf("expected nil client health error")
	}
}
