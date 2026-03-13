<script lang="ts">
  import { page } from "$app/stores";
  import { navForRole } from "$lib/navigation";
  import { operatorContext } from "$lib/stores/operator-context";
  import type { Role } from "$lib/types";
  import * as Select from "$lib/components/ui/select";

  const roles: Role[] = [
    "platform_admin",
    "tenant_admin",
    "operator",
    "approver",
    "auditor",
    "finance_viewer",
    "developer"
  ];

  const roleLabels: Record<Role, string> = {
    platform_admin: "Platform Admin",
    tenant_admin: "Tenant Admin",
    operator: "Operator",
    approver: "Approver",
    auditor: "Auditor",
    finance_viewer: "Finance Viewer",
    developer: "Developer"
  };

  let { children } = $props();
</script>

<div class="shell relative min-h-screen overflow-hidden px-4 py-5 sm:px-8">
  <div class="orb orb-a"></div>
  <div class="orb orb-b"></div>

  <div class="relative mx-auto grid max-w-7xl gap-6 lg:grid-cols-[260px_1fr]">
    <aside class="rounded-2xl border border-white/40 bg-white/70 p-4 shadow-xl backdrop-blur-xl">
      <div class="mb-6">
        <p class="font-mono text-xs tracking-[0.22em] text-slate-500">ARCHONHQ</p>
        <h1 class="mt-2 font-display text-2xl tracking-tight text-slate-900">Operator Console</h1>
      </div>

      <div class="mb-6 space-y-2">
        <p class="text-xs font-semibold tracking-wide text-slate-600">Acting role</p>
        <Select.Root
          type="single"
          value={$operatorContext.role}
          onValueChange={(role) => operatorContext.setRole(role as Role)}
        >
          <Select.Trigger class="w-full bg-white">{roleLabels[$operatorContext.role]}</Select.Trigger>
          <Select.Content>
            {#each roles as role}
              <Select.Item value={role} label={roleLabels[role]} />
            {/each}
          </Select.Content>
        </Select.Root>
        <p class="text-xs text-slate-500">{roleLabels[$operatorContext.role]} access simulation</p>
      </div>

      <nav class="space-y-1">
        {#each navForRole($operatorContext.role) as item}
          <a
            href={item.href}
            class={`nav-link ${$page.url.pathname === item.href ? "nav-link-active" : ""}`}
          >
            <span>{item.label}</span>
            <span class="text-[11px] text-slate-500">{item.description}</span>
          </a>
        {/each}
      </nav>
    </aside>

    <main class="rounded-2xl border border-white/40 bg-white/60 p-4 shadow-xl backdrop-blur-xl sm:p-6">
      {@render children?.()}
    </main>
  </div>
</div>

<style>
  .shell {
    background:
      radial-gradient(circle at 10% 10%, rgba(14, 165, 233, 0.14), transparent 42%),
      radial-gradient(circle at 90% 2%, rgba(245, 158, 11, 0.15), transparent 30%),
      linear-gradient(180deg, #f5f7fb 0%, #edf3ff 100%);
  }

  .orb {
    pointer-events: none;
    position: absolute;
    border-radius: 9999px;
    filter: blur(48px);
    opacity: 0.45;
  }

  .orb-a {
    top: -3rem;
    right: -6rem;
    width: 16rem;
    height: 16rem;
    background: #67e8f9;
  }

  .orb-b {
    bottom: -5rem;
    left: -5rem;
    width: 18rem;
    height: 18rem;
    background: #fbbf24;
  }

  .nav-link {
    display: flex;
    flex-direction: column;
    gap: 2px;
    border-radius: 0.75rem;
    padding: 0.65rem 0.75rem;
    font-size: 0.92rem;
    font-weight: 600;
    color: #1e293b;
    transition: background 120ms ease;
  }

  .nav-link:hover {
    background: rgba(148, 163, 184, 0.18);
  }

  .nav-link-active {
    background: linear-gradient(135deg, rgba(99, 102, 241, 0.2), rgba(56, 189, 248, 0.2));
    color: #0f172a;
  }

  @media (max-width: 1023px) {
    .shell {
      padding: 1rem;
    }
  }
</style>
