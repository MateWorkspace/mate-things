# Frontend Consistency Verification and Hardening Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Close the refactor with contract, route, accessibility, and browser coverage; verify the frontend from a clean dependency install; and document only the durable conventions introduced by the work.

**Architecture:** Fast unit/contract tests protect boundaries and shared primitives, route tests protect permission-aware Server Component composition, and a small Playwright suite protects critical user journeys. Verification is performed from the frontend directory and the final diff is scope-audited before integration.

**Tech Stack:** Vitest, Testing Library, MSW or the project's existing HTTP mock mechanism, Playwright, axe accessibility checks, Next.js production build, TypeScript, ESLint, Prettier.

## Global Constraints

- Start from the reviewed Phase 3 commit and work only below `mate-things/frontend`.
- Do not weaken assertions merely to make the suite pass; fix the behavior or correct an obsolete fixture against the current typed API contract.
- Do not require a live production backend for deterministic CI tests. Use the existing test harness or a frontend-local mock server.
- Keep browser tests small and scenario-driven; exhaustive input matrices belong in Vitest.
- Never commit generated secrets, session tokens, `.env` values, Playwright traces, screenshots, videos, or dependency directories.

---

## Task 1: Restore a clean and reproducible test baseline

**Files:**

- Modify if required: `package.json`
- Modify if required: `package-lock.json`
- Modify: `src/app/(authenticated)/node-classes/[id]/page.test.tsx`
- Modify if required: `vitest.config.mts`
- Modify if required: `src/test/fixtures.ts`
- Modify if required: `src/test/form-data.ts`

**Verification shape:**

```bash
npm ci
npm run typecheck
```

Expected: dependency installation is driven solely by `package-lock.json`, and typecheck resolves Playwright, Vitest, Testing Library, and current API fixture types without a parent/global module tree.

- [ ] Record `node --version` and `npm --version`; compare with `package.json` engines and repository runtime documentation.
- [ ] Run `npm ci` from `mate-things/frontend`. If installation fails, diagnose the lockfile/runtime mismatch before changing dependencies.
- [ ] Run `npm run typecheck` and capture every baseline failure. Correct the stale Node Class fixture so it matches the current `ActionResponse` type instead of using removed `node_class_id`.
- [ ] Resolve missing declared test dependencies through `devDependencies` and the lockfile only when a clean install proves they are absent. Do not depend on a parent/global `node_modules` symlink.
- [ ] Run `npm run typecheck`, `npm run lint`, and `npm test` until the baseline and all Phase 1–3 tests execute from the clean install.
- [ ] Commit only actual reproducibility/fixture changes: `git add package.json package-lock.json src vitest.config.* && git commit -m "test(frontend): restore reproducible test baseline"`.

## Task 2: Complete boundary and shared-infrastructure contract coverage

**Files:**

- Create: `src/lib/api/client.test.ts`
- Create: `src/lib/session/session.test.ts`
- Modify: `src/proxy.test.ts`
- Modify: `src/config/route-policies.test.ts`
- Modify: `src/lib/query.test.ts`
- Modify: `src/lib/api/collect-all-pages.test.ts`
- Modify: `src/lib/actions/entity-search-actions.test.ts`
- Modify: `src/components/combobox/async-entity-combobox.test.tsx`
- Modify: `src/components/filters/filter-route-contracts.test.tsx`

**Test shape:**

```ts
it("forwards JSON errors without leaking authorization", async () => {
  mockFetch.mockResolvedValue(jsonResponse(403, { message: "Forbidden" }));
  await expect(apiRequest("/nodes", { token: "secret" })).rejects.toMatchObject(
    { status: 403 },
  );
  expect(mockLogger).not.toHaveBeenCalledWith(
    expect.stringContaining("secret"),
  );
});

it.each(NAVIGATION_HREFS)("has exactly one policy for %s", (href) => {
  expect(ROUTE_POLICIES.filter((policy) => policy.href === href)).toHaveLength(
    1,
  );
});
```

- [ ] Add API-client tests for base URL, cookie forwarding, JSON/non-JSON bodies, empty success responses, structured API errors, network errors, and authentication refresh behavior. Never log authorization headers or cookies.
- [ ] Add session/proxy matrices covering valid, expired-refreshable, expired-unrefreshable, malformed, and missing sessions across public/protected/nested routes.
- [ ] Ensure route policy tests prove navigation, proxy, and authorization helpers cannot drift: every navigable href has exactly one policy and every protected feature prefix has an expected policy.
- [ ] Complete parser/all-page tests for malicious arrays, Unicode/whitespace, `Number.MAX_SAFE_INTEGER`, duplicate/non-progress pages, zero limits, and safety ceiling.
- [ ] Complete selector action/combobox tests for all six adapters, permission suppression, stale requests, keyboard behavior, loading/error/empty pages, and selected-value submission.
- [ ] Complete the route filter contract matrix for every card/table collection and exact existing query keys.
- [ ] Run only these tests first, then `npm test`.
- [ ] Commit: `git add src/lib src/proxy.test.ts src/config src/components && git commit -m "test(frontend): cover shared boundary contracts"`.

## Task 3: Add permission-aware route composition coverage

**Files:**

- Create: `src/app/(authenticated)/dashboard/page.permissions.test.tsx`
- Create: `src/app/(authenticated)/nodes/page.permissions.test.tsx`
- Modify: `src/app/(authenticated)/nodes/[id]/page.test.tsx`
- Modify: `src/app/(authenticated)/node-classes/page.test.tsx`
- Modify: `src/app/(authenticated)/node-classes/[id]/page.test.tsx`
- Create: `src/app/(authenticated)/firmware/page.permissions.test.tsx`
- Create: `src/app/(authenticated)/firmware/[id]/page.permissions.test.tsx`
- Modify: `src/app/(authenticated)/actions/page.test.tsx`
- Create: `src/app/(authenticated)/actions/[id]/page.permissions.test.tsx`
- Modify: `src/app/(authenticated)/action-history/page.test.tsx`
- Modify: `src/app/(authenticated)/telemetry/page.test.tsx`
- Modify: `src/app/(authenticated)/node-logs/page.test.tsx`
- Create: `src/app/(authenticated)/broadcast-sessions/page.permissions.test.tsx`
- Create: `src/app/(authenticated)/ble-direct/page.permissions.test.tsx`
- Create: `src/app/(authenticated)/admin/users/page.permissions.test.tsx`
- Create: `src/app/(authenticated)/admin/users/[id]/page.permissions.test.tsx`
- Create: `src/app/(authenticated)/admin/access-control/page.permissions.test.tsx`
- Create: `src/app/(authenticated)/admin/payload-schemas/page.permissions.test.tsx`
- Create: `src/app/(authenticated)/admin/payload-schemas/[id]/page.permissions.test.tsx`
- Create: `src/app/(authenticated)/admin/api-keys/page.permissions.test.tsx`
- Create: `src/test/permission-scenarios.ts`
- Create: `src/test/route-render.tsx`

**Required route matrix:**

| Area           | Routes                                                                                                                                   |
| -------------- | ---------------------------------------------------------------------------------------------------------------------------------------- |
| Fleet          | `/dashboard`, `/nodes`, `/nodes/[id]`, `/node-classes`, `/node-classes/[id]`, `/firmware`, `/firmware/[id]`                              |
| Operations     | `/actions`, `/actions/[id]`, `/action-history`, `/telemetry`, `/node-logs`, `/broadcast-sessions`, `/ble-direct`                         |
| Administration | `/admin/users`, `/admin/users/[id]`, `/admin/access-control`, `/admin/payload-schemas`, `/admin/payload-schemas/[id]`, `/admin/api-keys` |

**Test shape:**

```ts
it.each([
  {
    route: NodeLogsPage,
    routePermission: "node_log:get",
    ancillaryPermission: "node:get",
    lookup: getNodeById,
  },
  {
    route: ActionHistoryPage,
    routePermission: "action_log:get",
    ancillaryPermission: "action:get",
    lookup: getActionById,
  },
])(
  "suppresses forbidden ancillary lookup",
  async ({ route, routePermission, lookup }) => {
    mockPermissionScenario({ allow: [routePermission] });
    await renderServerPage(route, { node_id: "raw-id", action_id: "raw-id" });
    expect(lookup).not.toHaveBeenCalled();
  },
);
```

- [ ] Build reusable permission scenarios for unauthenticated, route-denied, collection-read-only, mutation-capable, remove-only, and ancillary-read-denied users. Keep scenario helpers data-only.
- [ ] For each route, test authorized data fetch, denied-fetch suppression, visible/hidden actions, loading, 404 where applicable, expected empty state, unexpected transport error propagation, and normalized out-of-range pagination.
- [ ] Explicitly cover record routes where `action_log:get`, `telemetry:get`, or `node_log:get` is present but `action:get`/`node:get` is absent; selectors and label calls must remain suppressed.
- [ ] Explicitly cover remove-only destructive dialogs without introducing a frontend read requirement, while still asserting the backend-authoritative confirmation payload is sent.
- [ ] Run route tests in bounded area groups, then run the whole suite.
- [ ] Commit: `git add src/app/'(authenticated)' src/test && git commit -m "test(frontend): cover permission aware routes"`.

## Task 4: Complete workflow regression coverage

**Files:**

- Create: `src/app/(authenticated)/actions/_lib/actions.contract.test.ts`
- Create: `src/app/(authenticated)/admin/access-control/_lib/actions.contract.test.ts`
- Create: `src/app/(authenticated)/admin/api-keys/_lib/actions.contract.test.ts`
- Modify: `src/app/(authenticated)/admin/api-keys/_components/api-key-dialogs.test.tsx`
- Modify: `src/app/(authenticated)/admin/payload-schemas/_lib/actions.test.ts`
- Create: `src/app/(authenticated)/admin/users/_lib/actions.contract.test.ts`
- Modify: `src/app/(authenticated)/firmware/_lib/actions.test.ts`
- Modify: `src/app/(authenticated)/firmware/_components/FirmwareForm.test.tsx`
- Modify: `src/app/(authenticated)/node-classes/_lib/actions.test.ts`
- Modify: `src/app/(authenticated)/nodes/[id]/_lib/actions.test.ts`
- Modify: `src/app/(authenticated)/nodes/[id]/_components/DeleteNodeDialog.test.tsx`
- Modify: `src/components/ui/dialog.test.tsx`

**Test shape:**

```ts
it.each(PUBLIC_ACTION_CASES)(
  "rejects malformed direct calls: $name",
  async ({ action, invalid, transport }) => {
    await expect(action(INITIAL_ACTION_STATE, invalid)).resolves.toMatchObject({
      status: "error",
    });
    expect(transport).not.toHaveBeenCalled();
  },
);
```

- [ ] Add a table-driven Server Action test harness that invokes every public action with malformed direct input and without permission. Assert transport is never called in both cases.
- [ ] Cover create/edit/delete/reset/revoke/regenerate/assign/dispatch/OTA flows for pending, field error, server error, success, reset, repeat use, and revalidation/redirect behavior.
- [ ] Preserve special contracts: firmware Keep/Replace/Clear multipart intent, generated API-key show-once secret, immutable node device ID, explicit payload-schema update discriminant, and partial assignment retry.
- [ ] Cover all destructive dialogs for typed confirmation, pending non-dismissal, focus restoration, cancel, server rejection, and successful close/reset. Assert only authoritative backend confirmation claims in UI copy/tests.
- [ ] Run workflow tests, typecheck, lint, and the entire test suite.
- [ ] Commit: `git add src/app src/components/ui/dialog.test.tsx && git commit -m "test(frontend): cover mutation workflows"`.

## Task 5: Add focused browser journeys

**Files:**

- Create: `e2e/support/auth-fixtures.ts`
- Modify: `e2e/support/mock-backend.mjs`
- Create: `e2e/collection-filters.spec.ts`
- Create: `e2e/selector-keyboard.spec.ts`
- Create: `e2e/admin-workflows.spec.ts`
- Create: `e2e/firmware-schema-intent.spec.ts`
- Create: `e2e/refresh.spec.ts`
- Create: `e2e/ble-unsupported.spec.ts`
- Modify: existing authentication/accessibility specs under `e2e/`
- Modify if required: `playwright.config.ts`

**Fixture shape:**

```ts
export const test = base.extend<{ permissions: string[] }>({
  permissions: [[], { option: true }],
  page: async ({ page, permissions }, use) => {
    await mockAuthenticatedBackend(page, { permissions });
    await use(page);
  },
});
```

**Journey shape:**

```ts
test("filters node logs without changing query names", async ({ page }) => {
  await page.goto("/node-logs");
  await page.getByLabel("Message").fill("timeout");
  await page.getByRole("button", { name: "Apply filters" }).click();
  await expect(page).toHaveURL(/search=timeout/);
  await expect(page).toHaveURL(/page=1/);
});
```

- [ ] Create deterministic frontend-local fixtures for authenticated permissions and API responses. Mask tokens and secrets in logs/traces.
- [ ] Test the canonical filter journey on Action History, Node Logs, and Nodes: open mobile drawer, fill fields, apply, observe exact URL keys, paginate, edit filter (page resets), clear, and preserve browser Back/Forward behavior.
- [ ] Test all keyboard combobox behaviors through one representative route, plus adapter smoke checks for action, node, node class, role, and user selectors.
- [ ] Test an administrative repeat-use dialog and API-key generation: pending dismissal blocked, secret copied/closed, and secret absent when reopened.
- [ ] Test firmware replacement Keep/Replace/Clear by inspecting the intercepted multipart request.
- [ ] Test smart refresh pause/manual/visibility behavior with controlled clock/server timestamps.
- [ ] Test BLE unsupported-browser messaging without requiring physical hardware. Keep real-device BLE out of CI.
- [ ] Run each spec alone, then `npm run test:e2e`. Retain traces only on failure and remove local artifacts before commit.
- [ ] Commit: `git add e2e playwright.config.ts && git commit -m "test(frontend): add critical browser journeys"`.

## Task 6: Accessibility, performance, and responsive verification

**Files:**

- Modify: `e2e/authenticated-accessibility.spec.ts`
- Modify only when a check reveals a defect: `src/components/filters/FilterPanel.tsx`
- Modify only when a check reveals a defect: `src/components/combobox/AsyncEntityCombobox.tsx`
- Modify only when a check reveals a defect: `src/components/table/TableFrame.tsx`
- Modify only when a check reveals a defect: `src/components/table/ExpandableTableRow.tsx`
- Modify only when a check reveals a defect: `src/components/collection/ResourceCardLink.tsx`
- Modify only when a check reveals a defect: `src/components/ui/dialog.tsx`
- Modify only when a check reveals a defect: `src/components/layout/AppShell.tsx`
- Modify only when a check reveals a defect: `src/app/(authenticated)/ble-direct/_components/BleDirectRoot.tsx`

**Measurement shape:**

```ts
const renderCounts = renderBleWithCounters();
emitBleLog("sample");
expect(renderCounts.current()).toEqual({
  log: 2,
  systemInfo: 1,
  wifi: 1,
  settings: 1,
});
```

Browser accessibility checks must use the existing axe helper/spec convention; violations are asserted as an empty array after excluding no rules unless the exclusion is documented in the test with an upstream issue.

- [ ] Run automated accessibility scans for the shell, one card collection, Action History, Node Logs, one detail route, each dialog state, and the combobox open/loading/error states.
- [ ] Keyboard-walk skip link, app bar/profile dialog, sidebar, filter drawer, selectors, cards, tables, pagination, dialogs, and tabs. Record and fix focus loss, traps, hidden focus, or illogical order.
- [ ] At 320px, 375px, 768px, 1024px, and 1440px, inspect overflow and touch targets. Tables may scroll inside their shell but must not force page-wide horizontal scrolling.
- [ ] Use React Profiler or a render-counter test to confirm BLE log events do not rerender unrelated panels and filter typing does not rerender an entire Server Component page client-side.
- [ ] Inspect production build output for accidental client-boundary expansion. Shared server-compatible filter/card/table layout modules must not gain `"use client"` unless they directly need state/browser APIs.
- [ ] Re-run focused checks after each defect fix, then the full unit and browser suites.
- [ ] Commit any fixes and tests together: `git add src e2e && git commit -m "fix(frontend): complete accessibility hardening"`.

## Task 7: Update durable frontend guidance and perform final scope audit

**Files:**

- Modify: `AGENTS.md`
- Modify: `docs/superpowers/specs/2026-08-10-frontend-consistency-refactor-design.md` only if implementation intentionally changed the approved design
- Modify: these four phase plans only to mark completed checkboxes or record approved deviations

**Documentation shape:**

```md
- Raw API modules under `src/lib/api` import `server-only`; Client Components call permission-checked selector actions from `src/lib/actions`.
- Collection filters compose `src/components/filters`; preserve established query keys when adding fields.
- Use `FirmwareSchemaIntent` Keep/Replace/Clear so omitted and empty schemas remain distinct.
```

- [ ] Update `frontend/AGENTS.md` only with durable rules established by the implementation: server-only transport/public action separation, route policy source, shared filter/table/card/combobox locations, query parsing, dialog lifecycle, firmware schema intent, and verification commands.
- [ ] Do not document proposed components that were not actually shipped. Keep route inventory and examples synchronized with code.
- [ ] Run `rg -n 'TODO|FIXME|XXX|\.only\(|\.skip\(' src e2e` and resolve any refactor leftovers or explicitly justified pre-existing skips.
- [ ] From `mate-things/frontend`, run:

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

- [ ] Before Phase 1, record the reviewed starting commit in `FRONTEND_REFACTOR_BASE=$(git rev-parse HEAD)`. From `mate-things`, run `git status --short -- frontend` and `git diff --name-only "$FRONTEND_REFACTOR_BASE"..HEAD`; verify every refactor path starts with `frontend/`.
- [ ] Inspect `git diff --stat "$FRONTEND_REFACTOR_BASE"..HEAD`, public route filenames, query-contract tests, and production build routes. Confirm no backend file was staged or committed.
- [ ] Commit documentation only after all checks pass: `git add frontend/AGENTS.md frontend/docs && git commit -m "docs(frontend): document shared frontend architecture"`.

## Final Integration Gate

- [ ] Use `superpowers:requesting-code-review` for a correctness, security-boundary, accessibility, and maintainability review of the complete frontend range.
- [ ] Address review findings with `superpowers:receiving-code-review`, rerun the affected focused checks, then rerun the full command block above.
- [ ] Use `superpowers:verification-before-completion`; report exact commands and fresh output rather than claiming inferred success.
- [ ] Use `superpowers:finishing-a-development-branch` to choose integration. Because the user requested frontend-only scope, reject any integration range containing paths outside `frontend/`.
