import type { Role } from "$lib/types";
import { canAccessRoute } from "$lib/auth/guards";

export interface AppNavItem {
  label: string;
  href: string;
  description: string;
}

const allNavItems: AppNavItem[] = [
  { label: "Dashboard", href: "/dashboard", description: "Operational pulse" },
  { label: "Tasks", href: "/tasks", description: "Lifecycle + traces" },
  { label: "Approvals", href: "/approvals", description: "Human gates" },
  { label: "Fleet", href: "/fleet", description: "Node health" },
  { label: "Ledger", href: "/ledger", description: "Settlement flow" },
  { label: "Reliability", href: "/reliability", description: "Scoring trend" },
  { label: "Pricing", href: "/pricing", description: "Quotes + bids" },
  { label: "Providers", href: "/settings/providers", description: "Model policy" },
  { label: "Admin", href: "/admin", description: "Tenant controls" }
];

export function navForRole(role: Role): AppNavItem[] {
  return allNavItems.filter((item) => canAccessRoute(item.href, role));
}
