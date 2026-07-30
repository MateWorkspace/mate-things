# Mate Things Frontend Operations Cockpit Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Deliver the complete approved Mate Things operations cockpit across five working, reviewable phases.

**Architecture:** A protected Next.js App Router shell supplies profile and effective permissions to Server Components and small interactive Client Components. Typed server-only API wrappers remain the only backend access layer; URL search parameters drive collections, and reusable card/filter/refresh primitives keep resource pages consistent.

**Tech Stack:** Next.js 16.2.11 App Router, React 19.2.4, TypeScript strict mode, Tailwind CSS v4, Lucide React, Vitest, Testing Library, and Playwright.

## Global Constraints

- Read the relevant local Next.js 16 guide under `frontend/node_modules/next/dist/docs/` before implementing the corresponding feature.
- Keep JWT access and refresh tokens in `httpOnly` cookies; never expose tokens to Client Components or browser storage.
- Backend rules remain authoritative; the frontend owns presentation and thin typed data access only.
- Components never call backend `fetch` directly; use one typed file per backend resource under `src/lib/api/`.
- Server Components are the default; push `"use client"` to the smallest interactive leaf.
- Use semantic Tailwind v4 theme tokens in both automatic light and dark modes; do not hardcode component colors.
- Use card-based collections. Nodes must never use a table and must not expose node creation.
- Navigation and every mutation control are gated by exact effective permissions.
- Build mobile-first and verify mobile, tablet, and desktop.
- Every task follows red-green-refactor and ends in a focused commit.
- `npm run lint`, `npm run typecheck`, `npm test`, and `npm run build` must pass before a phase is complete.

---

## Execution order

1. [Foundation, authenticated shell, and account](./2026-07-30-frontend-foundation-shell-account.md)
2. [Fleet management](./2026-07-30-frontend-fleet-management.md)
3. [Operations and observability](./2026-07-30-frontend-operations-observability.md)
4. [Administration](./2026-07-30-frontend-administration.md)
5. [Dashboard integration and hardening](./2026-07-30-frontend-dashboard-hardening.md)

Each phase starts from the committed result of the previous phase. Do not begin
a later phase with failing verification from an earlier phase.

## Cross-phase public interfaces

Foundation produces:

```ts
export type PermissionName = string;

export interface SessionContext {
  user: UserResponse;
  permissions: ReadonlySet<PermissionName>;
}

export function hasPermission(
  permissions: ReadonlySet<string>,
  permission: string,
): boolean;

export function canAccessAny(
  permissions: ReadonlySet<string>,
  required: readonly string[],
): boolean;

export interface FormActionState {
  status: "idle" | "success" | "error";
  title?: string;
  message?: string;
  fieldErrors?: Record<string, string>;
}
```

Collection infrastructure produces:

```ts
export interface CollectionPage {
  page: number;
  limit: number;
  search?: string;
}

export interface RefreshState {
  status: "live" | "paused" | "stale";
  updatedAt: Date;
  refresh(): void;
  pause(): void;
  resume(): void;
}
```

All later phases consume those exact contracts rather than redefining
permission, pagination, or refresh behavior.

## Final acceptance

- All approved routes are reachable and protected.
- Every backend resource and permission has an intentional UI location.
- The five approved Superdesign targets are reproduced as one coherent system.
- All collections have loading, empty, no-results, error, stale, read-only, and
  mutation-pending states.
- Full verification and the approved end-to-end journeys pass.
