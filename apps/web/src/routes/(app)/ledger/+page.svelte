<script lang="ts">
  import LedgerEntriesTable from "$lib/components/LedgerEntriesTable.svelte";
  import { defaultAccountId, fetchLedgerEntries } from "$lib/api/services";

  const accountId = defaultAccountId();
  const ledgerPromise = fetchLedgerEntries(accountId);
</script>

<section class="space-y-4">
  <header>
    <p class="font-mono text-xs tracking-[0.22em] text-slate-500">ECONOMICS</p>
    <h2 class="font-display text-3xl tracking-tight text-slate-900">Ledger</h2>
    <p class="mt-1 text-sm text-slate-600">
      Track reserve holds, settlements, and operator earnings movement for <code>{accountId}</code>.
    </p>
  </header>

  {#await ledgerPromise}
    <div class="rounded-xl border border-slate-200/80 bg-white/85 p-6 text-sm text-slate-600">Loading ledger entries...</div>
  {:then entries}
    <LedgerEntriesTable {entries} />
  {:catch err}
    <div class="rounded-xl border border-rose-200 bg-rose-50 p-6 text-sm text-rose-700">
      Failed to load ledger entries: {err.message}
    </div>
  {/await}
</section>
