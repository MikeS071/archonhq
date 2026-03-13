<script lang="ts">
  import { Badge } from "$lib/components/ui/badge";
  import type { Verification } from "$lib/types";

  let { items }: { items: Verification[] } = $props();

  const verdictTone: Record<Verification["verdict"], string> = {
    pass: "bg-emerald-100 text-emerald-700",
    warn: "bg-amber-100 text-amber-700",
    fail: "bg-rose-100 text-rose-700"
  };
</script>

<ul class="space-y-3">
  {#each items as item (item.id)}
    <li class="rounded-lg border border-slate-200/80 bg-white/85 p-4">
      <div class="flex flex-wrap items-center justify-between gap-2">
        <p class="font-medium text-slate-900">{item.summary}</p>
        <Badge class={verdictTone[item.verdict]}>{item.verdict}</Badge>
      </div>
      <p class="mt-2 font-mono text-xs text-slate-500">{item.id} · {item.createdAt}</p>
    </li>
  {/each}
</ul>
