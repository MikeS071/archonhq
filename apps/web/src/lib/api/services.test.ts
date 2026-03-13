import { afterEach, describe, expect, it, vi } from "vitest";
import {
  defaultAccountId,
  defaultOperatorId,
  fetchApprovals,
  fetchDashboard,
  fetchFleetNodes,
  fetchLedgerEntries,
  fetchNode,
  fetchPricing,
  fetchProviderPolicies,
  fetchReliabilitySeries,
  fetchTask,
  fetchTaskDetail,
  fetchTaskFeed
} from "$lib/api/services";

function jsonResponse(body: unknown, status = 200) {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "content-type": "application/json" }
  });
}

afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe("api services", () => {
  it("maps task feed payload to UI task model", async () => {
    const fetchMock = vi.fn(async () =>
      jsonResponse({
        tasks: [
          {
            task_id: "task_1",
            workspace_id: "ws_1",
            task_family: "research.extract",
            title: "Task Title",
            status: "running",
            created_at: new Date(Date.now() - 60_000).toISOString()
          }
        ]
      })
    );
    vi.stubGlobal("fetch", fetchMock);

    const result = await fetchTaskFeed(10);

    expect(fetchMock).toHaveBeenCalledTimes(1);
    expect(result).toHaveLength(1);
    expect(result[0]).toMatchObject({
      id: "task_1",
      workspaceId: "ws_1",
      family: "research.extract",
      status: "running",
      priority: "high"
    });
  });

  it("maps unknown task status to queued", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () =>
        jsonResponse({
          tasks: [
            {
              task_id: "task_1",
              workspace_id: "ws_1",
              task_family: "research.extract",
              title: "Task Title",
              status: "mystery_status",
              created_at: new Date().toISOString()
            }
          ]
        })
      )
    );

    const result = await fetchTaskFeed(10);
    expect(result[0].status).toBe("queued");
    expect(result[0].priority).toBe("medium");
  });

  it("fetchTask maps direct task response", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () =>
        jsonResponse({
          task_id: "task_77",
          workspace_id: "ws_a",
          task_family: "doc.section.write",
          title: "Write doc",
          status: "awaiting_approval",
          created_at: new Date().toISOString()
        })
      )
    );

    const task = await fetchTask("task_77");
    expect(task.id).toBe("task_77");
    expect(task.priority).toBe("urgent");
  });

  it("fetchApprovals maps queue response", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () =>
        jsonResponse({
          approvals: [
            {
              approval_id: "apr_1",
              task_id: "task_1",
              status: "pending",
              created_at: new Date().toISOString()
            }
          ]
        })
      )
    );

    const approvals = await fetchApprovals();
    expect(approvals[0]).toMatchObject({
      id: "apr_1",
      taskId: "task_1"
    });
  });

  it("builds dashboard metrics from tasks + approvals endpoints", async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
      const url = input.toString();

      if (url.includes("/v1/tasks/feed")) {
        return jsonResponse({
          tasks: [
            {
              task_id: "task_a",
              workspace_id: "ws_1",
              task_family: "research.extract",
              title: "A",
              status: "running",
              created_at: new Date().toISOString()
            },
            {
              task_id: "task_b",
              workspace_id: "ws_1",
              task_family: "doc.section.write",
              title: "B",
              status: "failed",
              created_at: new Date().toISOString()
            }
          ]
        });
      }

      return jsonResponse({
        approvals: [
          {
            approval_id: "apr_1",
            task_id: "task_a",
            status: "pending",
            created_at: new Date().toISOString()
          }
        ]
      });
    });

    vi.stubGlobal("fetch", fetchMock);

    const dashboard = await fetchDashboard();

    expect(dashboard.tasks).toHaveLength(2);
    expect(dashboard.approvals).toHaveLength(1);
    expect(dashboard.metrics[0].label).toBe("Active Tasks");
    expect(dashboard.metrics[1].label).toBe("Pending Approvals");
  });

  it("fetchTaskDetail builds events, artifacts and verifications", async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
      const url = input.toString();

      if (url.includes("/v1/tasks/task_2/results")) {
        return jsonResponse({
          results: [
            {
              result_id: "res_1",
              task_id: "task_2",
              lease_id: "lease_1",
              node_id: "node_1",
              status: "completed",
              created_at: new Date().toISOString(),
              output_refs: ["art_1"]
            }
          ]
        });
      }

      if (url.includes("/v1/tasks/task_2")) {
        return jsonResponse({
          task_id: "task_2",
          workspace_id: "ws_1",
          task_family: "verify.result",
          title: "Verify",
          status: "completed",
          created_at: new Date().toISOString()
        });
      }

      if (url.includes("/v1/artifacts/art_1")) {
        return jsonResponse({
          artifact_id: "art_1",
          media_type: "application/json",
          size_bytes: 10240,
          metadata: { filename: "out.json" }
        });
      }

      return jsonResponse({}, 404);
    });

    vi.stubGlobal("fetch", fetchMock);

    const detail = await fetchTaskDetail("task_2");
    expect(detail.events.length).toBeGreaterThan(0);
    expect(detail.artifacts[0].name).toBe("out.json");
    expect(detail.verifications[0].verdict).toBe("pass");
  });

  it("fetchTaskDetail supports empty result sets", async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
      const url = input.toString();

      if (url.includes("/v1/tasks/task_2/results")) {
        return jsonResponse({ results: [] });
      }

      return jsonResponse({
        task_id: "task_2",
        workspace_id: "ws_1",
        task_family: "verify.result",
        title: "Verify",
        status: "completed",
        created_at: new Date().toISOString()
      });
    });

    vi.stubGlobal("fetch", fetchMock);

    const detail = await fetchTaskDetail("task_2");
    expect(detail.artifacts).toHaveLength(0);
    expect(detail.verifications).toHaveLength(0);
  });

  it("fetchFleetNodes uses /v1/nodes and reliability snapshots", async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
      const url = input.toString();

      if (url.includes("/v1/nodes?limit=500")) {
        return jsonResponse({
          nodes: [
            {
              node_id: "node_1",
              operator_id: "op_1",
              runtime_type: "hermes",
              runtime_version: "1.9",
              status: "healthy",
              last_heartbeat_at: new Date().toISOString(),
              active_leases: 3
            }
          ]
        });
      }

      if (url.includes("/v1/reliability/subjects/node/node_1")) {
        return jsonResponse({ rf_value: 0.91, components: {} });
      }

      return jsonResponse({}, 404);
    });

    vi.stubGlobal("fetch", fetchMock);

    const nodes = await fetchFleetNodes();
    expect(nodes).toHaveLength(1);
    expect(nodes[0].activeLeases).toBe(3);
    expect(nodes[0].reliability).toBe(91);
  });

  it("fetchFleetNodes falls back to reliability=0 on reliability errors", async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
      const url = input.toString();

      if (url.includes("/v1/nodes?limit=500")) {
        return jsonResponse({
          nodes: [
            {
              node_id: "node_1",
              operator_id: "op_1",
              runtime_type: "hermes",
              runtime_version: "1.9",
              status: "degraded",
              last_heartbeat_at: new Date().toISOString(),
              active_leases: 1
            }
          ]
        });
      }

      return jsonResponse(
        {
          error: {
            code: "reliability_not_found",
            message: "missing"
          }
        },
        404
      );
    });

    vi.stubGlobal("fetch", fetchMock);

    const nodes = await fetchFleetNodes();
    expect(nodes[0].reliability).toBe(0);
    expect(nodes[0].status).toBe("degraded");
  });

  it("fetchNode returns detailed node payload", async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
      const url = input.toString();

      if (url.includes("/v1/nodes/node_1/leases")) {
        return jsonResponse({ leases: [{ lease_id: "l1" }, { lease_id: "l2" }] });
      }
      if (url.includes("/v1/nodes/node_1")) {
        return jsonResponse({
          node_id: "node_1",
          operator_id: "op_1",
          runtime_type: "hermes",
          runtime_version: "1.9",
          status: "offline",
          last_heartbeat_at: new Date().toISOString()
        });
      }
      if (url.includes("/v1/reliability/subjects/node/node_1")) {
        return jsonResponse({ rf_value: 0.88, components: {} });
      }
      return jsonResponse({}, 404);
    });

    vi.stubGlobal("fetch", fetchMock);

    const node = await fetchNode("node_1");
    expect(node.status).toBe("offline");
    expect(node.activeLeases).toBe(2);
    expect(node.reliability).toBe(88);
  });

  it("fetchLedgerEntries maps debit and credit values", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () =>
        jsonResponse({
          entries: [
            {
              entry_id: "e1",
              result_id: "r1",
              event_type: "settlement",
              net_amount: 10,
              created_at: new Date().toISOString(),
              status: "posted"
            },
            {
              entry_id: "e2",
              result_id: "r2",
              event_type: "reserve",
              net_amount: -5,
              created_at: new Date().toISOString(),
              status: "posted"
            }
          ]
        })
      )
    );

    const entries = await fetchLedgerEntries("acct_1");
    expect(entries[0].type).toBe("credit");
    expect(entries[1].type).toBe("debit");
  });

  it("fetchReliabilitySeries returns components map when present", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => jsonResponse({ rf_value: 0.8, components: { last_100: 0.75, last_30d: 0.9 } }))
    );

    const points = await fetchReliabilitySeries("op_1");
    expect(points).toHaveLength(2);
    expect(points[0].label).toBe("last_100");
  });

  it("fetchReliabilitySeries falls back to RF when no components", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => jsonResponse({ rf_value: 0.84, components: {} })));

    const points = await fetchReliabilitySeries("op_1");
    expect(points).toHaveLength(1);
    expect(points[0]).toMatchObject({ label: "RF", value: 84 });
  });

  it("fetchPricing and fetchProviderPolicies map payloads", async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
      const url = input.toString();

      if (url.includes("/v1/pricing/rate-cards")) {
        return jsonResponse({ rate_cards: [{ family: "research.extract", base_rate: 1.2 }] });
      }

      return jsonResponse({
        policies: [
          {
            provider: "OpenAI",
            model: "gpt-5",
            max_usd_per_task: 2.2,
            retries: 2,
            requires_approval: true
          }
        ]
      });
    });

    vi.stubGlobal("fetch", fetchMock);

    const pricing = await fetchPricing();
    const policies = await fetchProviderPolicies();

    expect(pricing.rateCards).toHaveLength(1);
    expect(policies[0]).toMatchObject({ provider: "OpenAI", requiresApproval: true });
  });

  it("default identifiers use fallbacks", () => {
    expect(defaultAccountId()).toBe("acct_ops");
    expect(defaultOperatorId()).toBe("op_1");
  });
});
