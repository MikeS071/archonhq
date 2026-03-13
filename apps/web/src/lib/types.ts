export type Role =
  | "platform_admin"
  | "tenant_admin"
  | "operator"
  | "approver"
  | "auditor"
  | "finance_viewer"
  | "developer";

export type TaskStatus =
  | "queued"
  | "leased"
  | "running"
  | "awaiting_approval"
  | "completed"
  | "failed";

export type Priority = "low" | "medium" | "high" | "urgent";

export interface Metric {
  label: string;
  value: string;
  change: string;
  direction: "up" | "down" | "flat";
}

export interface Task {
  id: string;
  title: string;
  family: string;
  status: TaskStatus;
  priority: Priority;
  workspaceId: string;
  assignee: string;
  updatedAt: string;
  joules: number;
}

export interface TaskEvent {
  id: string;
  at: string;
  label: string;
  actor: string;
}

export interface Artifact {
  id: string;
  name: string;
  size: string;
  kind: string;
}

export interface Verification {
  id: string;
  verdict: "pass" | "warn" | "fail";
  summary: string;
  createdAt: string;
}

export interface ApprovalItem {
  id: string;
  taskId: string;
  workspaceId: string;
  reason: string;
  createdAt: string;
  requester: string;
}

export interface FleetNode {
  id: string;
  operator: string;
  runtime: string;
  status: "healthy" | "degraded" | "offline";
  lastHeartbeat: string;
  activeLeases: number;
  reliability: number;
}

export interface LedgerEntry {
  id: string;
  accountId: string;
  type: "debit" | "credit";
  amount: number;
  description: string;
  postedAt: string;
}

export interface ReliabilityPoint {
  label: string;
  value: number;
}

export interface ProviderPolicy {
  provider: string;
  model: string;
  maxUsdPerTask: number;
  retries: number;
  requiresApproval: boolean;
}
