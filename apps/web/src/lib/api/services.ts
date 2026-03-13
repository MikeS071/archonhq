import type {
  ApprovalItem,
  Artifact,
  FleetNode,
  LedgerEntry,
  Metric,
  ProviderPolicy,
  ReliabilityPoint,
  Task,
  TaskEvent,
  Verification
} from "$lib/types";
import { apiGet } from "$lib/api/client";

interface TaskFeedResponse {
  tasks: Array<{
    task_id: string;
    workspace_id: string;
    task_family: string;
    title: string;
    status: string;
    created_at: string;
  }>;
}

interface TaskResponse {
  task_id: string;
  workspace_id: string;
  task_family: string;
  title: string;
  status: string;
  created_at: string;
}

interface ApprovalQueueResponse {
  approvals: Array<{
    approval_id: string;
    task_id: string;
    status: string;
    created_at: string;
  }>;
}

interface TaskResultsResponse {
  results: Array<{
    result_id: string;
    task_id: string;
    lease_id: string;
    node_id: string;
    status: string;
    created_at: string;
    output_refs: string[];
  }>;
}

interface ArtifactResponse {
  artifact_id: string;
  media_type: string;
  size_bytes: number;
  metadata: Record<string, unknown>;
}

interface NodeResponse {
  node_id: string;
  operator_id: string;
  runtime_type: string;
  runtime_version: string;
  status: string;
  last_heartbeat_at: string;
}

interface NodeListResponse {
  nodes: Array<{
    node_id: string;
    operator_id: string;
    runtime_type: string;
    runtime_version: string;
    status: string;
    last_heartbeat_at: string;
    active_leases: number;
  }>;
}

interface NodeLeasesResponse {
  leases: Array<{ lease_id: string }>;
}

interface ReliabilityResponse {
  rf_value: number;
  components: Record<string, number>;
}

interface LedgerEntriesResponse {
  entries: Array<{
    entry_id: string;
    result_id: string;
    event_type: string;
    net_amount: number;
    created_at: string;
    status: string;
  }>;
}

interface PricingRateCardsResponse {
  rate_cards: Array<Record<string, unknown>>;
}

interface PoliciesResponse {
  policies: Array<{
    provider: string;
    model: string;
    max_usd_per_task: number;
    retries: number;
    requires_approval: boolean;
  }>;
}

export interface DashboardPayload {
  metrics: Metric[];
  tasks: Task[];
  approvals: ApprovalItem[];
}

export interface TaskDetailPayload {
  task: Task;
  events: TaskEvent[];
  artifacts: Artifact[];
  verifications: Verification[];
}

export interface PricingPayload {
  rateCards: Array<Record<string, unknown>>;
}

function formatRelative(value: string): string {
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) {
    return "unknown";
  }

  const diffMs = Date.now() - parsed.getTime();
  const diffMins = Math.max(0, Math.floor(diffMs / 60000));

  if (diffMins < 1) return "just now";
  if (diffMins < 60) return `${diffMins} min ago`;

  const diffHours = Math.floor(diffMins / 60);
  if (diffHours < 24) return `${diffHours} hr ago`;

  const diffDays = Math.floor(diffHours / 24);
  return `${diffDays} d ago`;
}

function mapStatus(status: string): Task["status"] {
  if (
    status === "queued" ||
    status === "leased" ||
    status === "running" ||
    status === "awaiting_approval" ||
    status === "completed" ||
    status === "failed"
  ) {
    return status;
  }
  return "queued";
}

function priorityForStatus(status: Task["status"]): Task["priority"] {
  if (status === "awaiting_approval") return "urgent";
  if (status === "running" || status === "failed") return "high";
  if (status === "queued" || status === "leased") return "medium";
  return "low";
}

function mapTask(item: TaskResponse | TaskFeedResponse["tasks"][number]): Task {
  const status = mapStatus(item.status);
  return {
    id: item.task_id,
    title: item.title,
    family: item.task_family,
    status,
    priority: priorityForStatus(status),
    workspaceId: item.workspace_id,
    assignee: "unassigned",
    updatedAt: formatRelative(item.created_at),
    joules: 0
  };
}

function parseNodeStatus(status: string): FleetNode["status"] {
  if (status === "offline") return "offline";
  if (status === "degraded") return "degraded";
  return "healthy";
}

export function defaultAccountId(): string {
  const configured = (import.meta.env.PUBLIC_ARCHON_DEFAULT_ACCOUNT_ID as string | undefined)?.trim();
  return configured || "acct_ops";
}

export function defaultOperatorId(): string {
  const configured = (import.meta.env.PUBLIC_ARCHON_DEFAULT_OPERATOR_ID as string | undefined)?.trim();
  return configured || "op_1";
}

export async function fetchTaskFeed(limit = 50): Promise<Task[]> {
  const response = await apiGet<TaskFeedResponse>(`/v1/tasks/feed?limit=${limit}`);
  return response.tasks.map(mapTask);
}

export async function fetchTask(taskId: string): Promise<Task> {
  const response = await apiGet<TaskResponse>(`/v1/tasks/${taskId}`);
  return mapTask(response);
}

export async function fetchApprovals(): Promise<ApprovalItem[]> {
  const response = await apiGet<ApprovalQueueResponse>("/v1/approvals/queue");

  return response.approvals.map((approval) => ({
    id: approval.approval_id,
    taskId: approval.task_id,
    workspaceId: "unknown",
    reason: `Status: ${approval.status}`,
    createdAt: formatRelative(approval.created_at),
    requester: "system"
  }));
}

export async function fetchDashboard(): Promise<DashboardPayload> {
  const [tasks, approvals] = await Promise.all([fetchTaskFeed(40), fetchApprovals()]);

  const activeCount = tasks.filter((task) => task.status !== "completed" && task.status !== "failed").length;
  const completedCount = tasks.filter((task) => task.status === "completed").length;
  const failedCount = tasks.filter((task) => task.status === "failed").length;

  const metrics: Metric[] = [
    {
      label: "Active Tasks",
      value: String(activeCount),
      change: `${tasks.length} total in feed`,
      direction: "flat"
    },
    {
      label: "Pending Approvals",
      value: String(approvals.length),
      change: "live queue",
      direction: approvals.length > 0 ? "up" : "flat"
    },
    {
      label: "Completed",
      value: String(completedCount),
      change: "from current feed",
      direction: completedCount > 0 ? "up" : "flat"
    },
    {
      label: "Failed",
      value: String(failedCount),
      change: "needs triage",
      direction: failedCount > 0 ? "down" : "flat"
    }
  ];

  return {
    metrics,
    tasks,
    approvals
  };
}

export async function fetchTaskDetail(taskId: string): Promise<TaskDetailPayload> {
  const [task, resultsResponse] = await Promise.all([
    fetchTask(taskId),
    apiGet<TaskResultsResponse>(`/v1/tasks/${taskId}/results`)
  ]);

  const events: TaskEvent[] = [
    {
      id: `evt-created-${task.id}`,
      at: task.updatedAt,
      label: "Task observed in feed",
      actor: "api"
    },
    ...resultsResponse.results.map((result) => ({
      id: `evt-result-${result.result_id}`,
      at: formatRelative(result.created_at),
      label: `Result ${result.status}`,
      actor: result.node_id
    }))
  ];

  const artifactIds = Array.from(
    new Set(
      resultsResponse.results
        .flatMap((result) => result.output_refs || [])
        .filter((artifactId) => artifactId && artifactId.length > 0)
    )
  );

  const artifactResponses = await Promise.all(
    artifactIds.map((artifactId) => apiGet<ArtifactResponse>(`/v1/artifacts/${artifactId}`))
  );

  const artifacts: Artifact[] = artifactResponses.map((artifact) => ({
    id: artifact.artifact_id,
    name: String(artifact.metadata?.filename ?? artifact.artifact_id),
    size: `${Math.max(1, Math.round(artifact.size_bytes / 1024))} KB`,
    kind: artifact.media_type
  }));

  const verifications: Verification[] = resultsResponse.results.map((result) => ({
    id: `ver-${result.result_id}`,
    verdict: result.status === "accepted" || result.status === "completed" ? "pass" : "warn",
    summary: `Result ${result.result_id} is ${result.status}`,
    createdAt: formatRelative(result.created_at)
  }));

  return {
    task,
    events,
    artifacts,
    verifications
  };
}

export async function fetchFleetNodes(): Promise<FleetNode[]> {
  const listed = await apiGet<NodeListResponse>("/v1/nodes?limit=500");

  const resolved = await Promise.all(
    listed.nodes.map(async (node) => {
      const reliability = await apiGet<ReliabilityResponse>(`/v1/reliability/subjects/node/${node.node_id}`).catch(() => ({
        rf_value: 0,
        components: {}
      }));

      return {
        id: node.node_id,
        operator: node.operator_id,
        runtime: `${node.runtime_type} ${node.runtime_version}`,
        status: parseNodeStatus(node.status),
        lastHeartbeat: formatRelative(node.last_heartbeat_at),
        activeLeases: Number(node.active_leases) || 0,
        reliability: Math.round((reliability.rf_value || 0) * 100)
      } satisfies FleetNode;
    })
  );

  return resolved;
}

export async function fetchNode(nodeId: string): Promise<FleetNode> {
  const [node, leases, reliability] = await Promise.all([
    apiGet<NodeResponse>(`/v1/nodes/${nodeId}`),
    apiGet<NodeLeasesResponse>(`/v1/nodes/${nodeId}/leases`),
    apiGet<ReliabilityResponse>(`/v1/reliability/subjects/node/${nodeId}`).catch(() => ({
      rf_value: 0,
      components: {}
    }))
  ]);

  return {
    id: node.node_id,
    operator: node.operator_id,
    runtime: `${node.runtime_type} ${node.runtime_version}`,
    status: parseNodeStatus(node.status),
    lastHeartbeat: formatRelative(node.last_heartbeat_at),
    activeLeases: leases.leases.length,
    reliability: Math.round((reliability.rf_value || 0) * 100)
  };
}

export async function fetchLedgerEntries(accountId = defaultAccountId()): Promise<LedgerEntry[]> {
  const response = await apiGet<LedgerEntriesResponse>(`/v1/ledger/accounts/${accountId}/entries`);

  return response.entries.map((entry) => {
    const signed = Number(entry.net_amount ?? 0);
    return {
      id: entry.entry_id,
      accountId,
      type: signed >= 0 ? "credit" : "debit",
      amount: Math.abs(signed),
      description: `${entry.event_type} · ${entry.result_id || "n/a"}`,
      postedAt: formatRelative(entry.created_at)
    };
  });
}

export async function fetchReliabilitySeries(operatorId = defaultOperatorId()): Promise<ReliabilityPoint[]> {
  const response = await apiGet<ReliabilityResponse>(`/v1/operators/${operatorId}/reliability`);

  const pairs = Object.entries(response.components || {}).filter(([, value]) => Number.isFinite(value));

  if (pairs.length === 0) {
    return [{ label: "RF", value: Math.round(response.rf_value * 100) }];
  }

  return pairs.map(([label, value]) => ({
    label,
    value: Math.round(Number(value) * 100)
  }));
}

export async function fetchPricing(): Promise<PricingPayload> {
  const response = await apiGet<PricingRateCardsResponse>("/v1/pricing/rate-cards");
  return { rateCards: response.rate_cards || [] };
}

export async function fetchProviderPolicies(): Promise<ProviderPolicy[]> {
  const response = await apiGet<PoliciesResponse>("/v1/policies");

  return response.policies.map((policy) => ({
    provider: policy.provider,
    model: policy.model,
    maxUsdPerTask: policy.max_usd_per_task,
    retries: policy.retries,
    requiresApproval: policy.requires_approval
  }));
}
