<script lang="ts">
  import { fetchNode } from "$lib/api/services";
  import { Badge } from "$lib/components/ui/badge";
  import { Card, CardContent, CardHeader, CardTitle } from "$lib/components/ui/card";

  let { data }: { data: { nodeId: string } } = $props();

  const nodePromise = $derived(fetchNode(data.nodeId));
</script>

<section class="space-y-4">
  <header>
    <p class="font-mono text-xs tracking-[0.22em] text-slate-500">NODE DETAIL</p>
    <h2 class="font-display text-3xl tracking-tight text-slate-900">{data.nodeId}</h2>
  </header>

  {#await nodePromise}
    <div class="rounded-xl border border-slate-200/80 bg-white/85 p-6 text-sm text-slate-600">Loading node detail...</div>
  {:then node}
    <Card class="border-slate-200/80 bg-white/85">
      <CardHeader>
        <CardTitle>Runtime summary</CardTitle>
      </CardHeader>
      <CardContent class="grid gap-3 text-sm sm:grid-cols-2">
        <p><span class="font-semibold">Operator:</span> {node.operator}</p>
        <p><span class="font-semibold">Runtime:</span> {node.runtime}</p>
        <p><span class="font-semibold">Heartbeat:</span> {node.lastHeartbeat}</p>
        <p><span class="font-semibold">Active leases:</span> {node.activeLeases}</p>
        <p><span class="font-semibold">Reliability:</span> {node.reliability}%</p>
        <p>
          <span class="font-semibold">Status:</span>
          <Badge class="ml-2 bg-sky-100 text-sky-700">{node.status}</Badge>
        </p>
      </CardContent>
    </Card>
  {:catch err}
    <div class="rounded-xl border border-rose-200 bg-rose-50 p-6 text-sm text-rose-700">
      Failed to load node detail: {err.message}
    </div>
  {/await}
</section>
