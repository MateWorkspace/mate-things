# Frontend Runtime Optimization and Decomposition Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remove duplicated refresh work, add reliable route boundaries, shorten critical rendering paths, and split Firmware, Node Class, and BLE code into testable responsibilities without changing user-visible behavior.

**Architecture:** Server Components remain responsible for data and permission gates; narrow Client Components own transitions, dialogs, and device APIs. Refresh uses one transition-aware controller, feature routes receive local boundaries, and large modules become feature-local components/services with explicit data flow.

**Tech Stack:** Next.js 16 App Router and streaming, React 19 transitions, TypeScript, Web Bluetooth API, Vitest, Testing Library.

## Global Constraints

- Start from the reviewed Phase 2 commit and work only below `mate-things/frontend`.
- Read the local Next.js 16 fetching, streaming, caching, and error-handling documentation before Tasks 1 and 2.
- Preserve route paths, query keys, permissions, BLE protocol bytes, reconnect timings, log capacity, and firmware mutation contracts.
- Optimize measured or structurally proven duplication only; do not add memoization without a stable prop/subscription boundary.
- Keep commits independently revertible by responsibility.

---

## Task 1: Unify transition-aware smart refresh

**Files:**

- Create: `src/hooks/use-smart-refresh.test.tsx`
- Modify: `src/hooks/use-smart-refresh.ts`
- Modify: `src/components/refresh/RefreshBoundary.tsx`
- Modify: `src/components/refresh/RefreshControl.tsx`
- Modify: `src/app/(authenticated)/dashboard/_components/DashboardRefresh.tsx`
- Modify: `src/app/(authenticated)/dashboard/page.tsx`
- Create: `src/components/refresh/RefreshBoundary.test.tsx`

**Interfaces:**

```ts
type SmartRefreshOptions = {
  updatedAt: string;
  intervalMs?: number;
  enabled?: boolean;
};

type SmartRefreshResult = {
  isPending: boolean;
  isPaused: boolean;
  refresh(): void;
  togglePaused(): void;
  lastUpdatedAt: string;
};
```

**Test shape:**

```tsx
it("waits for a new server timestamp", async () => {
  const { rerender } = render(
    <RefreshBoundary updatedAt="2026-08-10T10:00:00Z">content</RefreshBoundary>,
  );
  await userEvent.click(screen.getByRole("button", { name: /refresh/i }));
  expect(screen.getByText(/10:00/)).toBeVisible();
  rerender(
    <RefreshBoundary updatedAt="2026-08-10T10:01:00Z">content</RefreshBoundary>,
  );
  expect(screen.getByText(/10:01/)).toBeVisible();
});
```

**Implementation shape:**

```ts
const [isPending, startTransition] = useTransition();
const refresh = useCallback(() => {
  if (isPending) return;
  startTransition(() => router.refresh());
}, [isPending, router]);
const lastUpdatedAt = updatedAt;
```

- [ ] Write failing fake-timer tests for automatic refresh, manual refresh, pause/resume, hidden-tab suspension, visibility resume, cleanup, and no overlapping transition request.
- [ ] Prove the timestamp does not advance when `router.refresh()` is merely invoked; it advances only when a new server-provided `updatedAt` prop is rendered.
- [ ] Run `npm test -- src/hooks/use-smart-refresh.test.tsx src/components/refresh/RefreshBoundary.test.tsx` and confirm RED.
- [ ] Refactor `use-smart-refresh` around `startTransition`, page visibility, and the server timestamp. Keep timers in the hook and presentation in `RefreshControl`.
- [ ] Make Dashboard compose `RefreshBoundary`; delete its duplicate interval/visibility controller. Pass the timestamp produced by its server data fetch.
- [ ] Run focused tests, typecheck, lint, and build.
- [ ] Commit: `git add src/hooks src/components/refresh src/app/'(authenticated)'/dashboard && git commit -m "refactor(frontend): unify smart refresh"`.

## Task 2: Complete route boundaries and shorten shell critical work

**Files:**

- Create: `src/app/(authenticated)/ble-direct/loading.tsx`
- Create: `src/app/(authenticated)/ble-direct/error.tsx`
- Create: `src/app/(authenticated)/broadcast-sessions/loading.tsx`
- Create: `src/app/(authenticated)/broadcast-sessions/error.tsx`
- Create: `src/app/(authenticated)/admin/api-keys/loading.tsx`
- Create: `src/app/(authenticated)/admin/api-keys/error.tsx`
- Create: `src/app/(authenticated)/node-classes/loading.tsx`
- Create: `src/app/(authenticated)/node-classes/error.tsx`
- Create: `src/app/(authenticated)/node-classes/[id]/loading.tsx`
- Create: `src/app/(authenticated)/node-classes/[id]/error.tsx`
- Create: `src/app/(authenticated)/node-classes/[id]/not-found.tsx`
- Create: `src/app/(authenticated)/nodes/[id]/error.tsx`
- Modify: `src/app/(authenticated)/layout.tsx`
- Modify: `src/components/layout/AppShell.tsx`
- Modify: `src/components/layout/AppBar.tsx`
- Create: `src/app/(authenticated)/layout.test.tsx`
- Create: `src/app/(authenticated)/route-boundaries.test.tsx`

**Test shape:**

```tsx
it("does not await a role-name request before rendering the shell", async () => {
  await renderServerComponent(AuthenticatedLayout, { children: <p>Ready</p> });
  expect(getRoleById).not.toHaveBeenCalled();
  expect(screen.getByText("Ready")).toBeVisible();
});
```

**Implementation shape:**

```tsx
const session = await requireSessionContext();
return (
  <AppShell
    user={session.user}
    permissions={[...session.permissions]}
    roleName={undefined}
  >
    {children}
  </AppShell>
);
```

- [ ] Write route-boundary tests asserting feature-specific loading copy/skeleton, retryable error UI, and explicit detail 404 behavior for the listed routes.
- [ ] Write a layout test proving AppShell rendering is not blocked by a sequential role-name request after session resolution.
- [ ] Run both tests and confirm RED.
- [ ] Add meaningful local boundaries using the shared loading/error/not-found state components. Error boundaries must be Client Components and expose reset; not-found must not imply access denial.
- [ ] Remove the role lookup from the authenticated layout critical path. Prefer an already-present session claim; if none exists, omit the role label from initial shell data rather than fetching it serially. Do not weaken the session gate.
- [ ] Where an optional download/label lookup remains, catch only documented expected `ApiError` statuses; rethrow authorization, server, and network failures.
- [ ] Run focused tests, typecheck, lint, and build.
- [ ] Commit: `git add src/app/'(authenticated)' src/components/layout && git commit -m "refactor(frontend): add feature route boundaries"`.

## Task 3: Decompose Firmware forms and actions

**Files:**

- Create: `src/app/(authenticated)/firmware/_components/form/FirmwareFields.tsx`
- Create: `src/app/(authenticated)/firmware/_components/form/ConfigSchemaFields.tsx`
- Create: `src/app/(authenticated)/firmware/_components/dialogs/UploadFirmwareDialog.tsx`
- Create: `src/app/(authenticated)/firmware/_components/dialogs/EditFirmwareDialog.tsx`
- Create: `src/app/(authenticated)/firmware/_components/dialogs/ReplaceBinaryDialog.tsx`
- Create: `src/app/(authenticated)/firmware/_components/dialogs/DeleteFirmwareDialog.tsx`
- Create: `src/app/(authenticated)/firmware/_lib/firmware-form-data.ts`
- Create: `src/app/(authenticated)/firmware/_lib/firmware-form-data.test.ts`
- Modify: `src/app/(authenticated)/firmware/_components/FirmwareForm.tsx`
- Modify: `src/app/(authenticated)/firmware/_lib/actions.ts`
- Modify: `src/app/(authenticated)/firmware/page.tsx`
- Modify: `src/app/(authenticated)/firmware/[id]/page.tsx`

**Test shape:**

```ts
it.each([
  ["keep", undefined],
  ["clear", []],
])("maps %s schema mode", async (mode, configSchema) => {
  await replaceFirmwareBinaryAction(
    INITIAL_ACTION_STATE,
    formData({ id: "fw-1", schema_mode: mode, binary }),
  );
  expect(replaceFirmwareBinary).toHaveBeenCalledWith(
    expect.objectContaining({ configSchema }),
  );
});
```

**Implementation shape:**

```tsx
export function ReplaceBinaryDialog({ firmware, canReplace }: Props) {
  if (!canReplace) return null;
  return (
    <Dialog dismissible={!pending}>
      <ReplaceBinaryActionForm key={formKey} firmware={firmware} />
    </Dialog>
  );
}
```

- [ ] Extend Phase 1 firmware tests to characterize upload, edit, replacement Keep/Replace/Clear, delete confirmation, OTA entry, pending dismissal, errors, and successful reset.
- [ ] Write failing pure-parser tests for metadata, binary, schema intent, and delete form data. Each parser returns a typed success/error result and preserves user-entered values.
- [ ] Run firmware tests and confirm the parser tests RED.
- [ ] Extract shared fields/schema controls without coupling them to a specific dialog's action state.
- [ ] Extract four dialogs. Each owns only open state, action submission, feedback, and success close/reset; reuse `useActionDialog` and shared form messages.
- [ ] Split the large action module into private parsers/permission helpers plus named public actions. Keep `"use server"` only on public action entry points and keep raw API imports server-only.
- [ ] Leave `FirmwareForm.tsx` as a small compatibility composition or update callers and remove it after `rg` shows no imports.
- [ ] Run focused tests, typecheck, lint, and build.
- [ ] Commit: `git add src/app/'(authenticated)'/firmware && git commit -m "refactor(frontend): decompose firmware workflows"`.

## Task 4: Decompose Node Class relationships

**Files:**

- Create: `src/app/(authenticated)/node-classes/[id]/_components/NodeClassOverview.tsx`
- Create: `src/app/(authenticated)/node-classes/[id]/_components/NodeClassNodes.tsx`
- Create: `src/app/(authenticated)/node-classes/[id]/_components/NodeClassFirmware.tsx`
- Create: `src/app/(authenticated)/node-classes/[id]/_components/NodeClassActions.tsx`
- Create: `src/app/(authenticated)/node-classes/[id]/_components/relationship-section.test.tsx`
- Modify: `src/app/(authenticated)/node-classes/[id]/page.tsx`
- Modify: `src/app/(authenticated)/node-classes/_components/NodeClassActionChecklist.tsx`
- Modify: `src/app/(authenticated)/node-classes/[id]/page.test.tsx`

**Test shape:**

```tsx
it("hides relationship mutation while retaining readable relationships", () => {
  render(<NodeClassActions actions={actions} canAssign={false} />);
  expect(screen.getByRole("heading", { name: "Actions" })).toBeVisible();
  expect(
    screen.queryByRole("button", { name: /save assignments/i }),
  ).toBeNull();
});
```

**Implementation shape:**

```tsx
const [nodes, firmware, actions] = await Promise.all([
  canReadNodes ? listAllNodesForClass(nodeClass.id) : Promise.resolve([]),
  canReadFirmware
    ? listAllFirmwaresForClass(nodeClass.id)
    : Promise.resolve([]),
  canReadActions ? listAllActionsForClass(nodeClass.id) : Promise.resolve([]),
]);
```

- [ ] Expand the existing page test to characterize permissions, counts, relationship empty states, detail links, firmware/actions metadata, and checklist mutation entry.
- [ ] Add failing component tests for semantic section headings, definition lists, empty states, and permission-hidden actions.
- [ ] Run the page/component tests and confirm only new component tests RED.
- [ ] Extract four presentation components with typed props. Fetch data in the page in parallel where permissions permit; components must not start their own route-level requests.
- [ ] Reuse Phase 2 `DetailSection`, `MetadataGrid`, card links, and collection primitives without hiding relationship-specific content behind a schema.
- [ ] Keep the action checklist stateful boundary narrow and preserve Phase 1 partial-failure/refresh behavior.
- [ ] Run focused tests, typecheck, lint, and build.
- [ ] Commit: `git add src/app/'(authenticated)'/node-classes && git commit -m "refactor(frontend): split node class relationships"`.

## Task 5: Split BLE transport, connection state, and rendering

**Files:**

- Rename: `src/app/(authenticated)/ble-direct/_lib/ble/BleClient.ts` to `src/app/(authenticated)/ble-direct/_lib/ble/ble-client.ts`
- Create: `src/app/(authenticated)/ble-direct/_lib/ble/gatt-transport.ts`
- Create: `src/app/(authenticated)/ble-direct/_lib/ble/connection-machine.ts`
- Create: `src/app/(authenticated)/ble-direct/_lib/ble/ble-store.ts`
- Create: `src/app/(authenticated)/ble-direct/_lib/ble/ble-client.test.ts`
- Create: `src/app/(authenticated)/ble-direct/_lib/ble/connection-machine.test.ts`
- Rename: `src/app/(authenticated)/ble-direct/_lib/useBleConnection.ts` to `src/app/(authenticated)/ble-direct/_lib/use-ble-connection.ts`
- Modify: `src/app/(authenticated)/ble-direct/_components/BleDirectRoot.tsx`
- Modify: `src/app/(authenticated)/ble-direct/_components/ConnectionBanner.tsx`
- Modify: `src/app/(authenticated)/ble-direct/_components/blocks/LogBlock.tsx`
- Modify: `src/app/(authenticated)/ble-direct/_components/blocks/SystemInfoBlock.tsx`
- Modify: `src/app/(authenticated)/ble-direct/_components/blocks/WifiManagerBlock.tsx`
- Rename: `src/app/(authenticated)/ble-direct/_components/blocks/SettingsBlock/SettingsBlock.tsx` to `src/app/(authenticated)/ble-direct/_components/blocks/settings-block/SettingsBlock.tsx`
- Rename: `src/app/(authenticated)/ble-direct/_components/blocks/SettingsBlock/SettingsField.tsx` to `src/app/(authenticated)/ble-direct/_components/blocks/settings-block/SettingsField.tsx`
- Create: `src/app/(authenticated)/ble-direct/_components/BleDirectRoot.test.tsx`
- Modify: `src/app/(authenticated)/ble-direct/page.tsx`

**Interfaces:**

```ts
export type BleLogEntry = { id: number; text: string; receivedAt: number };
export type BleConnectionSnapshot = {
  status: BleStatus;
  error?: string;
  deviceName?: string;
};

export interface GattTransport {
  connect(): Promise<void>;
  disconnect(): Promise<void>;
  write(characteristic: string, data: Uint8Array): Promise<void>;
  subscribe(
    characteristic: string,
    listener: (data: DataView) => void,
  ): Promise<() => void>;
}
```

**Test shape:**

```ts
it("ignores disconnect from an obsolete generation", () => {
  const connected = reduceConnection(initialState, {
    type: "connect.started",
    generation: 2,
  });
  const stale = reduceConnection(connected, {
    type: "device.disconnected",
    generation: 1,
  });
  expect(stale).toBe(connected);
});

it("keeps stable panels out of log subscriptions", () => {
  const renders = renderBleWithCounters();
  emitLog("hello");
  expect(renders.log).toBe(2);
  expect(renders.systemInfo).toBe(1);
});
```

**Implementation shape:**

```ts
appendLog(text: string) {
  const entry = { id: this.nextLogId++, text, receivedAt: Date.now() };
  this.logs = [...this.logs.slice(-(MAX_LOG_ENTRIES - 1)), entry];
  this.emitLogChange();
}
```

- [ ] Add mocked Web Bluetooth tests for unsupported browser, user-cancelled chooser, connect/disconnect, remote disconnect, timeout, reconnect/restart, characteristic notification, and cleanup on unmount.
- [ ] Add reducer/state-machine tests for valid transitions and refusal of stale async events from an earlier connection generation.
- [ ] Add rendering tests proving a log notification updates LogBlock but does not rerender stable System Info, Wi-Fi, or Settings panels. Verify live-region behavior and focus during pending/error states.
- [ ] Run BLE tests and confirm failures expose current coupling/index-key behavior.
- [ ] Extract browser/GATT calls behind `GattTransport`; move reconnect/timeouts into a pure connection machine; keep codec/protocol/UUID modules unchanged.
- [ ] Implement a small external store or split subscriptions with `useSyncExternalStore`: connection/data selectors subscribe independently from the bounded log list.
- [ ] Store monotonically identified log entries, cap at the existing 500 entries, and key rows by `entry.id` instead of array index.
- [ ] Update root/panels to subscribe only to needed slices. Apply `memo` only to panels whose props are demonstrably stable after the split.
- [ ] Perform renames with `git mv`, update exact import casing, and keep BLE protocol/timing behavior unchanged.
- [ ] Run focused tests, typecheck, lint, and build in a browser-capable type environment.
- [ ] Commit: `git add src/app/'(authenticated)'/ble-direct && git commit -m "refactor(frontend): isolate BLE runtime state"`.

## Task 6: Remove confirmed dead code and naming drift

**Files:**

- Inspect/remove if unreferenced: `src/assets/matethings-logo.svg`
- Inspect/remove if unreferenced: `src/assets/vertical.svg`
- Inspect/remove if unreferenced: `src/assets/horizontal.svg`
- Inspect/remove if superseded: `src/lib/route-access.ts`
- Modify: `src/app/(authenticated)/ble-direct/page.tsx`
- Modify: `src/app/(authenticated)/ble-direct/_components/BleDirectRoot.tsx`
- Modify: `src/app/(authenticated)/ble-direct/_components/blocks/settings-block/SettingsBlock.tsx`

**Verification shape:**

```bash
rg -n 'matethings-logo|vertical\.svg|horizontal\.svg|route-access|BleClient|useBleConnection|SettingsBlock' src e2e
npm run typecheck
npm test
```

An asset/module is deletable only when the first command has no live import for it and the latter two commands remain green after removal.

- [ ] Use `rg -n "matethings-logo|vertical\.svg|horizontal\.svg|route-access|BleClient|useBleConnection|SettingsBlock" src e2e` and record every match.
- [ ] Delete only files with zero runtime/test/config references and a clear replacement. Do not delete assets based on filename intuition.
- [ ] Run typecheck and the entire test suite after each deletion/rename group so dynamic/case-sensitive imports cannot be missed.
- [ ] Run `find src -type f | sort` and check the frontend naming conventions documented in `frontend/AGENTS.md`; keep React component filenames PascalCase where the guide requires them and directories/hooks/helpers kebab-case.
- [ ] Run lint, build, `git diff --check`, and inspect the deletion list.
- [ ] Commit: `git add -A src && git commit -m "chore(frontend): remove superseded frontend code"`.

## Phase Gate

- [ ] Run `npm run typecheck`, `npm run lint`, `npm test`, and `npm run build`.
- [ ] Manually exercise smart refresh, every new error/loading/not-found boundary, firmware replacement modes, node-class relationships, and the BLE unsupported/disconnect flows.
- [ ] Compare BLE protocol constants, command encoding, reconnect delays, timeouts, and log capacity before/after the phase.
- [ ] Run `npm run format:check`, `git diff --check`, and verify all changed paths are inside `frontend/`.
