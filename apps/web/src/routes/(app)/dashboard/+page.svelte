<script lang="ts">
  import ApprovalQueue from "$lib/components/ApprovalQueue.svelte";
  import MetricCard from "$lib/components/MetricCard.svelte";
  import TaskTable from "$lib/components/TaskTable.svelte";
  import { fetchDashboard } from "$lib/api/services";

  const dashboardPromise = fetchDashboard();
</script>

<section class="space-y-6">
  <header class="flex flex-wrap items-end justify-between gap-3">
    <div>
      <p class="font-mono text-xs tracking-[0.22em] text-slate-500">M5 DASHBOARD</p>
      <h2 class="font-display text-3xl tracking-tight text-slate-900">System Pulse</h2>
      <p class="mt-1 text-sm text-slate-600">Live operator state across tasks, approvals, and settlements.</p>
    </div>
  </header>

  {#await dashboardPromise}
    <div class="rounded-xl border border-slate-200/80 bg-white/85 p-6 text-sm text-slate-600">Loading dashboard data...</div>
  {:then data}
    <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
      {#each data.metrics as metric (metric.label)}
        <MetricCard {metric} />
      {/each}
    </div>

    <div class="space-y-3">
      <h3 class="font-display text-xl text-slate-900">Active tasks</h3>
      <TaskTable tasks={data.tasks.slice(0, 6)} />
    </div>

    <div class="space-y-3">
      <h3 class="font-display text-xl text-slate-900">Approval queue</h3>
      <ApprovalQueue approvals={data.approvals} />
    </div>
  {:catch err}
    <div class="rounded-xl border border-rose-200 bg-rose-50 p-6 text-sm text-rose-700">
      Failed to load dashboard API data: {err.message}
    </div>
  {/await}
</section>
