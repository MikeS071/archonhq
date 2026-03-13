<script lang="ts">
  import { Badge } from "$lib/components/ui/badge";
  import { Button } from "$lib/components/ui/button";
  import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow
  } from "$lib/components/ui/table";
  import type { FleetNode } from "$lib/types";

  let { nodes }: { nodes: FleetNode[] } = $props();

  const statusTone: Record<FleetNode["status"], string> = {
    healthy: "bg-emerald-100 text-emerald-700",
    degraded: "bg-amber-100 text-amber-700",
    offline: "bg-rose-100 text-rose-700"
  };
</script>

<div class="rounded-xl border border-slate-200/80 bg-white/85 p-1 shadow-sm">
  <Table>
    <TableHeader>
      <TableRow>
        <TableHead>Node</TableHead>
        <TableHead>Operator</TableHead>
        <TableHead>Runtime</TableHead>
        <TableHead>Status</TableHead>
        <TableHead>Heartbeat</TableHead>
        <TableHead>Leases</TableHead>
        <TableHead>Reliability</TableHead>
        <TableHead class="text-right">Action</TableHead>
      </TableRow>
    </TableHeader>
    <TableBody>
      {#each nodes as node (node.id)}
        <TableRow>
          <TableCell class="font-medium text-slate-900">{node.id}</TableCell>
          <TableCell>{node.operator}</TableCell>
          <TableCell>{node.runtime}</TableCell>
          <TableCell>
            <Badge class={statusTone[node.status]}>{node.status}</Badge>
          </TableCell>
          <TableCell>{node.lastHeartbeat}</TableCell>
          <TableCell>{node.activeLeases}</TableCell>
          <TableCell>{node.reliability}%</TableCell>
          <TableCell class="text-right">
            <Button href={`/fleet/nodes/${node.id}`} variant="outline" size="sm">Inspect</Button>
          </TableCell>
        </TableRow>
      {/each}
    </TableBody>
  </Table>
</div>
