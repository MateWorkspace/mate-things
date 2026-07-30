# Frontend Dashboard Integration and Hardening Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Connect the fleet-health dashboard, close cross-page gaps, verify responsive/accessibility behavior, and pass final end-to-end acceptance.

**Architecture:** The dashboard composes bounded existing endpoints in parallel and renders only permission-backed metrics. Final work centralizes route states and uses Playwright to exercise the real app shell and high-value workflows without weakening server-side authorization.

**Tech Stack:** Next.js 16, React 19, TypeScript, Tailwind CSS v4, Vitest, Testing Library, Playwright.

## Global Constraints

- Complete all preceding plans first.
- Do not add a backend aggregate endpoint in this frontend phase.
- Do not display a metric that cannot be computed honestly from bounded API
  responses.
- Final acceptance covers light/dark and mobile/tablet/desktop.

---

### Task 1: Permission-aware fleet-health dashboard

**Files:**
- Modify: `frontend/src/app/(authenticated)/dashboard/page.tsx`
- Create: `frontend/src/app/(authenticated)/dashboard/_lib/dashboard-data.ts`
- Create: `frontend/src/app/(authenticated)/dashboard/_lib/dashboard-data.test.ts`
- Create: `frontend/src/app/(authenticated)/dashboard/_components/FleetMetricCard.tsx`
- Create: `frontend/src/app/(authenticated)/dashboard/_components/UrgentAttention.tsx`
- Create: `frontend/src/app/(authenticated)/dashboard/_components/RecentWarnings.tsx`
- Create: `frontend/src/app/(authenticated)/dashboard/_components/DashboardRefresh.tsx`
- Test: `frontend/src/app/(authenticated)/dashboard/_components/UrgentAttention.test.tsx`

**Interfaces:**
- Produces:

```ts
export interface DashboardData {
  nodes?: { total: number; connected: number; disconnected: NodeResponse[] };
  failedActions?: ActionLogResponse[];
  recentWarnings?: NodeLogResponse[];
}
```

- [ ] **Step 1: Write permission/bounded-fetch tests**

```ts
it("does not request node logs without node_log:get", async () => {
  await loadDashboardData(new Set(["node:get"]));
  expect(listNodes).toHaveBeenCalled();
  expect(listNodeLogs).not.toHaveBeenCalled();
});
```

- [ ] **Step 2: Verify failure**

```bash
npm test -- src/app/'(authenticated)'/dashboard
```

- [ ] **Step 3: Implement bounded parallel composition**

Request only permitted sources. Use a small recent time window for action/node
logs and a bounded node page. Derive labels only from returned data, and label
page-scoped estimates explicitly if the endpoint cannot return a complete
aggregate.

- [ ] **Step 4: Reproduce the approved dashboard hierarchy**

Render freshness controls, health cards, urgent-attention cards, and recent
warnings. Every card links to a real filtered route. Remove empty sections and
let remaining cards reflow.

- [ ] **Step 5: Test and commit**

```bash
npm test -- src/app/'(authenticated)'/dashboard
npm run typecheck
git add frontend/src/app/'(authenticated)'/dashboard
git commit -m "feat(frontend): connect fleet health dashboard"
```

### Task 2: Shared route states and metadata

**Files:**
- Create: `frontend/src/app/(authenticated)/loading.tsx`
- Create: `frontend/src/app/(authenticated)/error.tsx`
- Create: `frontend/src/app/(authenticated)/not-found.tsx`
- Create: `frontend/src/components/layout/AccessDenied.tsx`
- Create: `frontend/src/lib/route-access.ts`
- Create: `frontend/src/lib/route-access.test.ts`
- Modify: authenticated route pages to export accurate metadata.

**Interfaces:**
- Consumes: permissions and shared state cards.
- Produces: consistent loading/error/not-found/access-denied behavior.

- [ ] **Step 1: Write route-access tests**

```ts
it("requires node:get for a node detail route", () => {
  expect(requiredPermissions("/nodes/abc")).toEqual(["node:get"]);
});

it("accepts any Access Control read permission", () => {
  expect(canVisit("/admin/access-control", new Set(["permission:get"])))
    .toBe(true);
});
```

- [ ] **Step 2: Verify failure**

```bash
npm test -- src/lib/route-access.test.ts
```

- [ ] **Step 3: Implement access and route states**

Page-level checks render `AccessDenied` before restricted reads. Errors provide
Retry, preserve safe copy, and never show backend stack details. Not-found
states link back to their collection.

- [ ] **Step 4: Add accurate metadata**

Use static titles for collections and `generateMetadata` for resolved detail
names. Do not duplicate backend reads when React cache can share the resource
lookup.

- [ ] **Step 5: Test and commit**

```bash
npm test -- src/lib/route-access.test.ts
npm run typecheck
npm run lint
git add frontend/src/app/'(authenticated)' frontend/src/components/layout/AccessDenied.tsx frontend/src/lib/route-access*
git commit -m "feat(frontend): unify protected route states"
```

### Task 3: Responsive, dark-mode, and accessibility audit

**Files:**
- Create: `frontend/e2e/accessibility.spec.ts`
- Create: `frontend/e2e/responsive-shell.spec.ts`
- Modify: affected layout/primitives/routes found by the audit.

**Interfaces:**
- Consumes: the complete UI.
- Produces: verified keyboard, responsive, light/dark behavior.

- [ ] **Step 1: Write shell viewport tests**

```ts
for (const viewport of [
  { name: "mobile", width: 390, height: 844 },
  { name: "tablet", width: 820, height: 1180 },
  { name: "desktop", width: 1440, height: 1000 },
]) {
  test(`${viewport.name} shell remains usable`, async ({ page }) => {
    await page.setViewportSize(viewport);
    await signIn(page);
    await page.goto("/nodes");
    await expect(page.getByRole("heading", { name: "Nodes" })).toBeVisible();
    await expect(page.locator("body")).not.toHaveCSS("overflow-x", "scroll");
  });
}
```

- [ ] **Step 2: Add keyboard/dialog assertions**

Test sidebar drawer focus, Escape close, focus restoration, profile tabs,
nested card actions, filter drawer, pagination, and reduced-motion emulation.
Run an automated accessibility scan on each representative page:

```ts
import AxeBuilder from "@axe-core/playwright";

const results = await new AxeBuilder({ page })
  .include("main")
  .analyze();
expect(results.violations).toEqual([]);
```

- [ ] **Step 3: Test automatic color schemes**

Run key pages with Playwright `colorScheme: "light"` and `"dark"`. Assert
critical text/status is visible and capture screenshots for dashboard, nodes,
node detail, Access Control, and profile dialog at the approved widths.

- [ ] **Step 4: Fix every reproducible issue**

Make focused semantic/token/layout changes, rerunning the failing test after
each change. Do not suppress accessibility findings with blanket exclusions.

- [ ] **Step 5: Commit**

```bash
npm run test:e2e -- e2e/accessibility.spec.ts e2e/responsive-shell.spec.ts
git add frontend/e2e frontend/src
git commit -m "fix(frontend): harden responsive accessibility"
```

### Task 4: End-to-end workflows and final verification

**Files:**
- Create: `frontend/e2e/helpers/session.ts`
- Create: `frontend/e2e/profile.spec.ts`
- Create: `frontend/e2e/nodes.spec.ts`
- Create: `frontend/e2e/operations.spec.ts`
- Create: `frontend/e2e/administration.spec.ts`
- Modify: defects discovered by end-to-end tests.

**Interfaces:**
- Consumes: complete frontend and a seeded backend test environment.
- Produces: acceptance evidence for the approved design.

- [ ] **Step 1: Add deterministic session helpers**

Provide helpers that log in through the UI and expose seeded super/admin/user
credentials only through test environment variables. Never commit credentials.

- [ ] **Step 2: Implement required journeys**

Cover:

```ts
test("profile edit, password change, and logout", profileJourney);
test("filter nodes and update a node", nodeJourney);
test("dispatch an action and inspect its result", actionJourney);
test("select firmware and confirm OTA", otaJourney);
test("filter telemetry and node logs", observabilityJourney);
test("manage a user and role permissions", administrationJourney);
test("create and edit a payload schema version", schemaJourney);
```

Each journey asserts both visible success and the refreshed authoritative data.

- [ ] **Step 3: Run full verification**

```bash
npm test
npm run typecheck
npm run lint
npm run test:e2e
npm run build
```

Expected: every command exits 0 with no skipped critical journey.

- [ ] **Step 4: Inspect production output**

Start `npm run start` from the production build and smoke-test login,
dashboard, nodes, profile, and one admin route through the Go backend proxy
topology. Confirm no API hostname or token is present in client bundles.

- [ ] **Step 5: Commit final fixes**

```bash
git add frontend
git commit -m "test(frontend): complete operations cockpit acceptance"
```

- [ ] **Step 6: Request final code review**

Use `superpowers:requesting-code-review`, address verified findings, rerun the
full verification block, then use `superpowers:finishing-a-development-branch`
to offer merge/integration choices.
