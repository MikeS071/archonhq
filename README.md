# ArchonHQ

**AI automation that saves time and effort for busy small businesses and solopreneurs.**

Live at [archonhq.ai](https://archonhq.ai)

---

## What it offers

Practical AI automation templates and consulting services:

- **Automation Templates** — Ready-to-use systems for coaching reports, content creation, and daily operations. Cut manual work by 50-75%.
- **Automation Quick-Win Service** — Custom automation built for your specific workflow. $2,500–$5,000.

## Products

| Product | Price | Description |
|---------|-------|-------------|
| Coaching Report Automation Template | $39–$59 | Automated client reports, 75% less manual work |
| AI Content & Operations Pack | $49–$69 | Templates for emails, content, and admin tasks |

## Services

**Automation Quick-Win Service** — $2,500–$5,000

A 60-minute call to identify one repetitive task draining your time, followed by a custom automation built for your business.

[Book a call →](https://archonhq.ai/services)

## Tech stack

Next.js 16 (App Router) · TypeScript · Tailwind v4 · shadcn/ui · Drizzle ORM · PostgreSQL

## Quick start

```bash
# 1. Clone and install
git clone https://github.com/MikeS071/archonhq
cd archonhq
npm install

# 2. Configure environment
cp .env.local.example .env.local
# Fill in: DATABASE_URL, NEXTAUTH_SECRET, GOOGLE_CLIENT_ID/SECRET

# 3. Run DB migrations
npm run migrate

# 4. Start dev server
npm run dev
```

## Scripts

| Command | What it does |
|---|---|
| `npm run dev` | Local dev server |
| `npm run build` | Production build |
| `npm run migrate` | Run Drizzle schema migrations |
| `npm run lint` | Run ESLint |
| `npm test` | Run unit tests |

## Repository structure

```
src/app/              Pages (landing, products, services, insights, admin)
src/components/       UI components (shadcn/ui)
src/db/               Drizzle schema and migrations
src/lib/              Shared utilities, DB client, auth
```

## License

Apache 2.0
