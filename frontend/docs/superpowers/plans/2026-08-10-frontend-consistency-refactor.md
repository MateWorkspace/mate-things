# Frontend Consistency Refactor Master Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Deliver the approved frontend consistency refactor through four independently reviewable phases without changing backend code, public routes, query parameter names, or intended permission behavior.

**Architecture:** Correct transport/action/query boundaries first, then build compositional filter, collection, table, and combobox primitives. Migrate routes incrementally before decomposing large runtime features and completing browser-level verification.

**Tech Stack:** Next.js 16 App Router, React 19 Server Components and Server Actions, TypeScript strict mode, Tailwind CSS v4, Vitest, Testing Library, Playwright.

**Approved design:** [Frontend Consistency Refactor Design](../specs/2026-08-10-frontend-consistency-refactor-design.md)

## Global Constraints

- Modify only files below `mate-things/frontend`.
- Ignore and preserve every worktree change outside `mate-things/frontend`.
- Preserve all public route paths and existing query parameter names.
- Preserve backend authority for validation, authorization, and device orchestration.
- Every public Server Action rechecks its exact permission and validates direct-call input.
- Raw `src/lib/api/*` wrappers are server-only transport functions, not public Server Actions.
- Action History and Node Logs remain the reference table-filter layout.
- Card collections remain cards; record-dense tables remain tables.
- Use test-first changes and keep each task independently buildable.
- Read the relevant Next.js 16 guide in `node_modules/next/dist/docs/` before changing routing, Server Actions, caching, or streaming.

## Execution Order

1. [Phase 1: Correctness and Foundations](2026-08-10-frontend-consistency-foundations.md)
2. [Phase 2: Shared Collection Systems](2026-08-10-frontend-shared-collections.md)
3. [Phase 3: Runtime Optimization and Decomposition](2026-08-10-frontend-runtime-decomposition.md)
4. [Phase 4: Verification and Hardening](2026-08-10-frontend-consistency-verification.md)

Each phase starts from the reviewed commit produced by the preceding phase.
Do not parallelize phases because later interfaces depend on earlier ones.

## Phase Gates

- **Phase 1 gate:** no raw API wrapper is a public Server Action; route policy,
  query parsing, permissions, dialog lifecycle, option loading, and firmware
  schema intent have explicit tests.
- **Phase 2 gate:** every filtered collection uses shared filter primitives;
  all async selectors use the accessible shared engine; table and card chrome
  has one implementation each.
- **Phase 3 gate:** refresh behavior is unified; local route boundaries exist;
  Firmware, Node Class, and BLE responsibilities are split without behavior
  regressions.
- **Phase 4 gate:** typecheck, lint, unit tests, production build, and scoped
  Playwright workflows pass from a clean dependency install.

## Final Verification

Run from `mate-things/frontend`:

```bash
npm ci
npm run typecheck
npm run lint
npm test
npm run build
npm run test:e2e
npm run format:check
git diff --check
```

Confirm `git status --short -- frontend` is clean from the repository root and
that no commit contains a path outside `frontend/`.
