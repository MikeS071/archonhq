<script lang="ts">
  import ArtifactList from "$lib/components/ArtifactList.svelte";
  import TaskTimeline from "$lib/components/TaskTimeline.svelte";
  import VerificationList from "$lib/components/VerificationList.svelte";
  import { fetchTaskDetail } from "$lib/api/services";
  import { Badge } from "$lib/components/ui/badge";
  import { Card, CardContent, CardHeader, CardTitle } from "$lib/components/ui/card";
  import * as Tabs from "$lib/components/ui/tabs";

  let { data }: { data: { taskId: string } } = $props();

  const detailPromise = $derived(fetchTaskDetail(data.taskId));
</script>

<section class="space-y-5">
  {#await detailPromise}
    <div class="rounded-xl border border-slate-200/80 bg-white/85 p-6 text-sm text-slate-600">Loading task detail...</div>
  {:then detail}
    <header class="space-y-1">
      <p class="font-mono text-xs tracking-[0.22em] text-slate-500">TASK DETAIL</p>
      <h2 class="font-display text-3xl tracking-tight text-slate-900">{detail.task.title}</h2>
      <div class="flex flex-wrap items-center gap-2 text-sm text-slate-600">
        <Badge class="bg-indigo-100 text-indigo-700">{detail.task.id}</Badge>
        <span>family: {detail.task.family}</span>
        <span>workspace: {detail.task.workspaceId}</span>
        <span>updated: {detail.task.updatedAt}</span>
      </div>
    </header>

    <Card class="border-slate-200/80 bg-white/85">
      <CardHeader>
        <CardTitle>Execution trace</CardTitle>
      </CardHeader>
      <CardContent>
        <Tabs.Root value="timeline" class="w-full">
          <Tabs.List>
            <Tabs.Trigger value="timeline">Timeline</Tabs.Trigger>
            <Tabs.Trigger value="artifacts">Artifacts</Tabs.Trigger>
            <Tabs.Trigger value="verification">Verification</Tabs.Trigger>
          </Tabs.List>
          <Tabs.Content value="timeline" class="mt-4">
            <TaskTimeline events={detail.events} />
          </Tabs.Content>
          <Tabs.Content value="artifacts" class="mt-4">
            {#if detail.artifacts.length === 0}
              <p class="text-sm text-slate-600">No artifacts attached to task results yet.</p>
            {:else}
              <ArtifactList artifacts={detail.artifacts} />
            {/if}
          </Tabs.Content>
          <Tabs.Content value="verification" class="mt-4">
            {#if detail.verifications.length === 0}
              <p class="text-sm text-slate-600">No verification records returned by API.</p>
            {:else}
              <VerificationList items={detail.verifications} />
            {/if}
          </Tabs.Content>
        </Tabs.Root>
      </CardContent>
    </Card>
  {:catch err}
    <div class="rounded-xl border border-rose-200 bg-rose-50 p-6 text-sm text-rose-700">
      Failed to load task detail: {err.message}
    </div>
  {/await}
</section>
