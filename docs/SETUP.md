---
title: "Setup Guide"
---

# Setup Guide

This guide gets ArchonHQ running locally.

## Prerequisites

- Node.js 20+
- npm 10+
- PostgreSQL 16+
- A Google OAuth app (for admin login)

## 1. Install Dependencies

```bash
npm install
```

## 2. Configure Environment

Create `.env.local` in repo root:

```bash
DATABASE_URL=REPLACE_WITH_YOUR_DATABASE_URL
NEXTAUTH_SECRET=replace_with_a_long_random_secret
AUTH_SECRET=same_as_nexthauth_secret
GOOGLE_CLIENT_ID=your_google_client_id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your_google_client_secret
NEXTAUTH_URL=http://localhost:3000
```

For products, add Whop product IDs:

```bash
WHOP_COACHING_TEMPLATE_ID=your_whop_product_id
WHOP_CONTENT_PACK_ID=your_whop_product_id
```

## 3. Start PostgreSQL

```sql
CREATE USER archon_user WITH PASSWORD 'archon_pass';
CREATE DATABASE archonhq OWNER archon_user;
```

## 4. Run Migrations

```bash
npm run migrate
```

## 5. Start the App

```bash
npm run dev
```

Open: `http://localhost:3000`

## Common Setup Errors

- `DATABASE_URL` missing → DB connection errors
- `NEXTAUTH_SECRET` missing → auth may break
- `GOOGLE_CLIENT_ID/SECRET` missing → cannot sign in
