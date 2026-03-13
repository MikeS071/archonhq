<script lang="ts">
  import { Button } from "$lib/components/ui/button";
  import { Input } from "$lib/components/ui/input";
  import * as Select from "$lib/components/ui/select";
  import { Switch } from "$lib/components/ui/switch";
  import { Textarea } from "$lib/components/ui/textarea";
  import type { ProviderPolicy } from "$lib/types";

  let { policies }: { policies: ProviderPolicy[] } = $props();
</script>

<div class="space-y-4">
  {#each policies as policy (policy.provider)}
    <section class="rounded-xl border border-slate-200/80 bg-white/85 p-4 shadow-sm">
      <header class="mb-3 flex items-center justify-between">
        <div>
          <h3 class="font-semibold text-slate-900">{policy.provider}</h3>
          <p class="text-xs text-slate-600">Current model: {policy.model}</p>
        </div>
        <Switch checked={policy.requiresApproval} />
      </header>
      <div class="grid gap-3 md:grid-cols-2">
        <div class="space-y-2">
          <label class="text-sm font-medium text-slate-700" for={`model-${policy.provider}`}>Model</label>
          <Select.Root type="single" value={policy.model}>
            <Select.Trigger id={`model-${policy.provider}`} class="w-full">{policy.model}</Select.Trigger>
            <Select.Content>
              <Select.Item value={policy.model} label={policy.model} />
              <Select.Item value="fallback" label="fallback" />
            </Select.Content>
          </Select.Root>
        </div>
        <div class="space-y-2">
          <label class="text-sm font-medium text-slate-700" for={`budget-${policy.provider}`}>Max USD / task</label>
          <Input id={`budget-${policy.provider}`} type="number" value={String(policy.maxUsdPerTask)} />
        </div>
        <div class="space-y-2">
          <label class="text-sm font-medium text-slate-700" for={`retries-${policy.provider}`}>Retries</label>
          <Input id={`retries-${policy.provider}`} type="number" value={String(policy.retries)} />
        </div>
        <div class="space-y-2">
          <label class="text-sm font-medium text-slate-700" for={`note-${policy.provider}`}>Operator note</label>
          <Textarea id={`note-${policy.provider}`} placeholder="Reason for this policy" rows={2} />
        </div>
      </div>
    </section>
  {/each}

  <div class="flex justify-end">
    <Button>Save policy draft</Button>
  </div>
</div>
