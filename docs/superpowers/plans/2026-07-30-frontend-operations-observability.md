# Frontend Operations and Observability Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Deliver action definition/dispatch/history, telemetry, node logs, and controlled smart refresh.

**Architecture:** Global pages and node tabs share typed record cards and URL filters. A client refresh boundary periodically calls a same-route refresh without owning backend credentials; initial and authoritative data remain server-rendered.

**Tech Stack:** Next.js 16, React 19, TypeScript, Tailwind CSS v4, Vitest, Testing Library, Playwright.

## Global Constraints

- Complete foundation and fleet plans first.
- Do not invent backend pagination for counted action-log, telemetry, or
  node-log endpoints.
- Filter-scoped deletion must preview filters and require confirmation.
- Refresh pauses for hidden tabs, forms, and modal confirmation.

---

### Task 1: Node-log API and shared record filters

**Files:**
- Create: `frontend/src/lib/api/node-logs.ts`
- Modify: `frontend/src/lib/api/index.ts`
- Create: `frontend/src/lib/api/node-logs.test.ts`
- Create: `frontend/src/lib/record-filters.ts`
- Create: `frontend/src/lib/record-filters.test.ts`
- Create: `frontend/src/components/records/TimeRangeFilter.tsx`
- Create: `frontend/src/components/records/JsonPayload.tsx`
- Create: `frontend/src/components/records/RecordWindow.tsx`
- Create: `frontend/src/components/records/RecordWindow.test.tsx`

**Interfaces:**
- Produces:

```ts
export type NodeLogLevel = "NONE" | "ERROR" | "WARN" | "INFO" | "DEBUG";

export interface NodeLogResponse {
  id: number;
  node_device_id: string;
  level: NodeLogLevel;
  tag: string;
  message: string;
  logged_at: string;
  created_at: string;
}
```

- [ ] **Step 1: Write failing API and filter tests**

```ts
it("builds the supported node-log filter query", async () => {
  await listNodeLogs({ node_device_id: "node-42", level: "WARN" });
  expect(apiFetch).toHaveBeenCalledWith(
    "/node-logs?node_device_id=node-42&level=WARN",
  );
});

it("rejects an inverted time range", () => {
  expect(parseRecordFilters({
    start: "2026-07-30T12:00:00Z",
    end: "2026-07-30T11:00:00Z",
  }).error).toMatch(/after/i);
});
```

- [ ] **Step 2: Verify failure**

```bash
npm test -- src/lib/api/node-logs.test.ts src/lib/record-filters.test.ts
```

- [ ] **Step 3: Implement exact backend wrappers**

Mirror node log list/delete with `CountDataResponse` and `CountResponse`.
`parseRecordFilters` accepts ISO datetimes, trims IDs/metric/schema/tag values,
and returns a field error when start follows end.

- [ ] **Step 4: Implement reusable record controls**

`TimeRangeFilter` uses `datetime-local` inputs and serializes UTC ISO strings.
`JsonPayload` renders formatted JSON in a disclosure with an accessible Copy
button. `RecordWindow` renders the first 100 records and an accessible “Show
100 more” button until all currently returned records are visible, preventing
an unbounded initial DOM without pretending the backend paginates.

- [ ] **Step 5: Test and commit**

```bash
npm test -- src/lib/api/node-logs.test.ts src/lib/record-filters.test.ts src/components/records
npm run typecheck
git add frontend/src/lib/api frontend/src/lib/record-filters* frontend/src/components/records
git commit -m "feat(frontend): add node log and record contracts"
```

### Task 2: Controlled smart refresh

**Files:**
- Create: `frontend/src/hooks/use-smart-refresh.ts`
- Create: `frontend/src/hooks/use-smart-refresh.test.ts`
- Create: `frontend/src/components/refresh/RefreshControl.tsx`
- Create: `frontend/src/components/refresh/RefreshBoundary.tsx`
- Create: `frontend/src/components/refresh/RefreshControl.test.tsx`

**Interfaces:**
- Produces `RefreshState` from the suite index.
- `RefreshBoundary` consumes `intervalMs`, `updatedAt`, `suspended`, and
  children; refresh uses `router.refresh()`.

- [ ] **Step 1: Write timer/visibility tests**

```ts
it("pauses while the document is hidden and resumes when visible", () => {
  const { result } = renderHook(() => useSmartRefresh({ intervalMs: 30_000 }));
  setDocumentVisibility("hidden");
  expect(result.current.status).toBe("paused");
  setDocumentVisibility("visible");
  expect(result.current.status).toBe("live");
});
```

- [ ] **Step 2: Verify failure**

```bash
npm test -- src/hooks/use-smart-refresh.test.ts
```

- [ ] **Step 3: Implement non-overlapping refresh state**

Use one timeout scheduled after the prior refresh completes, listen for
`visibilitychange`, cancel on unmount, and set `stale` when the refresh callback
rejects while keeping the previous rendered data.

- [ ] **Step 4: Implement accessible controls**

Render text status, relative updated time, Pause/Resume, and Refresh. Announce
state changes through a polite live region. Respect `suspended` while a form or
dialog is active.

- [ ] **Step 5: Test and commit**

```bash
npm test -- src/hooks/use-smart-refresh.test.ts src/components/refresh
npm run typecheck
git add frontend/src/hooks/use-smart-refresh* frontend/src/components/refresh
git commit -m "feat(frontend): add controlled smart refresh"
```

### Task 3: Actions and dispatch

**Files:**
- Create: `frontend/src/app/(authenticated)/actions/page.tsx`
- Create: `frontend/src/app/(authenticated)/actions/[id]/page.tsx`
- Create: `frontend/src/app/(authenticated)/actions/_components/ActionCard.tsx`
- Create: `frontend/src/app/(authenticated)/actions/_components/ActionForm.tsx`
- Create: `frontend/src/app/(authenticated)/actions/_components/DispatchActionDialog.tsx`
- Create: `frontend/src/app/(authenticated)/actions/_lib/actions.ts`
- Test: `frontend/src/app/(authenticated)/actions/_lib/actions.test.ts`
- Test: `frontend/src/app/(authenticated)/actions/_components/DispatchActionDialog.test.tsx`

**Interfaces:**
- Consumes: action, node, node-class, payload-schema, and action-log wrappers.
- Produces: action CRUD and node-compatible dispatch.

- [ ] **Step 1: Write permission and payload tests**

```ts
it("requires action:dispatch before calling the backend", async () => {
  mockPermissions(["action:get"]);
  const result = await dispatchActionFormAction(
    EMPTY_STATE,
    formData({ action_id: "a1", node_id: "n1", payload: "{}" }),
  );
  expect(result.status).toBe("error");
  expect(dispatchAction).not.toHaveBeenCalled();
});
```

- [ ] **Step 2: Verify failure**

```bash
npm test -- src/app/'(authenticated)'/actions
```

- [ ] **Step 3: Implement card CRUD**

Cards show name, description, node class, schema/version, and permitted actions.
Create/edit forms select valid class and schema/version. Delete uses resource
identity confirmation.

- [ ] **Step 4: Implement dispatch**

Restrict node selection to the action’s node class. Parse payload with
`JSON.parse`, require an object, serialize optional `executed_at` to ISO, call
`dispatchAction`, then link to `/action-history?execution_id=<id>` using the
returned log.

- [ ] **Step 5: Test and commit**

```bash
npm test -- src/app/'(authenticated)'/actions
npm run typecheck
npm run lint
git add frontend/src/app/'(authenticated)'/actions
git commit -m "feat(frontend): add action management and dispatch"
```

### Task 4: Action History and Telemetry

**Files:**
- Create: `frontend/src/app/(authenticated)/action-history/page.tsx`
- Create: `frontend/src/app/(authenticated)/action-history/_components/ActionLogCard.tsx`
- Create: `frontend/src/app/(authenticated)/action-history/_components/ActionLogFilters.tsx`
- Create: `frontend/src/app/(authenticated)/action-history/_lib/actions.ts`
- Create: `frontend/src/app/(authenticated)/telemetry/page.tsx`
- Create: `frontend/src/app/(authenticated)/telemetry/_components/TelemetryCard.tsx`
- Create: `frontend/src/app/(authenticated)/telemetry/_components/TelemetryFilters.tsx`
- Create: `frontend/src/app/(authenticated)/telemetry/_lib/actions.ts`
- Test: `frontend/src/app/(authenticated)/action-history/_components/ActionLogCard.test.tsx`
- Test: `frontend/src/app/(authenticated)/telemetry/_components/TelemetryCard.test.tsx`

**Interfaces:**
- Consumes: counted endpoints, `TimeRangeFilter`, `JsonPayload`,
  `RefreshBoundary`.
- Produces: global filtered record views and filter-scoped delete actions.

- [ ] **Step 1: Write record-card tests**

```tsx
it.each(["UNEXECUTED", "UNRESPONDED", "FAILED", "SUCCESS"] as const)(
  "shows the %s action status as text",
  (status) => {
    render(<ActionLogCard log={actionLog({ action_status: status })} />);
    expect(screen.getByText(statusLabel(status))).toBeVisible();
  },
);
```

- [ ] **Step 2: Verify failure**

```bash
npm test -- src/app/'(authenticated)'/action-history src/app/'(authenticated)'/telemetry
```

- [ ] **Step 3: Implement filtered record pages**

Action History uses supported execution-time/action/node filters. Telemetry uses
recorded-time/device/metric/schema filters. Each page shows total count,
bounded-time guidance, expandable payloads, and smart refresh.

- [ ] **Step 4: Implement scoped deletion**

Server Actions reconstruct filters from submitted hidden values, check
`action_log:remove` or `telemetry_record:remove`, call delete, and return the
backend count. The confirmation dialog lists every active filter and never
submits an unfiltered delete without a second typed confirmation.

- [ ] **Step 5: Test and commit**

```bash
npm test -- src/app/'(authenticated)'/action-history src/app/'(authenticated)'/telemetry
npm run typecheck
npm run lint
git add frontend/src/app/'(authenticated)'/action-history frontend/src/app/'(authenticated)'/telemetry
git commit -m "feat(frontend): add action history and telemetry"
```

### Task 5: Node Logs and node-tab integration

**Files:**
- Create: `frontend/src/app/(authenticated)/node-logs/page.tsx`
- Create: `frontend/src/app/(authenticated)/node-logs/_components/NodeLogCard.tsx`
- Create: `frontend/src/app/(authenticated)/node-logs/_components/NodeLogFilters.tsx`
- Create: `frontend/src/app/(authenticated)/node-logs/_lib/actions.ts`
- Test: `frontend/src/app/(authenticated)/node-logs/_components/NodeLogCard.test.tsx`
- Modify: `frontend/src/app/(authenticated)/nodes/[id]/page.tsx`

**Interfaces:**
- Consumes: node-log wrapper, record filters, refresh, node device ID.
- Produces: global and node-filtered logs.

- [ ] **Step 1: Write level and message tests**

```tsx
it("renders level, tag, device, and expandable message", async () => {
  render(<NodeLogCard log={WARN_LOG} />);
  expect(screen.getByText("Warning")).toBeVisible();
  expect(screen.getByText("wifi_manager")).toBeVisible();
  expect(screen.getByText("AC276E5E030C")).toBeVisible();
  await userEvent.click(screen.getByRole("button", { name: "Show message" }));
  expect(screen.getByText(WARN_LOG.message)).toBeVisible();
});
```

- [ ] **Step 2: Verify failure**

```bash
npm test -- src/app/'(authenticated)'/node-logs
```

- [ ] **Step 3: Implement global log view**

Support exact backend filters `logged_at_start`, `logged_at_end`,
`node_device_id`, and `level`. Render semantic level badges and expandable
messages. Add smart refresh and bounded-range guidance.

- [ ] **Step 4: Integrate node tabs**

The node detail Actions, Telemetry, and Node Logs tabs reuse the global card
components while applying immutable node filters. Their “View all” links open
the global page with the same node filter.

- [ ] **Step 5: Verify phase and commit**

```bash
npm test
npm run typecheck
npm run lint
npm run build
git add frontend/src/app/'(authenticated)'/node-logs frontend/src/app/'(authenticated)'/nodes/'[id]'
git commit -m "feat(frontend): add live node logs"
```
