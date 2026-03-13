<script lang="ts">
  import FileArchive from "@lucide/svelte/icons/file-archive";
  import FileText from "@lucide/svelte/icons/file-text";
  import Database from "@lucide/svelte/icons/database";
  import type { Artifact } from "$lib/types";

  let { artifacts }: { artifacts: Artifact[] } = $props();

  function iconFor(kind: string) {
    if (kind === "dataset") return Database;
    if (kind === "document") return FileText;
    return FileArchive;
  }
</script>

<ul class="grid gap-3 sm:grid-cols-2">
  {#each artifacts as artifact (artifact.id)}
    {@const Icon = iconFor(artifact.kind)}
    <li class="rounded-lg border border-slate-200/80 bg-white/85 p-4">
      <div class="flex items-start gap-3">
        <Icon class="mt-0.5 size-4 text-indigo-600" />
        <div>
          <p class="font-medium text-slate-900">{artifact.name}</p>
          <p class="text-xs text-slate-600">{artifact.kind} • {artifact.size}</p>
        </div>
      </div>
    </li>
  {/each}
</ul>
