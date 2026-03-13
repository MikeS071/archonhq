<script lang="ts">
  import type { ReliabilityPoint } from "$lib/types";

  let { points }: { points: ReliabilityPoint[] } = $props();

  const maxValue = $derived(Math.max(...points.map((point) => point.value), 100));
</script>

<div class="rounded-xl border border-slate-200/80 bg-white/85 p-4 shadow-sm">
  <div class="flex items-end gap-3">
    {#each points as point (point.label)}
      <div class="flex flex-1 flex-col items-center gap-2">
        <div class="relative flex h-44 w-full items-end rounded-md bg-slate-100/80">
          <div
            class="w-full rounded-md bg-gradient-to-t from-indigo-500 to-cyan-400"
            style={`height: ${(point.value / maxValue) * 100}%`}
            title={`${point.label}: ${point.value}%`}
          ></div>
        </div>
        <p class="text-xs font-medium text-slate-600">{point.label}</p>
        <p class="text-xs font-semibold text-slate-900">{point.value}%</p>
      </div>
    {/each}
  </div>
</div>
