# Contributing to ArchonHQ

Thanks for your interest in ArchonHQ — an open coordination protocol for distributed AI agents!

This project is still in early alpha / specification stage. Contributions are very welcome, especially around:
- Feedback on architecture decisions (ADRs)
- Suggestions for missing specs (e.g., agent attestation, scaling concerns)
- Code for stubs / adapters (Hermes integration, Paperclip workflows)
- Tests, docs, diagrams, or examples
- Bug reports on any runnable parts

## How to Contribute

1. **Discuss first** — Open an issue for big ideas/changes. This helps align on direction early (especially since much is still spec-level).

2. **Small changes** — Feel free to open a PR directly (typos, docs tweaks, small code improvements).

3. **Development flow** (current solo style):
   - Use Codex / similar LLMs to generate milestone code/docs from prompts.
   - Commit specs → stubs → tests incrementally.
   - Update ADRs when decisions change.
   - Keep monorepo structure: hub (Go), frontend (Svelte + shadcn-svelte), docs/specs.

4. **Commit style** — Conventional-ish: `feat: add X`, `docs: update README`, `fix: mermaid parse error`, etc. (not strict).

5. **Code style** — Follow Go standard (gofmt), SvelteKit + shadcn-svelte conventions, and Node.js v24 runtime compatibility. Run linters if/when added.

## Pull Requests

- Target `main` branch.
- Include clear description: what + why + any trade-offs.
- Link related issues if applicable.

## Issues

Use the templates in `.github/ISSUE_TEMPLATE/` for bugs, features, or questions.

Questions? Just open an issue — happy to chat!

License: MIT (see LICENSE)
