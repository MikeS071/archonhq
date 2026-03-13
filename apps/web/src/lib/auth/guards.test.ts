import { describe, expect, it } from "vitest";
import { canAccessRoute } from "$lib/auth/guards";

describe("canAccessRoute", () => {
  it("allows open pages for any role", () => {
    expect(canAccessRoute("/tasks", "developer")).toBe(true);
    expect(canAccessRoute("/fleet", "auditor")).toBe(true);
  });

  it("restricts admin page to tenant/platform admins", () => {
    expect(canAccessRoute("/admin", "platform_admin")).toBe(true);
    expect(canAccessRoute("/admin", "tenant_admin")).toBe(true);
    expect(canAccessRoute("/admin", "operator")).toBe(false);
    expect(canAccessRoute("/admin", "approver")).toBe(false);
  });

  it("allows pricing for finance and operators", () => {
    expect(canAccessRoute("/pricing", "finance_viewer")).toBe(true);
    expect(canAccessRoute("/pricing", "operator")).toBe(true);
    expect(canAccessRoute("/pricing", "developer")).toBe(false);
  });
});
