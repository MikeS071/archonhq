<script lang="ts">
  import ReliabilityChart from "$lib/components/ReliabilityChart.svelte";
  import { defaultOperatorId, fetchReliabilitySeries } from "$lib/api/services";

  const operatorId = defaultOperatorId();
  const reliabilityPromise = fetchReliabilitySeries(operatorId);
</script>

<section class="space-y-4">
  <header>
    <p class="font-mono text-xs tracking-[0.22em] text-slate-500">QUALITY</p>
    <h2 class="font-display text-3xl tracking-tight text-slate-900">Reliability</h2>
    <p class="mt-1 text-sm text-slate-600">Current reliability snapshot components for operator <code>{operatorId}</code>.</p>
  </header>

  {#await reliabilityPromise}
    <div class="rounded-xl border border-slate-200/80 bg-white/85 p-6 text-sm text-slate-600">Loading reliability snapshot...</div>
  {:then points}
    <ReliabilityChart {points} />
  {:catch err}
    <div class="rounded-xl border border-rose-200 bg-rose-50 p-6 text-sm text-rose-700">
      Failed to load reliability snapshot: {err.message}
    </div>
  {/await}
</section>
