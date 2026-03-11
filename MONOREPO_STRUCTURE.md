# MONOREPO_STRUCTURE.md

## Repo tree

```text
joulework-network/
├─ README.md
├─ SPEC.md
├─ MONOREPO_STRUCTURE.md
├─ CODEX_INITIAL_PROMPT.md
├─ .env.example
├─ .gitignore
├─ docker-compose.yml
├─ Makefile
├─ go.work
├─ pnpm-workspace.yaml
├─ package.json
├─ migrations/
├─ scripts/
├─ deploy/
│  ├─ docker/
│  └─ k8s/
├─ apps/
│  ├─ api/
│  ├─ web/
│  ├─ worker-node/
│  └─ admin-cli/
├─ services/
│  ├─ scheduler/
│  ├─ approvals/
│  ├─ verification/
│  ├─ reduction/
│  ├─ reliability/
│  ├─ joulework/
│  ├─ pricing/
│  ├─ ledger/
│  ├─ notifications/
│  └─ paperclip-connector/
├─ pkg/
│  ├─ domain/
│  ├─ events/
│  ├─ auth/
│  ├─ db/
│  ├─ nats/
│  ├─ redis/
│  ├─ objectstore/
│  ├─ policy/
│  ├─ telemetry/
│  ├─ workeradapter/
│  ├─ hermesadapter/
│  ├─ pricingengine/
│  ├─ settlement/
│  ├─ scoring/
│  ├─ materializers/
│  └─ apierrors/
├─ integrations/
│  ├─ paperclip/
│  └─ hermes/
├─ docs/
├─ frontend/
├─ examples/
└─ test/
```

## Dependency rules
- apps -> services/pkg/integrations
- services -> pkg/integrations
- integrations -> pkg
- pkg -> pkg only

Disallow:
- service-to-service direct imports
- frontend direct DB access
- Paperclip as source of truth
