<script lang="ts">
  import ShieldAlert from "@lucide/svelte/icons/shield-alert";
  import ShieldCheck from "@lucide/svelte/icons/shield-check";
  import { canAccessRoute } from "$lib/auth/guards";
  import { Card, CardContent, CardHeader, CardTitle } from "$lib/components/ui/card";
  import { operatorContext } from "$lib/stores/operator-context";

  const hasAccess = $derived(canAccessRoute("/admin", $operatorContext.role));
</script>

<section class="space-y-4">
  <header>
    <p class="font-mono text-xs tracking-[0.22em] text-slate-500">ADMIN</p>
    <h2 class="font-display text-3xl tracking-tight text-slate-900">Tenant Controls</h2>
  </header>

  <Card class="border-slate-200/80 bg-white/85">
    <CardHeader>
      <CardTitle class="flex items-center gap-2">
        {#if hasAccess}
          <ShieldCheck class="size-5 text-emerald-600" />
          Access granted for {$operatorContext.role}
        {:else}
          <ShieldAlert class="size-5 text-rose-600" />
          Access denied for {$operatorContext.role}
        {/if}
      </CardTitle>
    </CardHeader>
    <CardContent class="text-sm text-slate-700">
      {#if hasAccess}
        <p>Admin actions available: role grants, tenant policy edits, and integration controls.</p>
      {:else}
        <p>Switch role to <code>tenant_admin</code> or <code>platform_admin</code> to manage this surface.</p>
      {/if}
    </CardContent>
  </Card>
</section>
