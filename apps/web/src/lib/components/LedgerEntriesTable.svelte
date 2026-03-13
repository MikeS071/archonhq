<script lang="ts">
  import { Badge } from "$lib/components/ui/badge";
  import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow
  } from "$lib/components/ui/table";
  import type { LedgerEntry } from "$lib/types";

  let { entries }: { entries: LedgerEntry[] } = $props();
</script>

<div class="rounded-xl border border-slate-200/80 bg-white/85 p-1 shadow-sm">
  <Table>
    <TableHeader>
      <TableRow>
        <TableHead>Entry</TableHead>
        <TableHead>Account</TableHead>
        <TableHead>Type</TableHead>
        <TableHead>Description</TableHead>
        <TableHead>Posted</TableHead>
        <TableHead class="text-right">Amount (USD)</TableHead>
      </TableRow>
    </TableHeader>
    <TableBody>
      {#each entries as entry (entry.id)}
        <TableRow>
          <TableCell class="font-mono text-xs">{entry.id}</TableCell>
          <TableCell class="font-medium">{entry.accountId}</TableCell>
          <TableCell>
            <Badge class={entry.type === "credit" ? "bg-emerald-100 text-emerald-700" : "bg-rose-100 text-rose-700"}>
              {entry.type}
            </Badge>
          </TableCell>
          <TableCell>{entry.description}</TableCell>
          <TableCell>{entry.postedAt}</TableCell>
          <TableCell class={`text-right font-semibold ${entry.type === "credit" ? "text-emerald-700" : "text-rose-700"}`}>
            {entry.type === "credit" ? "+" : "-"}${entry.amount.toFixed(2)}
          </TableCell>
        </TableRow>
      {/each}
    </TableBody>
  </Table>
</div>
