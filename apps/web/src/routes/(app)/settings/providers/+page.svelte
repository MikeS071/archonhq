<script lang="ts">
  import PolicyEditor from "$lib/components/PolicyEditor.svelte";
  import { fetchProviderPolicies } from "$lib/api/services";
  import { providerPolicies } from "$lib/mock-data";

  const policiesPromise = fetchProviderPolicies();
</script>

<section class="space-y-4">
  <header>
    <p class="font-mono text-xs tracking-[0.22em] text-slate-500">SETTINGS</p>
    <h2 class="font-display text-3xl tracking-tight text-slate-900">Providers</h2>
    <p class="mt-1 text-sm text-slate-600">Model policy from <code>/v1/policies</code>.</p>
  </header>

  {#await policiesPromise}
    <div class="rounded-xl border border-slate-200/80 bg-white/85 p-6 text-sm text-slate-600">Loading provider policies...</div>
  {:then policies}
    <PolicyEditor {policies} />
  {:catch err}
    <div class="space-y-4">
      <div class="rounded-xl border border-amber-200 bg-amber-50 p-4 text-sm text-amber-800">
        Provider policies API unavailable ({err.message}). Showing local editable defaults.
      </div>
      <PolicyEditor policies={providerPolicies} />
    </div>
  {/await}
</section>
