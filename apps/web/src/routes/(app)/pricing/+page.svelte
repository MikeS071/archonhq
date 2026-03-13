<script lang="ts">
  import { Card, CardContent, CardHeader, CardTitle } from "$lib/components/ui/card";
  import { fetchPricing } from "$lib/api/services";

  const pricingPromise = fetchPricing();
</script>

<section class="space-y-4">
  <header>
    <p class="font-mono text-xs tracking-[0.22em] text-slate-500">MARKET</p>
    <h2 class="font-display text-3xl tracking-tight text-slate-900">Pricing & Quotes</h2>
    <p class="mt-1 text-sm text-slate-600">Live rate-card payload from <code>/v1/pricing/rate-cards</code>.</p>
  </header>

  {#await pricingPromise}
    <div class="rounded-xl border border-slate-200/80 bg-white/85 p-6 text-sm text-slate-600">Loading pricing rate cards...</div>
  {:then data}
    <Card class="border-slate-200/80 bg-white/85">
      <CardHeader>
        <CardTitle>Rate Cards</CardTitle>
      </CardHeader>
      <CardContent class="space-y-3">
        {#if data.rateCards.length === 0}
          <p class="text-sm text-slate-600">No rate cards returned.</p>
        {:else}
          {#each data.rateCards as card, idx (`rate-card-${idx}`)}
            <pre class="overflow-auto rounded-lg bg-slate-900 p-3 text-xs text-slate-100">{JSON.stringify(card, null, 2)}</pre>
          {/each}
        {/if}
      </CardContent>
    </Card>
  {:catch err}
    <div class="rounded-xl border border-rose-200 bg-rose-50 p-6 text-sm text-rose-700">
      Failed to load pricing rate cards: {err.message}
    </div>
  {/await}
</section>
