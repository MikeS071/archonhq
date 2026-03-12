package domain

import "time"

const (
	TaskFamilyResearchExtract   = "research.extract"
	TaskFamilyDocSectionWrite   = "doc.section.write"
	TaskFamilyCodePatch         = "code.patch"
	TaskFamilyVerifyResult      = "verify.result"
	TaskFamilyReduceMerge       = "reduce.merge"
	TaskFamilyAutosearchImprove = "autosearch.self_improve"
)

var SupportedTaskFamilies = map[string]struct{}{
	TaskFamilyResearchExtract:   {},
	TaskFamilyDocSectionWrite:   {},
	TaskFamilyCodePatch:         {},
	TaskFamilyVerifyResult:      {},
	TaskFamilyReduceMerge:       {},
	TaskFamilyAutosearchImprove: {},
}

type Task struct {
	TaskID      string    `json:"task_id"`
	TenantID    string    `json:"tenant_id"`
	WorkspaceID string    `json:"workspace_id"`
	TaskFamily  string    `json:"task_family"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type Result struct {
	ResultID  string    `json:"result_id"`
	TenantID  string    `json:"tenant_id"`
	TaskID    string    `json:"task_id"`
	LeaseID   string    `json:"lease_id"`
	NodeID    string    `json:"node_id"`
	Status    string    `json:"status"`
	Signature string    `json:"signature,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
