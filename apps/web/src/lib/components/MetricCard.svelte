<script lang="ts">
  import TrendingDown from "@lucide/svelte/icons/trending-down";
  import TrendingUp from "@lucide/svelte/icons/trending-up";
  import Minus from "@lucide/svelte/icons/minus";
  import { Card, CardContent, CardHeader, CardTitle } from "$lib/components/ui/card";
  import type { Metric } from "$lib/types";

  let { metric }: { metric: Metric } = $props();

  const changeTone = {
    up: "text-emerald-600",
    down: "text-amber-600",
    flat: "text-slate-500"
  } as const;
</script>

<Card class="metric-glow border-slate-200/70 bg-white/85 shadow-sm backdrop-blur-sm">
  <CardHeader class="pb-3">
    <CardTitle class="text-sm font-semibold tracking-wide text-slate-600">{metric.label}</CardTitle>
  </CardHeader>
  <CardContent>
    <p class="text-3xl font-semibold tracking-tight text-slate-900">{metric.value}</p>
    <p class={`mt-2 inline-flex items-center gap-1 text-xs font-medium ${changeTone[metric.direction]}`}>
      {#if metric.direction === "up"}
        <TrendingUp class="size-3.5" />
      {:else if metric.direction === "down"}
        <TrendingDown class="size-3.5" />
      {:else}
        <Minus class="size-3.5" />
      {/if}
      {metric.change}
    </p>
  </CardContent>
</Card>

<style>
  .metric-glow {
    box-shadow: 0 10px 30px -25px rgba(15, 23, 42, 0.7);
  }
</style>
