<script lang="ts">
  import ApprovalQueue from "$lib/components/ApprovalQueue.svelte";
  import { fetchApprovals } from "$lib/api/services";

  const approvalsPromise = fetchApprovals();
</script>

<section class="space-y-4">
  <header>
    <p class="font-mono text-xs tracking-[0.22em] text-slate-500">HUMAN GATE</p>
    <h2 class="font-display text-3xl tracking-tight text-slate-900">Approvals Queue</h2>
    <p class="mt-1 text-sm text-slate-600">Review high-impact operations before state mutation.</p>
  </header>

  {#await approvalsPromise}
    <div class="rounded-xl border border-slate-200/80 bg-white/85 p-6 text-sm text-slate-600">Loading approvals queue...</div>
  {:then approvals}
    <ApprovalQueue {approvals} />
  {:catch err}
    <div class="rounded-xl border border-rose-200 bg-rose-50 p-6 text-sm text-rose-700">
      Failed to load approvals queue: {err.message}
    </div>
  {/await}
</section>
