# Web App

Svelte/SvelteKit + shadcn-svelte app scaffold for operator-facing dashboards.

## Frontend baseline

- Runtime: Node.js v24
- Package manager: pnpm
- UI framework: shadcn-svelte
- Location: `apps/web`

## shadcn-svelte setup commands

```bash
cd apps/web
pnpm dlx shadcn-svelte@latest init
pnpm dlx shadcn-svelte@latest add button
```

From repo root, use:

```bash
pnpm web:dev
pnpm web:build
```

## Current scaffold files

- `components.json`
- `src/app.css`
- `src/lib/utils.ts`
- `src/routes/+layout.svelte`
- `src/routes/+page.svelte`
