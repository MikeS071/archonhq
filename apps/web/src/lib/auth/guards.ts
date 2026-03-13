import type { Role } from "$lib/types";

const routeRoleRequirements: Record<string, Role[]> = {
  "/admin": ["platform_admin", "tenant_admin"],
  "/settings/providers": ["platform_admin", "tenant_admin"],
  "/pricing": ["platform_admin", "tenant_admin", "operator", "finance_viewer"]
};

export function canAccessRoute(pathname: string, role: Role): boolean {
  const requiredRoles = routeRoleRequirements[pathname];

  if (!requiredRoles) {
    return true;
  }

  return requiredRoles.includes(role);
}
