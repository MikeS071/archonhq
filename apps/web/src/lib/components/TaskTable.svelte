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
  import type { Task } from "$lib/types";

  let { tasks, showWorkspace = true }: { tasks: Task[]; showWorkspace?: boolean } = $props();

  const statusTone: Record<Task["status"], string> = {
    queued: "bg-slate-100 text-slate-700",
    leased: "bg-cyan-100 text-cyan-700",
    running: "bg-indigo-100 text-indigo-700",
    awaiting_approval: "bg-amber-100 text-amber-700",
    completed: "bg-emerald-100 text-emerald-700",
    failed: "bg-rose-100 text-rose-700"
  };

  const priorityTone: Record<Task["priority"], string> = {
    low: "bg-slate-100 text-slate-700",
    medium: "bg-sky-100 text-sky-700",
    high: "bg-orange-100 text-orange-700",
    urgent: "bg-red-100 text-red-700"
  };
</script>

<div class="rounded-xl border border-slate-200/80 bg-white/85 p-1 shadow-sm">
  <Table>
    <TableHeader>
      <TableRow>
        <TableHead>Task</TableHead>
        <TableHead>Family</TableHead>
        <TableHead>Status</TableHead>
        <TableHead>Priority</TableHead>
        {#if showWorkspace}
          <TableHead>Workspace</TableHead>
        {/if}
        <TableHead>Node</TableHead>
        <TableHead>JouleWork</TableHead>
        <TableHead>Updated</TableHead>
        <TableHead class="text-right">Action</TableHead>
      </TableRow>
    </TableHeader>
    <TableBody>
      {#each tasks as task (task.id)}
        <TableRow>
          <TableCell class="font-medium text-slate-900">{task.title}</TableCell>
          <TableCell class="text-slate-600">{task.family}</TableCell>
          <TableCell>
            <Badge class={statusTone[task.status]}>{task.status.replaceAll("_", " ")}</Badge>
          </TableCell>
          <TableCell>
            <Badge class={priorityTone[task.priority]}>{task.priority}</Badge>
          </TableCell>
          {#if showWorkspace}
            <TableCell class="font-mono text-xs text-slate-600">{task.workspaceId}</TableCell>
          {/if}
          <TableCell class="font-mono text-xs text-slate-600">{task.assignee}</TableCell>
          <TableCell>{task.joules}</TableCell>
          <TableCell>{task.updatedAt}</TableCell>
          <TableCell class="text-right">
            <Button href={`/tasks/${task.id}`} variant="outline" size="sm">Open</Button>
          </TableCell>
        </TableRow>
      {/each}
    </TableBody>
  </Table>
</div>
