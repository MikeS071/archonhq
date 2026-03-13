<script lang="ts">
  import { Input } from "$lib/components/ui/input";
  import TaskTable from "$lib/components/TaskTable.svelte";
  import { fetchTaskFeed } from "$lib/api/services";

  let query = $state("");
  const tasksPromise = fetchTaskFeed(100);
</script>

<section class="space-y-4">
  <header>
    <p class="font-mono text-xs tracking-[0.22em] text-slate-500">WORKFLOW</p>
    <h2 class="font-display text-3xl tracking-tight text-slate-900">Tasks</h2>
    <p class="mt-1 text-sm text-slate-600">Track queue, lease state, and JouleWork burn by workspace.</p>
  </header>

  <div class="max-w-md">
    <Input bind:value={query} placeholder="Search task id, title, family" aria-label="search tasks" />
  </div>

  {#await tasksPromise}
    <div class="rounded-xl border border-slate-200/80 bg-white/85 p-6 text-sm text-slate-600">Loading task feed...</div>
  {:then tasks}
    {@const filteredTasks = tasks.filter((task) => {
      if (!query.trim()) return true;

      const q = query.toLowerCase();
      return (
        task.id.toLowerCase().includes(q) ||
        task.title.toLowerCase().includes(q) ||
        task.family.toLowerCase().includes(q)
      );
    })}
    <TaskTable tasks={filteredTasks} />
  {:catch err}
    <div class="rounded-xl border border-rose-200 bg-rose-50 p-6 text-sm text-rose-700">
      Failed to load task feed: {err.message}
    </div>
  {/await}
</section>
