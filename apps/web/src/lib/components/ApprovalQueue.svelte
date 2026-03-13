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
  import type { ApprovalItem } from "$lib/types";

  let { approvals }: { approvals: ApprovalItem[] } = $props();
</script>

<div class="rounded-xl border border-slate-200/80 bg-white/85 p-1 shadow-sm">
  <Table>
    <TableHeader>
      <TableRow>
        <TableHead>Approval</TableHead>
        <TableHead>Task</TableHead>
        <TableHead>Reason</TableHead>
        <TableHead>Requester</TableHead>
        <TableHead>Created</TableHead>
        <TableHead class="text-right">Decision</TableHead>
      </TableRow>
    </TableHeader>
    <TableBody>
      {#each approvals as approval (approval.id)}
        <TableRow>
          <TableCell><Badge class="bg-amber-100 text-amber-700">{approval.id}</Badge></TableCell>
          <TableCell>
            <a class="font-medium text-indigo-700 hover:underline" href={`/tasks/${approval.taskId}`}>{approval.taskId}</a>
          </TableCell>
          <TableCell class="max-w-72 text-slate-700">{approval.reason}</TableCell>
          <TableCell class="font-mono text-xs text-slate-600">{approval.requester}</TableCell>
          <TableCell>{approval.createdAt}</TableCell>
          <TableCell class="text-right">
            <div class="inline-flex gap-2">
              <Button size="sm">Approve</Button>
              <Button size="sm" variant="outline">Deny</Button>
            </div>
          </TableCell>
        </TableRow>
      {/each}
    </TableBody>
  </Table>
</div>
