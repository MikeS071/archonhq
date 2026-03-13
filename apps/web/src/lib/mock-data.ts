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

export const dashboardMetrics: Metric[] = [
  { label: "Active Tasks", value: "38", change: "+12% this week", direction: "up" },
  { label: "Pending Approvals", value: "7", change: "-2 from yesterday", direction: "down" },
  { label: "Healthy Nodes", value: "92%", change: "+1.3%", direction: "up" },
  { label: "Settlement Throughput", value: "$14.2k", change: "steady", direction: "flat" }
];

export const tasks: Task[] = [
  {
    id: "task_001",
    title: "Extract partner compliance obligations",
    family: "research.extract",
    status: "running",
    priority: "high",
    workspaceId: "ws_ops",
    assignee: "node_az_21",
    updatedAt: "2 min ago",
    joules: 482
  },
  {
    id: "task_002",
    title: "Draft SOC2 control narrative",
    family: "doc.section.write",
    status: "awaiting_approval",
    priority: "urgent",
    workspaceId: "ws_compliance",
    assignee: "node_az_04",
    updatedAt: "7 min ago",
    joules: 301
  },
  {
    id: "task_003",
    title: "Patch retry strategy in lease worker",
    family: "code.patch",
    status: "queued",
    priority: "medium",
    workspaceId: "ws_platform",
    assignee: "unassigned",
    updatedAt: "11 min ago",
    joules: 0
  },
  {
    id: "task_004",
    title: "Verify settlement variance report",
    family: "verify.result",
    status: "completed",
    priority: "low",
    workspaceId: "ws_finance",
    assignee: "node_eu_09",
    updatedAt: "21 min ago",
    joules: 188
  },
  {
    id: "task_005",
    title: "Merge duplicate evidence findings",
    family: "reduce.merge",
    status: "leased",
    priority: "medium",
    workspaceId: "ws_compliance",
    assignee: "node_ap_08",
    updatedAt: "33 min ago",
    joules: 76
  },
  {
    id: "task_006",
    title: "Run bounded autosearch optimization",
    family: "autosearch.self_improve",
    status: "failed",
    priority: "high",
    workspaceId: "ws_platform",
    assignee: "node_eu_03",
    updatedAt: "58 min ago",
    joules: 514
  }
];

export const taskEvents: Record<string, TaskEvent[]> = {
  task_001: [
    { id: "ev_1", at: "11:04", label: "Task created", actor: "user:ops_admin" },
    { id: "ev_2", at: "11:06", label: "Lease claimed", actor: "node_az_21" },
    { id: "ev_3", at: "11:10", label: "Artifact uploaded", actor: "node_az_21" },
    { id: "ev_4", at: "11:12", label: "Verification queued", actor: "system" }
  ],
  task_002: [
    { id: "ev_5", at: "10:48", label: "Task completed", actor: "node_az_04" },
    { id: "ev_6", at: "10:49", label: "Approval requested", actor: "system" }
  ]
};

export const taskArtifacts: Record<string, Artifact[]> = {
  task_001: [
    { id: "art_1", name: "obligations.csv", size: "218 KB", kind: "dataset" },
    { id: "art_2", name: "source-links.md", size: "19 KB", kind: "report" }
  ],
  task_002: [{ id: "art_3", name: "soc2-draft.md", size: "43 KB", kind: "document" }]
};

export const taskVerifications: Record<string, Verification[]> = {
  task_001: [
    { id: "ver_1", verdict: "pass", summary: "Source citations complete", createdAt: "11:12" },
    { id: "ver_2", verdict: "warn", summary: "2 low-confidence claims", createdAt: "11:13" }
  ],
  task_002: [
    { id: "ver_3", verdict: "pass", summary: "Structure and policy checks passed", createdAt: "10:47" }
  ]
};

export const approvals: ApprovalItem[] = [
  {
    id: "apr_301",
    taskId: "task_002",
    workspaceId: "ws_compliance",
    reason: "Publish SOC2 narrative to customer portal",
    createdAt: "7 min ago",
    requester: "user:compliance_lead"
  },
  {
    id: "apr_302",
    taskId: "task_001",
    workspaceId: "ws_ops",
    reason: "Escalated budget threshold exceeded",
    createdAt: "14 min ago",
    requester: "system:policy"
  },
  {
    id: "apr_303",
    taskId: "task_005",
    workspaceId: "ws_compliance",
    reason: "External evidence merge approval",
    createdAt: "27 min ago",
    requester: "user:auditor_2"
  }
];

export const fleetNodes: FleetNode[] = [
  {
    id: "node_az_21",
    operator: "ops-team",
    runtime: "hermes 1.9",
    status: "healthy",
    lastHeartbeat: "20s ago",
    activeLeases: 4,
    reliability: 97
  },
  {
    id: "node_az_04",
    operator: "compliance-team",
    runtime: "hermes 1.9",
    status: "healthy",
    lastHeartbeat: "32s ago",
    activeLeases: 2,
    reliability: 95
  },
  {
    id: "node_eu_03",
    operator: "platform-team",
    runtime: "hermes 1.8",
    status: "degraded",
    lastHeartbeat: "2m ago",
    activeLeases: 1,
    reliability: 88
  },
  {
    id: "node_ap_08",
    operator: "finops",
    runtime: "hermes 1.9",
    status: "offline",
    lastHeartbeat: "17m ago",
    activeLeases: 0,
    reliability: 72
  }
];

export const ledgerEntries: LedgerEntry[] = [
  {
    id: "led_801",
    accountId: "acct_ops",
    type: "credit",
    amount: 221.4,
    description: "Task settlement payout",
    postedAt: "11:04"
  },
  {
    id: "led_802",
    accountId: "acct_ops",
    type: "debit",
    amount: 48.12,
    description: "Reserve hold",
    postedAt: "10:59"
  },
  {
    id: "led_803",
    accountId: "acct_finance",
    type: "credit",
    amount: 103.33,
    description: "Quality bonus",
    postedAt: "10:41"
  },
  {
    id: "led_804",
    accountId: "acct_platform",
    type: "debit",
    amount: 77.8,
    description: "Provider usage cost",
    postedAt: "10:13"
  }
];

export const reliabilitySeries: ReliabilityPoint[] = [
  { label: "Mon", value: 91 },
  { label: "Tue", value: 93 },
  { label: "Wed", value: 90 },
  { label: "Thu", value: 95 },
  { label: "Fri", value: 96 },
  { label: "Sat", value: 94 },
  { label: "Sun", value: 97 }
];

export const providerPolicies: ProviderPolicy[] = [
  {
    provider: "OpenAI",
    model: "gpt-5",
    maxUsdPerTask: 3.5,
    retries: 2,
    requiresApproval: true
  },
  {
    provider: "Anthropic",
    model: "claude-4-sonnet",
    maxUsdPerTask: 2.2,
    retries: 1,
    requiresApproval: false
  },
  {
    provider: "Local Inference",
    model: "hermes-local-v1",
    maxUsdPerTask: 0.4,
    retries: 3,
    requiresApproval: false
  }
];

export function getTask(taskId: string): Task | undefined {
  return tasks.find((task) => task.id === taskId);
}

export function getNode(nodeId: string): FleetNode | undefined {
  return fleetNodes.find((node) => node.id === nodeId);
}
