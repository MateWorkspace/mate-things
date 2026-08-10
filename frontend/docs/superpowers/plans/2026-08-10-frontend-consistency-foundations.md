# Frontend Consistency Foundations Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Correct the frontend's routing, server boundary, query, permission, dialog, and mutation semantics before shared visual components are introduced.

**Architecture:** Keep API transport in server-only modules, expose narrowly scoped and permission-checked Server Actions to Client Components, derive route behavior from one policy registry, and centralize query/action-state primitives. Preserve all routes and query parameter names while removing the unsupported post-pagination connection filter.

**Tech Stack:** Next.js 16 App Router, React 19 Server Components and Server Actions, TypeScript strict mode, Tailwind CSS v4, Vitest, Testing Library.

## Global Constraints

- Work only below `mate-things/frontend` and preserve unrelated worktree changes.
- Read the local Next.js 16 Server Actions, proxy, and error-handling documentation before Tasks 1 and 2.
- Do not change backend contracts except to accurately represent contracts the backend already exposes.
- Every public Server Action must validate direct-call input and recheck its exact permission.
- Preserve remove-only roles: do not add read-permission requirements to destructive operations whose backend already performs authoritative confirmation.
- Run the focused test before and after each implementation step; do not proceed from an unexplained failure.

---

## Task 1: Establish one route policy registry

**Files:**

- Create: `src/config/route-policies.ts`
- Create: `src/config/route-policies.test.ts`
- Modify: `src/config/navigation.ts`
- Modify: `src/lib/route-access.ts`
- Modify: `src/proxy.ts`
- Create: `src/proxy.test.ts`

**Interfaces:**

```ts
export type RoutePolicy = {
  href: string;
  protected: boolean;
  requiredAny?: readonly PermissionName[];
  navigation?: {
    group:
      "overview" | "fleet" | "operations" | "observability" | "administration";
    label: string;
  };
};

export const ROUTE_POLICIES: readonly RoutePolicy[];
export function findRoutePolicy(pathname: string): RoutePolicy | undefined;
export function isProtectedRoute(pathname: string): boolean;
export function canVisitRoute(
  pathname: string,
  permissions: ReadonlySet<string>,
): boolean;
```

**Test shape:**

```ts
it.each([
  ["/ble-direct", true],
  ["/broadcast-sessions/live", true],
  ["/admin/api-keys", true],
  ["/login", false],
])("classifies %s", (pathname, expected) => {
  expect(isProtectedRoute(pathname)).toBe(expected);
});
```

**Implementation shape:**

```ts
const matchesPrefix = (pathname: string, href: string) =>
  pathname === href || pathname.startsWith(`${href}/`);

export function findRoutePolicy(pathname: string) {
  return [...ROUTE_POLICIES]
    .sort((left, right) => right.href.length - left.href.length)
    .find((policy) => matchesPrefix(pathname, policy.href));
}
```

- [ ] Read `src/config/navigation.ts`, `src/lib/route-access.ts`, and `src/proxy.ts`; enumerate every navigation href and every detail prefix, including `/ble-direct`, `/broadcast-sessions`, and `/admin/api-keys`.
- [ ] Write a table-driven failing test asserting that each protected path is refreshed by proxy policy and that each navigation item resolves to the expected permission rule. Include exact paths and nested detail paths.
- [ ] Run `npm test -- src/config/route-policies.test.ts src/proxy.test.ts` and confirm the omissions fail.
- [ ] Implement `ROUTE_POLICIES` with longest-prefix matching on path-segment boundaries (`/nodes` matches `/nodes/abc`, never `/nodes-old`). Store navigation group/label on navigable policies and derive `isProtectedRoute` and `canVisitRoute` from the same entries.
- [ ] Refactor navigation and proxy to consume the registry. Keep only the icon lookup and presentation-specific group ordering in `navigation.ts`; do not import React/icon values into the server-safe policy module.
- [ ] Reduce `route-access.ts` to compatibility exports from the registry, or delete it and update all imports if `rg "route-access" src` proves it is dead.
- [ ] Run the focused tests, then `npm run typecheck`.
- [ ] Commit: `git add src/config src/lib/route-access.ts src/proxy.ts src/proxy.test.ts && git commit -m "refactor(frontend): centralize route policy"`.

## Task 2: Separate server-only transport from public selector actions

**Files:**

- Modify: `src/lib/api/action-logs.ts`
- Modify: `src/lib/api/actions.ts`
- Modify: `src/lib/api/api-keys.ts`
- Modify: `src/lib/api/auth.ts`
- Modify: `src/lib/api/client.ts`
- Modify: `src/lib/api/firmwares.ts`
- Modify: `src/lib/api/node-classes.ts`
- Modify: `src/lib/api/node-config.ts`
- Modify: `src/lib/api/node-logs.ts`
- Modify: `src/lib/api/nodes.ts`
- Modify: `src/lib/api/ota.ts`
- Modify: `src/lib/api/payload-schemas.ts`
- Modify: `src/lib/api/permissions.ts`
- Modify: `src/lib/api/preferences.ts`
- Modify: `src/lib/api/profile.ts`
- Modify: `src/lib/api/role-permissions.ts`
- Modify: `src/lib/api/roles.ts`
- Delete: `src/lib/api/session-actions.ts`
- Modify: `src/lib/api/telemetry.ts`
- Modify: `src/lib/api/users.ts`
- Create: `src/lib/actions/session-actions.ts`
- Create: `src/lib/actions/search-options.ts`
- Create: `src/lib/actions/search-options.test.ts`
- Create: `src/lib/actions/entity-search-actions.ts`
- Create: `src/lib/actions/entity-search-actions.test.ts`
- Modify: `src/components/actions/ActionSearchCombobox.tsx`
- Modify: `src/components/node-classes/NodeClassSearchCombobox.tsx`
- Modify: `src/components/nodes/NodeSearchCombobox.tsx`
- Modify: `src/components/nodes/NodeDeviceIdSearchCombobox.tsx`
- Modify: `src/components/roles/RoleSearchCombobox.tsx`
- Modify: `src/components/users/UserSearchCombobox.tsx`
- Modify: `src/components/layout/AppShell.tsx`
- Modify: `src/components/layout/LogoutButton.tsx`

**Interfaces:**

```ts
export type SearchOption = {
  value: string;
  label: string;
  description?: string;
};
export type SearchOptionsRequest = {
  query: string;
  page: number;
  limit?: number;
};
export type SearchOptionsPage = {
  items: SearchOption[];
  page: number;
  totalPages: number;
};

export async function searchNodesAction(
  input: unknown,
): Promise<SearchOptionsPage>;
export async function searchNodeDeviceIdsAction(
  input: unknown,
): Promise<SearchOptionsPage>;
export async function searchActionsAction(
  input: unknown,
): Promise<SearchOptionsPage>;
export async function searchNodeClassesAction(
  input: unknown,
): Promise<SearchOptionsPage>;
export async function searchRolesAction(
  input: unknown,
): Promise<SearchOptionsPage>;
export async function searchUsersAction(
  input: unknown,
): Promise<SearchOptionsPage>;
```

**Test shape:**

```ts
it("checks node permission before transport", async () => {
  mockRequirePermission.mockRejectedValue(new Error("forbidden"));
  await expect(searchNodesAction({ query: "lab", page: 1 })).rejects.toThrow(
    "forbidden",
  );
  expect(listNodes).not.toHaveBeenCalled();
});
```

**Implementation shape:**

```ts
"use server";

export async function searchNodesAction(
  input: unknown,
): Promise<SearchOptionsPage> {
  const request = parseSearchOptionsRequest(input);
  await requirePermission("node:get");
  const response = await listNodes(request);
  return toSearchOptionsPage(response, (node) => ({
    value: node.id,
    label: node.name,
    description: node.device_id,
  }));
}
```

- [ ] Add failing contract tests that scan/import API resource modules and prove they are not Client-callable Server Actions, while each selector action rejects malformed input and missing permission before invoking transport.
- [ ] Cover exact permission mappings: `node:get`, `action:get`, `node_class:get`, `role:get`, and `user:get`. Assert bounded `limit`, positive safe `page`, trimmed `query`, and transport errors are not swallowed.
- [ ] Run `npm test -- src/lib/actions/search-options.test.ts src/lib/actions/entity-search-actions.test.ts` and confirm RED.
- [ ] Replace top-level `"use server"` in the listed raw API wrappers with `import "server-only"`; add the marker to `client.ts`. Move `logoutAction` and `setSidebarCollapsedAction` unchanged from `src/lib/api/session-actions.ts` to `src/lib/actions/session-actions.ts`, update AppShell/LogoutButton imports, and delete the old mixed-boundary module. Keep shared response-only `types.ts` free of a server marker.
- [ ] Implement one unknown-input parser and the six permission-checked adapters. Return presentation-safe options rather than raw API response types.
- [ ] Temporarily update each existing combobox to call its explicit selector action. Do not refactor combobox rendering here; Phase 2 replaces it with one engine.
- [ ] Verify there is no client-to-transport import with `rg -l '"use client"' src | xargs -r rg 'lib/api/'` and inspect every match.
- [ ] Run the focused tests, `npm run typecheck`, and `npm run lint`.
- [ ] Commit: `git add src/lib/api src/lib/actions src/components && git commit -m "refactor(frontend): isolate server transport"`.

## Task 3: Centralize query parsing and complete-page collection

**Files:**

- Create: `src/lib/query.ts`
- Create: `src/lib/query.test.ts`
- Create: `src/lib/api/collect-all-pages.ts`
- Create: `src/lib/api/collect-all-pages.test.ts`
- Modify: `src/lib/collection-query.ts`
- Modify: `src/lib/record-filters.ts`
- Modify: `src/lib/api/actions.ts`
- Modify: `src/lib/api/node-classes.ts`
- Modify: `src/lib/api/payload-schemas.ts`
- Modify: `src/lib/api/roles.ts`
- Modify: `src/lib/api/permissions.ts`
- Modify: `src/lib/api/nodes.ts`

**Interfaces:**

```ts
export type RawSearchParams = Record<string, string | string[] | undefined>;
export function firstQueryValue(
  value: string | string[] | undefined,
): string | undefined;
export function parsePositiveSafeInteger(
  value: unknown,
  fallback: number,
): number;
export function parseBooleanQuery(value: unknown): boolean | undefined;
export function parseEnumQuery<const T extends readonly string[]>(
  value: unknown,
  values: T,
): T[number] | undefined;
export function parseLocalDateTime(value: unknown): Date | undefined;
export function toUtcQueryValue(value: Date): string;
export function buildCollectionUrl(
  pathname: string,
  values: Readonly<Record<string, string | undefined>>,
): string;
export function buildOutOfRangeRedirect(
  pathname: string,
  values: RawSearchParams,
  lastPage: number,
): string;

export async function collectAllPages<T>(options: {
  fetchPage: (
    page: number,
  ) => Promise<{ data: T[]; page: number; limit: number; total: number }>;
  keyOf: (item: T) => string;
  maxPages?: number;
}): Promise<T[]>;
```

**Test shape:**

```ts
it("deduplicates and stops when a page makes no progress", async () => {
  const fetchPage = vi
    .fn()
    .mockResolvedValueOnce({ data: [{ id: "a" }], page: 1, limit: 1, total: 3 })
    .mockResolvedValueOnce({
      data: [{ id: "a" }],
      page: 2,
      limit: 1,
      total: 3,
    });
  await expect(
    collectAllPages({ fetchPage, keyOf: (item) => item.id }),
  ).resolves.toEqual([{ id: "a" }]);
  expect(fetchPage).toHaveBeenCalledTimes(2);
});
```

**Implementation shape:**

```ts
export function firstQueryValue(value: string | string[] | undefined) {
  return Array.isArray(value) ? value[0] : value;
}

export function parseBooleanQuery(value: unknown) {
  return value === "true" ? true : value === "false" ? false : undefined;
}
```

- [ ] Write failing tests for first-value selection, safe integer overflow/zero/negative handling, exact enum matching, strict boolean parsing, valid/invalid local datetime conversion, canonical UTC output, filtered URL encoding, active-filter preservation during out-of-range redirect, and preservation of empty strings where empty has meaning.
- [ ] Write failing pagination tests for duplicate items, inconsistent response limits, a non-advancing backend page, empty intermediate pages, and a configurable safety ceiling. The helper must terminate without division by zero.
- [ ] Run `npm test -- src/lib/query.test.ts src/lib/api/collect-all-pages.test.ts` and confirm RED.
- [ ] Implement the shared parsers; make `collection-query.ts` and `record-filters.ts` compose them rather than maintain copies.
- [ ] Implement `collectAllPages` using response totals plus progress detection. Deduplicate by `keyOf` while retaining first-seen order.
- [ ] Replace resource-specific all-page loops and fixed 48/100 option caps in the named API modules with the helper. Preserve each API's server-side search arguments.
- [ ] Add/update focused resource tests for access-control role/permission options, user roles, action-compatible nodes, and payload schemas beyond page one.
- [ ] Run focused tests, `npm run typecheck`, and `npm run lint`.
- [ ] Commit: `git add src/lib && git commit -m "refactor(frontend): share query and paging policy"`.

## Task 4: Fix collection and ancillary-read correctness

**Files:**

- Modify: `src/app/(authenticated)/nodes/page.tsx`
- Modify: `src/app/(authenticated)/nodes/_components/NodeFilters.tsx`
- Modify: `src/app/(authenticated)/nodes/_components/NodeCollectionEmptyState.tsx`
- Modify: `src/app/(authenticated)/dashboard/page.tsx`
- Modify: `src/app/(authenticated)/dashboard/_lib/dashboard-data.ts`
- Modify: `src/app/(authenticated)/actions/page.tsx`
- Modify: `src/app/(authenticated)/action-history/page.tsx`
- Modify: `src/app/(authenticated)/action-history/_components/ActionLogFilters.tsx`
- Modify: `src/app/(authenticated)/telemetry/page.tsx`
- Modify: `src/app/(authenticated)/telemetry/_components/TelemetryFilters.tsx`
- Modify: `src/app/(authenticated)/node-logs/page.tsx`
- Modify: `src/app/(authenticated)/node-logs/_components/NodeLogFilters.tsx`
- Create: `src/app/(authenticated)/nodes/page.test.tsx`
- Create: `src/app/(authenticated)/action-history/page.test.tsx`
- Create: `src/app/(authenticated)/telemetry/page.test.tsx`
- Create: `src/app/(authenticated)/node-logs/page.test.tsx`
- Create: `src/app/(authenticated)/actions/page.test.tsx`

**Test shape:**

```ts
it("does not fetch a node label without node:get", async () => {
  mockPermissions(["node_log:get"]);
  await renderServerPage(NodeLogsPage, { node_id: "node-1" });
  expect(getNodeById).not.toHaveBeenCalled();
  expect(screen.getByDisplayValue("node-1")).toBeInTheDocument();
});

it("does not filter a fetched node page in memory", async () => {
  await renderServerPage(NodesPage, { is_connected: "true" });
  expect(listNodes).toHaveBeenCalledWith(
    expect.not.objectContaining({ is_connected: true }),
  );
});
```

**Implementation shape:**

```ts
const selectedNode =
  canReadNodes && nodeId
    ? await getOptionalById(() => getNodeById(nodeId))
    : null;

// Render nodeId as a read-only/raw-ID filter when canReadNodes is false.
```

- [ ] Write failing route tests proving the Nodes page never applies `is_connected` after a page has been fetched, Actions redirects an out-of-range page while retaining active filters, and record routes suppress node/action label reads when the corresponding permission is absent.
- [ ] Assert a 404 from an optional label lookup yields a raw-ID fallback, while 401/403/500 and network failures reach the route error boundary instead of becoming `null`.
- [ ] Run the five route tests and confirm RED.
- [ ] Remove the unsupported Nodes connection query UI and post-pagination filtering. Ignore an old `is_connected` URL value safely and remove dashboard links that manufacture it. Label dashboard connection figures prominently as a bounded sample until the backend provides exact aggregates.
- [ ] Apply the standard page normalization/out-of-range redirect to Actions, preserving its existing query keys.
- [ ] Pass capability flags into record filters. Render the relevant selector only with its read permission; otherwise render a non-searchable raw-ID field/label. Avoid any forbidden ancillary transport call.
- [ ] Replace broad `.catch(() => null)` with an `ApiError` 404-only helper; rethrow all unexpected failures.
- [ ] Run the focused route tests, `npm run typecheck`, and `npm run lint`.
- [ ] Commit: `git add src/app/'(authenticated)' && git commit -m "fix(frontend): correct collection permissions and paging"`.

## Task 5: Introduce reusable action and dialog state

**Files:**

- Create: `src/lib/forms/action-state.ts`
- Create: `src/lib/forms/action-state.test.ts`
- Create: `src/lib/forms/parse.ts`
- Create: `src/lib/forms/parse.test.ts`
- Create: `src/components/forms/ActionMessage.tsx`
- Create: `src/components/forms/FieldError.tsx`
- Create: `src/hooks/use-action-dialog.ts`
- Create: `src/hooks/use-action-dialog.test.tsx`
- Create: `src/hooks/use-action-feedback.ts`
- Create: `src/hooks/use-refresh-after-action.ts`
- Create: `src/hooks/use-first-invalid-field.ts`
- Create: `src/hooks/action-form-hooks.test.tsx`
- Modify: `src/components/ui/dialog.tsx`
- Modify: `src/components/ui/dialog.test.tsx`
- Modify: `src/app/(authenticated)/actions/_components/ActionForm.tsx`
- Modify: `src/app/(authenticated)/actions/_components/DispatchActionDialog.tsx`
- Modify: `src/app/(authenticated)/admin/access-control/_components/PermissionCard.tsx`
- Modify: `src/app/(authenticated)/admin/access-control/_components/RoleDetails.tsx`
- Modify: `src/app/(authenticated)/admin/access-control/_lib/state.ts`
- Modify: `src/app/(authenticated)/admin/api-keys/_components/DeleteApiKeyDialog.tsx`
- Modify: `src/app/(authenticated)/admin/api-keys/_components/GenerateApiKeyDialog.tsx`
- Modify: `src/app/(authenticated)/admin/api-keys/_components/RegenerateApiKeyDialog.tsx`
- Modify: `src/app/(authenticated)/admin/api-keys/_components/RevokeApiKeyDialog.tsx`
- Modify: `src/app/(authenticated)/admin/api-keys/_lib/state.ts`
- Modify: `src/app/(authenticated)/admin/payload-schemas/_components/PayloadSchemaForm.tsx`
- Modify: `src/app/(authenticated)/admin/payload-schemas/_lib/state.ts`
- Modify: `src/app/(authenticated)/admin/users/_components/PasswordResetForm.tsx`
- Modify: `src/app/(authenticated)/admin/users/_components/UserForm.tsx`
- Modify: `src/app/(authenticated)/admin/users/_lib/state.ts`
- Modify: `src/app/(authenticated)/firmware/_components/FirmwareForm.tsx`
- Modify: `src/app/(authenticated)/firmware/_components/OtaDialog.tsx`
- Modify: `src/app/(authenticated)/node-classes/_components/NodeClassActionChecklist.tsx`
- Modify: `src/app/(authenticated)/node-classes/_components/NodeClassForm.tsx`
- Modify: `src/app/(authenticated)/nodes/[id]/_components/DeleteNodeDialog.tsx`
- Modify: `src/app/(authenticated)/nodes/[id]/_components/NodeConfigForm.tsx`
- Modify: `src/app/(authenticated)/nodes/[id]/_components/NodeEditForm.tsx`
- Modify: `src/components/preferences/PreferencesDialog.tsx`
- Modify: `src/components/preferences/preferences-state.ts`
- Modify: `src/components/profile/ProfileForm.tsx`
- Modify: `src/components/profile/SecurityForm.tsx`
- Modify: `src/components/records/ScopedDeleteDialog.tsx`

**Interfaces:**

```ts
export type ActionState<TFields extends string = never> = {
  status: "idle" | "error" | "success";
  title?: string;
  message?: string;
  fieldErrors?: Partial<Record<TFields, string>>;
};

export const INITIAL_ACTION_STATE: ActionState;
export function useActionDialog(options: {
  state: ActionState;
  onSuccess?: () => void;
}): {
  open: boolean;
  setOpen(open: boolean): void;
  reset(): void;
  formKey: number;
};
```

**Test shape:**

```tsx
it("can reopen after success and blocks close while pending", async () => {
  const user = userEvent.setup();
  render(<Harness pending />);
  await user.keyboard("{Escape}");
  expect(screen.getByRole("dialog")).toBeVisible();
  rerender(<Harness pending={false} state={{ status: "success" }} />);
  await user.click(screen.getByRole("button", { name: /open/i }));
  expect(screen.getByRole("dialog")).toBeVisible();
});
```

**Implementation shape:**

```ts
export const INITIAL_ACTION_STATE = { status: "idle" } as const;

export function useActionDialog({ state, onSuccess }: UseActionDialogOptions) {
  const [open, setOpenState] = useState(false);
  const [formKey, setFormKey] = useState(0);
  const reset = useCallback(() => {
    setOpenState(false);
    setFormKey((key) => key + 1);
  }, []);
  useEffect(() => {
    if (state.status === "success") {
      onSuccess?.();
      reset();
    }
  }, [onSuccess, reset, state.status]);
  return { open, setOpen: setOpenState, reset, formKey };
}
```

Mount the component that owns `useActionState` with `key={formKey}` so incrementing the cycle creates a fresh action state. Do not place `useActionState` above that keyed boundary.

- [ ] Inventory route-local action-state shapes with `rg "ActionState|initial.*State|useActionState" src/app src/components`; document intentional payload extensions in the test table.
- [ ] Write failing tests for uniform success/error state, field errors, preserved raw form values on error, exactly-once success toast, authoritative refresh after success/partial failure, first-invalid-field focus, dialog reopening after success, focus restoration, and refusal to Escape/backdrop/native-cancel/close-button dismissal while pending.
- [ ] Run the focused tests and confirm RED.
- [ ] Implement the shared state/parser/display utilities without erasing feature-specific success payloads such as a newly generated secret.
- [ ] Add `dismissible?: boolean` to `Dialog`; when false, ignore Escape and backdrop close while leaving accessible name/description/focus behavior intact.
- [ ] Implement `useActionDialog` so a successful submission closes and resets the state on the next open. Do not retain a terminal success state that makes the dialog permanently unusable.
- [ ] Migrate route-local forms and dialogs incrementally. Replace only truly equivalent state boilerplate; retain domain validation next to its action.
- [ ] Run `npm test -- src/lib/forms src/hooks/use-action-dialog.test.tsx src/components/ui/dialog.test.tsx`, then full typecheck/lint.
- [ ] Commit: `git add src/lib/forms src/components/forms src/hooks src/components/ui/dialog* src/app/'(authenticated)' && git commit -m "refactor(frontend): unify action dialog lifecycle"`.

## Task 6: Correct secret, update, and partial-failure workflows

**Files:**

- Modify: `src/app/(authenticated)/admin/api-keys/_components/GenerateApiKeyDialog.tsx`
- Modify: `src/app/(authenticated)/admin/api-keys/_components/RegenerateApiKeyDialog.tsx`
- Modify: `src/app/(authenticated)/admin/api-keys/_lib/actions.ts`
- Create: `src/app/(authenticated)/admin/api-keys/_components/api-key-dialogs.test.tsx`
- Modify: `src/app/(authenticated)/admin/payload-schemas/_lib/actions.ts`
- Modify: `src/app/(authenticated)/admin/payload-schemas/_components/PayloadSchemaForm.tsx`
- Create: `src/app/(authenticated)/admin/payload-schemas/_lib/actions.test.ts`
- Modify: `src/app/(authenticated)/admin/access-control/_components/RoleDetails.tsx`
- Modify: `src/app/(authenticated)/admin/access-control/_lib/actions.ts`
- Modify: `src/app/(authenticated)/node-classes/_components/NodeClassActionChecklist.tsx`

**Test shape:**

```ts
it("never creates when an update id is missing", async () => {
  const result = await updatePayloadSchemaAction(
    INITIAL_ACTION_STATE,
    formData({ name: "Schema" }),
  );
  expect(result.status).toBe("error");
  expect(createPayloadSchema).not.toHaveBeenCalled();
  expect(updatePayloadSchema).not.toHaveBeenCalled();
});
```

**Implementation shape:**

```ts
export type AssignmentResult = {
  status: "success" | "partial" | "error";
  appliedIds: string[];
  failed: Array<{ id: string; message: string }>;
};

const id = requiredString(formData, "id");
if (!id.ok)
  return actionError("A payload schema ID is required.", { id: id.error });
```

- [ ] Write failing API-key tests proving a generated/regenerated secret is visible exactly once, cleared on close, absent on reopen, and never persisted in ordinary component state longer than the result view.
- [ ] Write a failing payload-schema action test proving missing/tampered update ID returns an update error and never calls create.
- [ ] Write failing assignment tests proving partial failure reports exact failed items, retains retryable selection, and refreshes server data after complete or partial application.
- [ ] Run the focused tests and confirm RED.
- [ ] Make API-key result state ephemeral and clear it on every close path; preserve copy-button and warning UX.
- [ ] Split payload schema create/update action entry points or require an explicit discriminant. Validate update ID before authorization/transport and never fall through to create.
- [ ] Make role-permission and node-class-action updates return structured partial results. Call `router.refresh()` in a transition after applied changes; do not claim data was reloaded unless refresh is triggered.
- [ ] Run focused tests, typecheck, and lint.
- [ ] Commit: `git add src/app/'(authenticated)'/admin src/app/'(authenticated)'/node-classes && git commit -m "fix(frontend): harden administrative workflows"`.

## Task 7: Preserve firmware schema intent

**Files:**

- Modify: `src/lib/api/firmwares.ts`
- Modify: `src/app/(authenticated)/firmware/_lib/actions.ts`
- Modify: `src/app/(authenticated)/firmware/_components/FirmwareForm.tsx`
- Create: `src/lib/api/firmwares.test.ts`
- Create: `src/app/(authenticated)/firmware/_lib/actions.test.ts`
- Create: `src/app/(authenticated)/firmware/_components/FirmwareForm.test.tsx`

**Interfaces:**

```ts
export type FirmwareSchemaIntent =
  | { mode: "keep" }
  | { mode: "replace"; schema: FirmwareConfigParameter[] }
  | { mode: "clear" };

export type ReplaceFirmwareBinaryInput = {
  id: string;
  binary: File;
  configSchema?: FirmwareConfigParameter[];
};
```

**Test shape:**

```ts
it.each([
  ["keep", false, undefined],
  ["replace", true, JSON.stringify([{ key: "rate", type: "uint32" }])],
  ["clear", true, "[]"],
])("serializes %s intent", async (_mode, present, expected) => {
  const body = await capturedMultipartBody();
  expect(body.has("config_schema")).toBe(present);
  if (present) expect(body.get("config_schema")).toBe(expected);
});
```

**Implementation shape:**

```ts
if (input.configSchema !== undefined) {
  body.append("config_schema", JSON.stringify(input.configSchema));
}

const configSchema =
  intent.mode === "keep"
    ? undefined
    : intent.mode === "clear"
      ? []
      : intent.schema;
```

- [ ] Write transport tests that inspect multipart fields: Keep omits `config_schema`, Replace includes serialized values verbatim, and Clear includes the literal JSON array `[]`.
- [ ] Write action tests for unknown mode, invalid schema JSON, and each valid intent. Assert the exact permission is checked before the wrapper is called.
- [ ] Write form tests for an explicit Keep/Replace/Clear control, conditional editor visibility, explanatory copy, and preserved selected mode after a validation error.
- [ ] Run the three focused test files and confirm RED.
- [ ] Make `configSchema` optional in the wrapper and append the multipart field only when defined.
- [ ] Parse an explicit intent discriminant in the action: Keep maps to `undefined`, Replace maps to validated schema, Clear maps to `[]`.
- [ ] Extract only the schema-intent control from the oversized form if needed to keep the change reviewable; the complete Firmware decomposition belongs to Phase 3.
- [ ] Run focused tests, `npm run typecheck`, `npm run lint`, and `npm test`.
- [ ] Commit: `git add src/lib/api/firmwares* src/app/'(authenticated)'/firmware && git commit -m "fix(frontend): preserve firmware schema intent"`.

## Phase Gate

- [ ] Run `npm run typecheck`, `npm run lint`, `npm test`, and `npm run build`.
- [ ] Run `git diff --check` and inspect `git diff --stat HEAD~7`.
- [ ] Verify every changed path is below `frontend/` from the repository root.
- [ ] Confirm the Phase 1 guarantees in the master plan and record any intentional deviation before starting Phase 2.
