import { describe, expect, it } from "vitest";
import { navForRole } from "$lib/navigation";

describe("navForRole", () => {
  it("includes admin links for tenant admin", () => {
    const nav = navForRole("tenant_admin");

    expect(nav.some((item) => item.href === "/admin")).toBe(true);
    expect(nav.some((item) => item.href === "/settings/providers")).toBe(true);
  });

  it("hides admin links for developer role", () => {
    const nav = navForRole("developer");

    expect(nav.some((item) => item.href === "/admin")).toBe(false);
    expect(nav.some((item) => item.href === "/settings/providers")).toBe(false);
  });
});
