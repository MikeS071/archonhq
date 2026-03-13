<script lang="ts">
  import FleetTable from "$lib/components/FleetTable.svelte";
  import MetricCard from "$lib/components/MetricCard.svelte";
  import { fetchFleetNodes } from "$lib/api/services";
  import type { Metric } from "$lib/types";

  const fleetPromise = fetchFleetNodes();

  function metricsForNodes(total: number, healthy: number): Metric[] {
    return [
      {
        label: "Tracked Nodes",
        value: String(total),
        change: "from /v1/nodes",
        direction: "flat"
      },
      {
        label: "Healthy",
        value: `${total === 0 ? 0 : Math.round((healthy / total) * 100)}%`,
        change: `${healthy} nodes healthy`,
        direction: healthy > 0 ? "up" : "flat"
      },
      {
        label: "Needs Attention",
        value: String(Math.max(total - healthy, 0)),
        change: "degraded/offline",
        direction: total - healthy > 0 ? "down" : "flat"
      }
    ];
  }
</script>

<section class="space-y-6">
  <header>
    <p class="font-mono text-xs tracking-[0.22em] text-slate-500">INFRA</p>
    <h2 class="font-display text-3xl tracking-tight text-slate-900">Fleet</h2>
    <p class="mt-1 text-sm text-slate-600">Runtime health, lease load, and reliability posture by node.</p>
  </header>

  {#await fleetPromise}
    <div class="rounded-xl border border-slate-200/80 bg-white/85 p-6 text-sm text-slate-600">Loading fleet nodes...</div>
  {:then nodes}
    {@const healthy = nodes.filter((node) => node.status === "healthy").length}
    <div class="grid gap-4 md:grid-cols-3">
      {#each metricsForNodes(nodes.length, healthy) as metric (metric.label)}
        <MetricCard {metric} />
      {/each}
    </div>

    <FleetTable {nodes} />
  {:catch err}
    <div class="rounded-xl border border-rose-200 bg-rose-50 p-6 text-sm text-rose-700">
      Failed to load fleet nodes: {err.message}
    </div>
  {/await}
</section>
