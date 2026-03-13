import { expect, test, type Page, type Route } from "@playwright/test";

async function json(route: Route, status: number, body: unknown) {
  await route.fulfill({
    status,
    contentType: "application/json",
    body: JSON.stringify(body)
  });
}

async function mockArchonApi(page: Page) {
  await page.route("**/v1/**", async (route) => {
    const req = route.request();
    const method = req.method();
    const url = new URL(req.url());
    const path = url.pathname;

    if (method === "GET" && path === "/v1/tasks/feed") {
      return json(route, 200, {
        tasks: [
          {
            task_id: "task_001",
            workspace_id: "ws_ops",
            task_family: "research.extract",
            title: "Extract partner compliance obligations",
            status: "running",
            created_at: new Date(Date.now() - 3 * 60 * 1000).toISOString()
          },
          {
            task_id: "task_002",
            workspace_id: "ws_compliance",
            task_family: "doc.section.write",
            title: "Draft SOC2 narrative",
            status: "awaiting_approval",
            created_at: new Date(Date.now() - 10 * 60 * 1000).toISOString()
          }
        ]
      });
    }

    if (method === "GET" && path === "/v1/approvals/queue") {
      return json(route, 200, {
        approvals: [
          {
            approval_id: "apr_301",
            task_id: "task_002",
            status: "pending",
            created_at: new Date(Date.now() - 6 * 60 * 1000).toISOString()
          }
        ]
      });
    }

    if (method === "GET" && path === "/v1/tasks/task_002") {
      return json(route, 200, {
        task_id: "task_002",
        workspace_id: "ws_compliance",
        task_family: "doc.section.write",
        title: "Draft SOC2 narrative",
        status: "awaiting_approval",
        created_at: new Date(Date.now() - 15 * 60 * 1000).toISOString()
      });
    }

    if (method === "GET" && path === "/v1/tasks/task_002/results") {
      return json(route, 200, {
        results: [
          {
            result_id: "res_01",
            task_id: "task_002",
            lease_id: "lease_01",
            node_id: "node_az_21",
            status: "completed",
            created_at: new Date(Date.now() - 8 * 60 * 1000).toISOString(),
            output_refs: ["art_1"]
          }
        ]
      });
    }

    if (method === "GET" && path === "/v1/artifacts/art_1") {
      return json(route, 200, {
        artifact_id: "art_1",
        media_type: "application/json",
        size_bytes: 24576,
        metadata: {
          filename: "soc2-draft.json"
        }
      });
    }

    if (method === "GET" && path === "/v1/nodes/node_az_21") {
      return json(route, 200, {
        node_id: "node_az_21",
        operator_id: "ops-team",
        runtime_type: "hermes",
        runtime_version: "1.9",
        status: "healthy",
        last_heartbeat_at: new Date(Date.now() - 45 * 1000).toISOString()
      });
    }

    if (method === "GET" && path === "/v1/nodes") {
      return json(route, 200, {
        nodes: [
          {
            node_id: "node_az_21",
            operator_id: "ops-team",
            runtime_type: "hermes",
            runtime_version: "1.9",
            status: "healthy",
            last_heartbeat_at: new Date(Date.now() - 45 * 1000).toISOString(),
            active_leases: 2
          }
        ]
      });
    }

    if (method === "GET" && path === "/v1/nodes/node_az_21/leases") {
      return json(route, 200, { leases: [{ lease_id: "lease_01" }, { lease_id: "lease_02" }] });
    }

    if (method === "GET" && path === "/v1/reliability/subjects/node/node_az_21") {
      return json(route, 200, { rf_value: 0.96, components: { last_100: 0.95, last_30d: 0.97 } });
    }

    if (method === "GET" && path === "/v1/operators/op_1/reliability") {
      return json(route, 200, { rf_value: 0.94, components: { last_100: 0.93, last_30d: 0.95 } });
    }

    if (method === "GET" && path === "/v1/ledger/accounts/acct_ops/entries") {
      return json(route, 200, {
        entries: [
          {
            entry_id: "led_1",
            result_id: "res_01",
            event_type: "settlement.posted",
            net_amount: 123.45,
            created_at: new Date(Date.now() - 30 * 60 * 1000).toISOString(),
            status: "posted"
          }
        ]
      });
    }

    if (method === "GET" && path === "/v1/pricing/rate-cards") {
      return json(route, 200, {
        rate_cards: [
          { family: "research.extract", base_rate: 1.2, min_rate: 0.9, max_rate: 2.4 },
          { family: "doc.section.write", base_rate: 1.1, min_rate: 0.8, max_rate: 2.1 }
        ]
      });
    }

    if (method === "GET" && path === "/v1/policies") {
      return json(route, 200, {
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
    }

    return json(route, 404, {
      error: {
        code: "not_found",
        message: `No mock for ${method} ${path}`
      }
    });
  });
}

test.describe("M5 operator UI", () => {
  test.beforeEach(async ({ page }) => {
    await mockArchonApi(page);
  });

  test("loads dashboard and tasks from /v1 APIs", async ({ page }) => {
    await page.goto("/dashboard");

    await expect(page.getByRole("heading", { name: "System Pulse" })).toBeVisible();
    await expect(page.getByText("Extract partner compliance obligations")).toBeVisible();
    await expect(page.getByText("apr_301")).toBeVisible();

    await page.getByRole("link", { name: "Tasks Lifecycle + traces" }).click();
    await expect(page).toHaveURL(/\/tasks$/);
    await expect(page.getByText("Draft SOC2 narrative")).toBeVisible();
  });

  test("renders task detail tabs from API data", async ({ page }) => {
    await page.goto("/tasks/task_002");

    await expect(page.getByRole("heading", { name: "Draft SOC2 narrative" })).toBeVisible();

    await page.getByRole("tab", { name: "Artifacts" }).click();
    await expect(page.getByText("soc2-draft.json")).toBeVisible();

    await page.getByRole("tab", { name: "Verification" }).click();
    await expect(page.getByText("Result res_01 is completed")).toBeVisible();
  });

  test("enforces role-aware admin surface", async ({ page }) => {
    await page.goto("/admin");
    await expect(page.getByText(/Access denied for operator/i)).toBeVisible();

    await page.evaluate(() => {
      localStorage.setItem(
        "archonhq.operator-context",
        JSON.stringify({ name: "Admin User", role: "tenant_admin", tenantId: "ten_01" })
      );
    });

    await page.reload();
    await expect(page.getByText(/Access granted for tenant_admin/i)).toBeVisible();
  });

  test("loads provider policies from /v1/policies", async ({ page }) => {
    await page.goto("/settings/providers");
    await expect(page.getByText(/Provider policies API unavailable/i)).toHaveCount(0);
    await expect(page.getByText("OpenAI")).toBeVisible();
    await expect(page.getByText(/Current model:\s*gpt-5/i)).toBeVisible();
  });

  test("shows provider fallback when /v1/policies is unavailable", async ({ page }) => {
    await page.route("**/v1/policies", async (route) => {
      await json(route, 501, {
        error: {
          code: "not_implemented",
          message: "policies endpoint not implemented"
        }
      });
    });

    await page.goto("/settings/providers");
    await expect(page.getByText(/Provider policies API unavailable/i)).toBeVisible();
    await expect(page.getByText("OpenAI")).toBeVisible();
  });
});
