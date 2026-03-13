<script lang="ts">
  import type { TaskEvent } from "$lib/types";

  let { events }: { events: TaskEvent[] } = $props();
</script>

<ol class="space-y-4">
  {#each events as event (event.id)}
    <li class="relative rounded-lg border border-slate-200/80 bg-white/85 p-4 pl-8">
      <span class="absolute top-6 left-3 h-2.5 w-2.5 rounded-full bg-indigo-500"></span>
      <p class="text-sm font-semibold text-slate-900">{event.label}</p>
      <p class="mt-1 text-xs text-slate-600">{event.actor}</p>
      <p class="mt-1 text-xs font-mono text-slate-500">{event.at}</p>
    </li>
  {/each}
</ol>
