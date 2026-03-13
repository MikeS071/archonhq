import { browser } from "$app/environment";
import { derived, writable } from "svelte/store";
import type { Role } from "$lib/types";

const STORAGE_KEY = "archonhq.operator-context";

export interface OperatorContext {
  name: string;
  role: Role;
  tenantId: string;
}

const defaultContext: OperatorContext = {
  name: "Avery Chen",
  role: "operator",
  tenantId: "ten_01"
};

function createOperatorContextStore() {
  const initial = browser ? readStoredContext() : defaultContext;
  const { subscribe, update, set } = writable<OperatorContext>(initial);

  if (browser) {
    subscribe((value) => {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(value));
    });
  }

  return {
    subscribe,
    set,
    setRole(role: Role) {
      update((ctx) => ({ ...ctx, role }));
    }
  };
}

function readStoredContext(): OperatorContext {
  const raw = localStorage.getItem(STORAGE_KEY);

  if (!raw) {
    return defaultContext;
  }

  try {
    const parsed = JSON.parse(raw) as Partial<OperatorContext>;

    if (!parsed.role || !parsed.name || !parsed.tenantId) {
      return defaultContext;
    }

    return {
      name: parsed.name,
      role: parsed.role,
      tenantId: parsed.tenantId
    };
  } catch {
    return defaultContext;
  }
}

export const operatorContext = createOperatorContextStore();

export const isAdminRole = derived(operatorContext, ($ctx) => {
  return $ctx.role === "platform_admin" || $ctx.role === "tenant_admin";
});
