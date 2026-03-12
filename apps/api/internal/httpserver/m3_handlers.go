package httpserver

import (
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/MikeS071/archonhq/pkg/apierrors"
)

type artifactUploadURLRequest struct {
	ArtifactID  string `json:"artifact_id,omitempty"`
	WorkspaceID string `json:"workspace_id"`
	FileName    string `json:"file_name"`
	MediaType   string `json:"media_type"`
	SizeBytes   int64  `json:"size_bytes"`
}

func (s *Server) handleArtifactUploadURLV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !s.validateNodeCredential(w, r, actor, corrID) {
		return
	}

	var req artifactUploadURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload.", corrID, nil)
		return
	}
	req.WorkspaceID = strings.TrimSpace(req.WorkspaceID)
	req.FileName = strings.TrimSpace(req.FileName)
	req.MediaType = strings.TrimSpace(req.MediaType)
	if req.WorkspaceID == "" || req.FileName == "" || req.MediaType == "" || req.SizeBytes <= 0 {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "workspace_id, file_name, media_type, and positive size_bytes are required.", corrID, nil)
		return
	}

	const workspaceTenantQ = "SELECT tenant_id FROM workspaces WHERE workspace_id = $1"
	var workspaceTenantID string
	if err := s.postgres.DB.QueryRowContext(r.Context(), workspaceTenantQ, req.WorkspaceID).Scan(&workspaceTenantID); err != nil {
		if err == sql.ErrNoRows {
			apierrors.Write(w, http.StatusNotFound, "workspace_not_found", "Workspace not found.", corrID, nil)
			return
		}
		apierrors.Write(w, http.StatusInternalServerError, "workspace_lookup_failed", "Failed to validate workspace.", corrID, nil)
		return
	}
	if !s.ensureTenantAccess(actor, workspaceTenantID) {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Cross-tenant write is not allowed.", corrID, nil)
		return
	}

	artifactID := strings.TrimSpace(req.ArtifactID)
	if artifactID == "" {
		artifactID = "art_" + randomID(6)
	}
	blobRef := s.objectStore.BlobRefForArtifact(actor.TenantID, req.WorkspaceID, artifactID, req.FileName)
	uploadURL, expiresAt, err := s.objectStore.UploadURL(blobRef)
	if err != nil {
		apierrors.Write(w, http.StatusInternalServerError, "artifact_upload_url_failed", "Failed to issue upload URL.", corrID, nil)
		return
	}

	s.appendEvent(r, actor.TenantID, "artifact", artifactID, "artifact.upload_url_issued", map[string]any{
		"workspace_id": req.WorkspaceID,
		"blob_ref":     blobRef,
		"media_type":   req.MediaType,
		"size_bytes":   req.SizeBytes,
	})

	writeJSON(w, http.StatusOK, map[string]any{
		"artifact_id":     artifactID,
		"workspace_id":    req.WorkspaceID,
		"blob_ref":        blobRef,
		"upload_url":      uploadURL,
		"upload_expires":  expiresAt,
		"correlation_id":  corrID,
		"required_sha256": true,
	})
}

type artifactRegisterRequest struct {
	ArtifactID  string         `json:"artifact_id,omitempty"`
	WorkspaceID string         `json:"workspace_id"`
	BlobRef     string         `json:"blob_ref"`
	SHA256      string         `json:"sha256"`
	MediaType   string         `json:"media_type"`
	SizeBytes   int64          `json:"size_bytes"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

func (s *Server) handleArtifactRegisterV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !s.validateNodeCredential(w, r, actor, corrID) {
		return
	}

	var raw map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload.", corrID, nil)
		return
	}
	for _, forbidden := range []string{"inline_bytes", "content", "data", "payload"} {
		if _, exists := raw[forbidden]; exists {
			apierrors.Write(w, http.StatusBadRequest, "artifact_inline_payload_forbidden", "Artifacts must be registered by object-store reference only.", corrID, nil)
			return
		}
	}
	payload, _ := json.Marshal(raw)
	var req artifactRegisterRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload.", corrID, nil)
		return
	}
	req.WorkspaceID = strings.TrimSpace(req.WorkspaceID)
	req.BlobRef = strings.TrimSpace(req.BlobRef)
	req.SHA256 = strings.TrimSpace(strings.ToLower(req.SHA256))
	req.MediaType = strings.TrimSpace(req.MediaType)
	if req.WorkspaceID == "" || req.BlobRef == "" || req.SHA256 == "" || req.MediaType == "" || req.SizeBytes <= 0 {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "workspace_id, blob_ref, sha256, media_type, and positive size_bytes are required.", corrID, nil)
		return
	}
	if len(req.SHA256) != 64 {
		apierrors.Write(w, http.StatusBadRequest, "invalid_sha256", "sha256 must be a 64-char lowercase hex string.", corrID, nil)
		return
	}
	if _, err := hex.DecodeString(req.SHA256); err != nil {
		apierrors.Write(w, http.StatusBadRequest, "invalid_sha256", "sha256 must be a 64-char lowercase hex string.", corrID, nil)
		return
	}

	const workspaceTenantQ = "SELECT tenant_id FROM workspaces WHERE workspace_id = $1"
	var workspaceTenantID string
	if err := s.postgres.DB.QueryRowContext(r.Context(), workspaceTenantQ, req.WorkspaceID).Scan(&workspaceTenantID); err != nil {
		if err == sql.ErrNoRows {
			apierrors.Write(w, http.StatusNotFound, "workspace_not_found", "Workspace not found.", corrID, nil)
			return
		}
		apierrors.Write(w, http.StatusInternalServerError, "workspace_lookup_failed", "Failed to validate workspace.", corrID, nil)
		return
	}
	if !s.ensureTenantAccess(actor, workspaceTenantID) {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Cross-tenant write is not allowed.", corrID, nil)
		return
	}
	if !s.objectStore.IsNamespacedBlobRef(req.BlobRef, actor.TenantID, req.WorkspaceID) {
		apierrors.Write(w, http.StatusBadRequest, "invalid_blob_ref", "blob_ref does not match tenant/workspace namespace.", corrID, nil)
		return
	}

	artifactID := strings.TrimSpace(req.ArtifactID)
	if artifactID == "" {
		artifactID = "art_" + randomID(6)
	}
	if req.Metadata == nil {
		req.Metadata = map[string]any{}
	}
	metadataJSON, _ := json.Marshal(req.Metadata)

	const insertArtifactQ = "INSERT INTO artifacts (artifact_id, tenant_id, workspace_id, blob_ref, sha256, media_type, size_bytes, metadata_json) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)"
	if _, err := s.postgres.DB.ExecContext(r.Context(), insertArtifactQ, artifactID, actor.TenantID, req.WorkspaceID, req.BlobRef, req.SHA256, req.MediaType, req.SizeBytes, metadataJSON); err != nil {
		apierrors.Write(w, http.StatusInternalServerError, "artifact_register_failed", "Failed to register artifact.", corrID, nil)
		return
	}

	s.appendEvent(r, actor.TenantID, "artifact", artifactID, "artifact.registered", map[string]any{
		"workspace_id": req.WorkspaceID,
		"blob_ref":     req.BlobRef,
		"media_type":   req.MediaType,
		"size_bytes":   req.SizeBytes,
	})

	writeJSON(w, http.StatusOK, map[string]any{
		"artifact_id":     artifactID,
		"workspace_id":    req.WorkspaceID,
		"blob_ref":        req.BlobRef,
		"sha256":          req.SHA256,
		"media_type":      req.MediaType,
		"size_bytes":      req.SizeBytes,
		"correlation_id":  corrID,
		"storage_pointer": true,
	})
}

type artifactRecord struct {
	ArtifactID  string
	TenantID    string
	WorkspaceID string
	BlobRef     string
	SHA256      string
	MediaType   string
	SizeBytes   int64
	Metadata    map[string]any
	CreatedAt   time.Time
}

func (s *Server) getArtifactByID(r *http.Request, artifactID string) (artifactRecord, error) {
	const q = "SELECT artifact_id, tenant_id, workspace_id, blob_ref, sha256, media_type, size_bytes, metadata_json, created_at FROM artifacts WHERE artifact_id = $1"
	var rec artifactRecord
	var metadataJSON []byte
	if err := s.postgres.DB.QueryRowContext(r.Context(), q, artifactID).Scan(
		&rec.ArtifactID, &rec.TenantID, &rec.WorkspaceID, &rec.BlobRef, &rec.SHA256, &rec.MediaType, &rec.SizeBytes, &metadataJSON, &rec.CreatedAt,
	); err != nil {
		return artifactRecord{}, err
	}
	if len(metadataJSON) > 0 {
		_ = json.Unmarshal(metadataJSON, &rec.Metadata)
	}
	if rec.Metadata == nil {
		rec.Metadata = map[string]any{}
	}
	return rec, nil
}

func (s *Server) handleGetArtifactV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	rec, err := s.getArtifactByID(r, r.PathValue("artifact_id"))
	if err != nil {
		if err == sql.ErrNoRows {
			apierrors.Write(w, http.StatusNotFound, "artifact_not_found", "Artifact not found.", corrID, nil)
			return
		}
		apierrors.Write(w, http.StatusInternalServerError, "artifact_lookup_failed", "Failed to fetch artifact.", corrID, nil)
		return
	}
	if !s.ensureTenantAccess(actor, rec.TenantID) {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Cross-tenant read is not allowed.", corrID, nil)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"artifact_id":     rec.ArtifactID,
		"tenant_id":       rec.TenantID,
		"workspace_id":    rec.WorkspaceID,
		"blob_ref":        rec.BlobRef,
		"sha256":          rec.SHA256,
		"media_type":      rec.MediaType,
		"size_bytes":      rec.SizeBytes,
		"metadata":        rec.Metadata,
		"created_at":      rec.CreatedAt,
		"correlation_id":  corrID,
		"storage_pointer": true,
	})
}

func (s *Server) handleArtifactDownloadURLV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	rec, err := s.getArtifactByID(r, r.PathValue("artifact_id"))
	if err != nil {
		if err == sql.ErrNoRows {
			apierrors.Write(w, http.StatusNotFound, "artifact_not_found", "Artifact not found.", corrID, nil)
			return
		}
		apierrors.Write(w, http.StatusInternalServerError, "artifact_lookup_failed", "Failed to fetch artifact.", corrID, nil)
		return
	}
	if !s.ensureTenantAccess(actor, rec.TenantID) {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Cross-tenant read is not allowed.", corrID, nil)
		return
	}
	downloadURL, expiresAt, err := s.objectStore.DownloadURL(rec.BlobRef)
	if err != nil {
		apierrors.Write(w, http.StatusInternalServerError, "artifact_download_url_failed", "Failed to issue download URL.", corrID, nil)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"artifact_id":     rec.ArtifactID,
		"download_url":    downloadURL,
		"download_expiry": expiresAt,
		"correlation_id":  corrID,
	})
}

type telemetryRefs struct {
	LogsArtifactID      string `json:"logs_artifact_id,omitempty"`
	ToolCallsArtifactID string `json:"tool_calls_artifact_id,omitempty"`
	MetricsArtifactID   string `json:"metrics_artifact_id,omitempty"`
}

type resultRecord struct {
	ResultID   string
	TenantID   string
	TaskID     string
	LeaseID    string
	NodeID     string
	Status     string
	Signature  string
	CreatedAt  time.Time
	OutputRefs []string
	Telemetry  telemetryRefs
}

func expectedResultSignature(nodeID, leaseID, resultID string) string {
	return fmt.Sprintf("signed:%s:%s:%s", nodeID, leaseID, resultID)
}

func verifyResultSignature(signature, nodeID, leaseID, resultID string) bool {
	expected := expectedResultSignature(nodeID, leaseID, resultID)
	return subtle.ConstantTimeCompare([]byte(signature), []byte(expected)) == 1
}

func (s *Server) handleGetResultV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	const q = "SELECT result_id, tenant_id, task_id, lease_id, node_id, status, signature, created_at FROM results WHERE result_id = $1"
	var rec resultRecord
	if err := s.postgres.DB.QueryRowContext(r.Context(), q, r.PathValue("result_id")).Scan(
		&rec.ResultID, &rec.TenantID, &rec.TaskID, &rec.LeaseID, &rec.NodeID, &rec.Status, &rec.Signature, &rec.CreatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			apierrors.Write(w, http.StatusNotFound, "result_not_found", "Result not found.", corrID, nil)
			return
		}
		apierrors.Write(w, http.StatusInternalServerError, "result_lookup_failed", "Failed to fetch result.", corrID, nil)
		return
	}
	if !s.ensureTenantAccess(actor, rec.TenantID) {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Cross-tenant read is not allowed.", corrID, nil)
		return
	}
	rec.OutputRefs = s.lookupResultOutputRefs(r, rec.ResultID)
	rec.Telemetry = s.lookupTelemetryRefs(r, rec.ResultID)
	writeJSON(w, http.StatusOK, map[string]any{
		"result_id":      rec.ResultID,
		"tenant_id":      rec.TenantID,
		"task_id":        rec.TaskID,
		"lease_id":       rec.LeaseID,
		"node_id":        rec.NodeID,
		"status":         rec.Status,
		"signature":      rec.Signature,
		"output_refs":    rec.OutputRefs,
		"telemetry_refs": rec.Telemetry,
		"created_at":     rec.CreatedAt,
		"correlation_id": corrID,
	})
}

func (s *Server) handleTaskResultsV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	const q = "SELECT result_id, tenant_id, task_id, lease_id, node_id, status, signature, created_at FROM results WHERE task_id = $1 AND tenant_id = $2 ORDER BY created_at DESC"
	rows, err := s.postgres.DB.QueryContext(r.Context(), q, r.PathValue("task_id"), actor.TenantID)
	if err != nil {
		apierrors.Write(w, http.StatusInternalServerError, "result_lookup_failed", "Failed to fetch task results.", corrID, nil)
		return
	}
	defer rows.Close()

	results := make([]map[string]any, 0)
	for rows.Next() {
		var rec resultRecord
		if err := rows.Scan(&rec.ResultID, &rec.TenantID, &rec.TaskID, &rec.LeaseID, &rec.NodeID, &rec.Status, &rec.Signature, &rec.CreatedAt); err != nil {
			apierrors.Write(w, http.StatusInternalServerError, "result_lookup_failed", "Failed to scan task results.", corrID, nil)
			return
		}
		results = append(results, map[string]any{
			"result_id":   rec.ResultID,
			"task_id":     rec.TaskID,
			"lease_id":    rec.LeaseID,
			"node_id":     rec.NodeID,
			"status":      rec.Status,
			"signature":   rec.Signature,
			"created_at":  rec.CreatedAt,
			"output_refs": s.lookupResultOutputRefs(r, rec.ResultID),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"results": results, "correlation_id": corrID})
}

func (s *Server) lookupResultOutputRefs(r *http.Request, resultID string) []string {
	const q = "SELECT artifact_id FROM result_output_refs WHERE result_id = $1"
	rows, err := s.postgres.DB.QueryContext(r.Context(), q, resultID)
	if err != nil {
		return []string{}
	}
	defer rows.Close()
	out := make([]string, 0)
	for rows.Next() {
		var artifactID string
		if err := rows.Scan(&artifactID); err == nil {
			out = append(out, artifactID)
		}
	}
	return out
}

func (s *Server) lookupTelemetryRefs(r *http.Request, resultID string) telemetryRefs {
	const q = "SELECT logs_artifact_id, tool_calls_artifact_id, metrics_artifact_id FROM run_telemetry_refs WHERE result_id = $1"
	var refs telemetryRefs
	_ = s.postgres.DB.QueryRowContext(r.Context(), q, resultID).Scan(&refs.LogsArtifactID, &refs.ToolCallsArtifactID, &refs.MetricsArtifactID)
	return refs
}
