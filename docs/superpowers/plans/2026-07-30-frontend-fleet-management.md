# Frontend Fleet Management Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Deliver card-based Nodes, Node Classes, Firmware, node configuration, and OTA workflows.

**Architecture:** Server pages parse URL search parameters and call typed API wrappers for initial data. Reusable collection/card primitives handle filters and pagination; complex node work lives in a dedicated tabbed detail route with small permission-gated client forms.

**Tech Stack:** Next.js 16 App Router, React 19, TypeScript, Tailwind CSS v4, Vitest, Testing Library, Playwright.

## Global Constraints

- Complete the foundation plan first.
- Nodes self-register over MQTT; never render Create Node.
- Node, class, and firmware reads happen in Server Components.
- Mutations use Server Actions and exact `*:add`, `*:set`, `*:remove`, or
  `ota:dispatch` checks.
- Cards, filters, and pagination follow the approved design system.

---

### Task 1: Collection query, pagination, and card infrastructure

**Files:**
- Create: `frontend/src/lib/collection-query.ts`
- Create: `frontend/src/lib/collection-query.test.ts`
- Create: `frontend/src/components/collection/CollectionToolbar.tsx`
- Create: `frontend/src/components/collection/Pagination.tsx`
- Create: `frontend/src/components/collection/Pagination.test.tsx`
- Create: `frontend/src/components/collection/ResourceCard.tsx`
- Create: `frontend/src/components/collection/FilterDrawer.tsx`

**Interfaces:**
- Produces: `parsePageQuery`, `CollectionToolbar`, `Pagination`,
  `ResourceCard`, and `FilterDrawer`.
- Consumes: `PageResponse` from `src/lib/api/types.ts`.

- [ ] **Step 1: Write failing query and pagination tests**

```ts
it("normalizes invalid page values", () => {
  expect(parsePageQuery({ page: "-2", limit: "999", search: " sensor " }))
    .toEqual({ page: 1, limit: 24, search: "sensor" });
});
```

```tsx
it("preserves filters while moving to the next page", () => {
  render(
    <Pagination
      page={{ page: 2, limit: 12, total_items: 40 }}
      searchParams={{ search: "sensor", node_class_id: "cold" }}
    />,
  );
  expect(screen.getByRole("link", { name: "Next page" })).toHaveAttribute(
    "href",
    expect.stringContaining("page=3"),
  );
});
```

- [ ] **Step 2: Verify failure**

Run:

```bash
npm test -- src/lib/collection-query.test.ts src/components/collection/Pagination.test.tsx
```

Expected: fail because the modules do not exist.

- [ ] **Step 3: Implement normalized URL parsing**

```ts
const ALLOWED_LIMITS = [12, 24, 48] as const;

export function parsePageQuery(
  raw: Record<string, string | string[] | undefined>,
): Required<Pick<PageQuery, "page" | "limit">> & Pick<PageQuery, "search"> {
  const page = Math.max(1, Number(raw.page) || 1);
  const candidate = Number(raw.limit) || 12;
  const limit = ALLOWED_LIMITS.includes(candidate as 12 | 24 | 48)
    ? candidate
    : 24;
  const search = String(raw.search ?? "").trim() || undefined;
  return { page, limit, search };
}
```

- [ ] **Step 4: Implement shared components**

`CollectionToolbar` renders desktop filters and opens `FilterDrawer` on small
screens. `Pagination` builds links by merging parameters through
`URLSearchParams`, exposes Previous/Next and page position, and disables
impossible navigation semantically.

- [ ] **Step 5: Run tests and commit**

```bash
npm test -- src/lib/collection-query.test.ts src/components/collection
npm run typecheck
npm run lint
git add frontend/src/lib/collection-query* frontend/src/components/collection
git commit -m "feat(frontend): add card collection infrastructure"
```

### Task 2: Nodes card collection

**Files:**
- Create: `frontend/src/app/(authenticated)/nodes/page.tsx`
- Create: `frontend/src/app/(authenticated)/nodes/loading.tsx`
- Create: `frontend/src/app/(authenticated)/nodes/error.tsx`
- Create: `frontend/src/app/(authenticated)/nodes/_components/NodeCard.tsx`
- Create: `frontend/src/app/(authenticated)/nodes/_components/NodeFilters.tsx`
- Create: `frontend/src/app/(authenticated)/nodes/_components/NodeCard.test.tsx`

**Interfaces:**
- Consumes: `listNodes`, `listNodeClasses`, `listFirmwares`, collection
  infrastructure, and effective permissions.
- Produces: `/nodes` and `NodeCard`.

- [ ] **Step 1: Write the failing card test**

```tsx
it("shows identity and status without a create action", () => {
  render(<NodeCard node={DISCONNECTED_NODE} permissions={["node:get"]} />);
  expect(screen.getByText("Cold Storage Sensor 07")).toBeVisible();
  expect(screen.getByText("Disconnected")).toBeVisible();
  expect(screen.getByText("AC276E5E030C")).toBeVisible();
  expect(screen.queryByRole("button", { name: /create node/i }))
    .not.toBeInTheDocument();
});
```

- [ ] **Step 2: Verify failure**

```bash
npm test -- src/app/'(authenticated)'/nodes/_components/NodeCard.test.tsx
```

Expected: missing component failure.

- [ ] **Step 3: Implement server page data flow**

Parse `page`, `limit`, `search`, `node_class_id`, and `firmware_id`. Fetch the
node page and permitted filter options in parallel:

```ts
const [nodes, classes, firmwares] = await Promise.all([
  listNodes(query),
  canReadClasses ? listNodeClasses({ limit: 48 }) : null,
  canReadFirmware ? listFirmwares({ limit: 48 }) : null,
]);
```

Connection state filtering must only be exposed when the backend supports it;
otherwise provide a client-visible connected/disconnected quick filter over
the current page and label it “This page”.

- [ ] **Step 4: Implement the responsive card grid**

Cards link to `/nodes/${node.id}` and show status, device ID, class/firmware
labels when resolved, description, and audit freshness. Edit and OTA affordances
appear only for `node:set` and `ota:dispatch`.

- [ ] **Step 5: Verify and commit**

```bash
npm test -- src/app/'(authenticated)'/nodes
npm run typecheck
npm run lint
git add frontend/src/app/'(authenticated)'/nodes
git commit -m "feat(frontend): add paginated node cards"
```

### Task 3: Node configuration API and detail workspace

**Files:**
- Create: `frontend/src/lib/api/node-config.ts`
- Modify: `frontend/src/lib/api/firmwares.ts`
- Modify: `frontend/src/lib/api/index.ts`
- Create: `frontend/src/app/(authenticated)/nodes/[id]/page.tsx`
- Create: `frontend/src/app/(authenticated)/nodes/[id]/loading.tsx`
- Create: `frontend/src/app/(authenticated)/nodes/[id]/not-found.tsx`
- Create: `frontend/src/app/(authenticated)/nodes/[id]/_components/NodeTabs.tsx`
- Create: `frontend/src/app/(authenticated)/nodes/[id]/_components/NodeOverview.tsx`
- Create: `frontend/src/app/(authenticated)/nodes/[id]/_components/NodeConfigForm.tsx`
- Create: `frontend/src/app/(authenticated)/nodes/[id]/_components/NodeEditForm.tsx`
- Create: `frontend/src/app/(authenticated)/nodes/[id]/_components/DeleteNodeDialog.tsx`
- Create: `frontend/src/app/(authenticated)/nodes/[id]/_lib/actions.ts`
- Test: `frontend/src/lib/api/node-config.test.ts`
- Test: `frontend/src/app/(authenticated)/nodes/[id]/_components/NodeTabs.test.tsx`

**Interfaces:**
- Produces:

```ts
export interface FirmwareConfigParameterResponse {
  key: string;
  value_type: string;
}

export interface NodeConfigValueResponse {
  key: string;
  value: string;
  updated_at?: string;
}

export function getNodeConfig(id: string): Promise<NodeConfigValueResponse[]>;
export function setNodeConfig(
  id: string,
  key: string,
  value: string,
): Promise<void>;
export function getFirmwareConfigParameters(
  id: string,
): Promise<FirmwareConfigParameterResponse[]>;
```

- [ ] **Step 1: Write API path tests**

Mock `apiFetch` and assert:

```ts
expect(apiFetch).toHaveBeenCalledWith("/nodes/node-1/config");
expect(apiFetch).toHaveBeenCalledWith("/nodes/node-1/config", {
  method: "PUT",
  body: { key: "sample_rate", value: 30 },
});
```

- [ ] **Step 2: Verify failure**

```bash
npm test -- src/lib/api/node-config.test.ts
```

Expected: missing wrapper failure.

- [ ] **Step 3: Implement wrappers from backend response contracts**

Use the exact `key`, `value_type`, `value`, and `updated_at` JSON keys from
`backend/internal/presentation/http/response/config.go`. Send string values
through `SetNodeConfigValueRequest` and export wrappers from
`src/lib/api/index.ts`.

- [ ] **Step 4: Implement URL-addressable node tabs**

Use `?tab=overview|configuration|firmware|actions|telemetry|logs`. Invalid
values normalize to `overview`. Fetch only the active tab’s expensive data.
`NodeConfigForm` maps firmware parameter types to native inputs and submits one
key/value mutation at a time through a permission-checking Server Action.
`NodeEditForm` checks `node:set` before calling `updateNode`. The delete dialog
checks `node:remove`, requires the exact device ID, calls `deleteNode`, and
redirects to `/nodes` only after confirmed backend success.

- [ ] **Step 5: Verify and commit**

```bash
npm test -- src/lib/api/node-config.test.ts src/app/'(authenticated)'/nodes/'[id]'
npm run typecheck
npm run lint
git add frontend/src/lib/api frontend/src/app/'(authenticated)'/nodes/'[id]'
git commit -m "feat(frontend): add node operations workspace"
```

### Task 4: Node Classes CRUD

**Files:**
- Create: `frontend/src/app/(authenticated)/node-classes/page.tsx`
- Create: `frontend/src/app/(authenticated)/node-classes/[id]/page.tsx`
- Create: `frontend/src/app/(authenticated)/node-classes/_components/NodeClassCard.tsx`
- Create: `frontend/src/app/(authenticated)/node-classes/_components/NodeClassForm.tsx`
- Create: `frontend/src/app/(authenticated)/node-classes/_lib/actions.ts`
- Test: `frontend/src/app/(authenticated)/node-classes/_components/NodeClassCard.test.tsx`
- Test: `frontend/src/app/(authenticated)/node-classes/_lib/actions.test.ts`

**Interfaces:**
- Consumes: node-class API wrappers, `ResourceCard`, `Pagination`, session
  permissions.
- Produces: list/detail/create/edit/delete node-class workflows.

- [ ] **Step 1: Write permission and validation tests**

```ts
it("rejects creation without node_class:add", async () => {
  mockPermissions(["node_class:get"]);
  const state = await createNodeClassAction(
    EMPTY_STATE,
    formData({ name: "Cold Storage", description: "Temperature nodes" }),
  );
  expect(state.status).toBe("error");
  expect(createNodeClass).not.toHaveBeenCalled();
});
```

- [ ] **Step 2: Verify failure**

```bash
npm test -- src/app/'(authenticated)'/node-classes
```

- [ ] **Step 3: Implement cards, forms, and actions**

Use modal create/edit forms. Validate non-empty trimmed name, call existing
typed wrappers, and return `FormActionState`. Delete confirmation includes the
class name and explains that the backend can reject dependent resources.

- [ ] **Step 4: Implement class detail relationships**

Fetch class, its firmware page, filtered nodes, and filtered actions in
parallel when permissions allow. Render each relationship as compact cards
linking to its global/detail route.

- [ ] **Step 5: Verify and commit**

```bash
npm test -- src/app/'(authenticated)'/node-classes
npm run typecheck
npm run lint
git add frontend/src/app/'(authenticated)'/node-classes
git commit -m "feat(frontend): add node class management"
```

### Task 5: Firmware binaries and OTA

**Files:**
- Create: `frontend/src/app/(authenticated)/firmware/page.tsx`
- Create: `frontend/src/app/(authenticated)/firmware/[id]/page.tsx`
- Create: `frontend/src/app/(authenticated)/firmware/_components/FirmwareCard.tsx`
- Create: `frontend/src/app/(authenticated)/firmware/_components/FirmwareForm.tsx`
- Create: `frontend/src/app/(authenticated)/firmware/_components/OtaDialog.tsx`
- Create: `frontend/src/app/(authenticated)/firmware/_lib/actions.ts`
- Test: `frontend/src/app/(authenticated)/firmware/_lib/actions.test.ts`
- Test: `frontend/src/app/(authenticated)/firmware/_components/OtaDialog.test.tsx`

**Interfaces:**
- Consumes: firmware and OTA wrappers, compatible-firmware endpoint, node
  context, and exact permissions.
- Produces: firmware upload/download/replace/edit/delete and node-scoped OTA.

Extend the firmware wrapper with:

```ts
export interface FirmwareConfigSchemaItem {
  key: string;
  value_type: string;
}

export function createFirmware(
  nodeClassId: string,
  name: string,
  file: File | Blob,
  configSchema: FirmwareConfigSchemaItem[],
): Promise<FirmwareCreateResponse>;

export function replaceFirmwareBinary(
  id: string,
  file: File | Blob,
  configSchema: FirmwareConfigSchemaItem[],
): Promise<FirmwareBinaryStatResponse>;
```

- [ ] **Step 1: Write mutation tests**

```ts
it("does not dispatch incompatible firmware", async () => {
  mockAvailableFirmwares([]);
  const result = await dispatchOtaAction({
    nodeId: "node-1",
    firmwareId: "firmware-2",
  });
  expect(result.message).toMatch(/not available/i);
  expect(dispatchOtaByNodeId).not.toHaveBeenCalled();
});
```

- [ ] **Step 2: Verify failure**

```bash
npm test -- src/app/'(authenticated)'/firmware
```

- [ ] **Step 3: Implement firmware collection and details**

Upload uses `FormData` with `node_class_id`, `name`, `file`, and a JSON-encoded
`config_schema` array of `{ key, value_type }`. Binary replacement sends
`file` and the updated `config_schema`. Details expose checksum, size, binary
availability, config parameters, class, preferences, audit fields, and
download using `getFirmwareBinaryUrlById`.

- [ ] **Step 4: Implement node-scoped OTA confirmation**

Select only from `listAvailableFirmwaresByNodeId`. Resolve the backend download
URL server-side, present node and firmware identity, require confirmation, and
call `dispatchOtaByNodeId`. Never expose an OTA button without
`ota:dispatch`.

- [ ] **Step 5: Complete phase verification and commit**

```bash
npm test
npm run typecheck
npm run lint
npm run build
git add frontend/src/app/'(authenticated)'/firmware frontend/src/app/'(authenticated)'/nodes/'[id]'
git commit -m "feat(frontend): add firmware and OTA workflows"
```
