# Frontend Shared Collection Systems Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace duplicated filter, combobox, table, card, and page chrome with small compositional primitives while preserving every route, query parameter, display mode, and domain-specific interaction.

**Architecture:** Shared components own layout, accessibility, and interaction mechanics; route components retain domain labels, option sources, query names, permissions, columns, and card content. Action History and Node Logs define the canonical filter and table experience, and the same filter shell wraps card collections.

**Tech Stack:** Next.js 16, React 19, TypeScript generics, Tailwind CSS v4, Vitest, Testing Library, user-event.

## Global Constraints

- Start from the reviewed Phase 1 commit.
- Work only below `mate-things/frontend`; preserve routes and URL query keys exactly.
- Do not create a generic schema-driven “resource page.” Prefer compositional children and typed adapters.
- Keep nodes, node classes, firmware, actions, users, payload schemas, and broadcasts as cards.
- Keep action history, node logs, telemetry, and API keys as dense tables where their existing responsive behavior permits it.
- Every interactive control must retain visible labels, keyboard operation, focus visibility, and 44px minimum coarse-pointer targets.

---

## Task 1: Build the canonical filter system

**Files:**

- Create: `src/components/filters/FilterPanel.tsx`
- Create: `src/components/filters/FilterField.tsx`
- Create: `src/components/filters/SearchFilter.tsx`
- Create: `src/components/filters/SelectFilter.tsx`
- Create: `src/components/filters/CheckboxFilter.tsx`
- Create: `src/components/filters/filter-controls.test.tsx`
- Modify: `src/components/collection/FilterDrawer.tsx`
- Modify: `src/components/collection/CollectionToolbar.tsx`

**Interfaces:**

```tsx
type FilterPanelProps = {
  action?: string;
  children: React.ReactNode;
  activeCount?: number;
  submitLabel?: string;
  resetHref?: string;
  hiddenValues?: Readonly<
    Record<string, string | readonly string[] | undefined>
  >;
};

type FilterFieldProps = {
  label: string;
  htmlFor: string;
  hint?: string;
  children: React.ReactNode;
};
```

**Test shape:**

```tsx
it("submits preserved hidden values and resets page", async () => {
  render(
    <FilterPanel
      hiddenValues={{ sort: "name", tag: ["a", "b"] }}
      resetHref="/nodes"
    >
      <SearchFilter name="search" label="Search" defaultValue="lab" />
    </FilterPanel>,
  );
  expect(screen.getByDisplayValue("name")).toHaveAttribute("name", "sort");
  expect(screen.getByDisplayValue("1")).toHaveAttribute("name", "page");
});
```

**Implementation shape:**

```tsx
export function FilterField({
  label,
  htmlFor,
  hint,
  children,
}: FilterFieldProps) {
  return (
    <div className="space-y-2">
      <Label htmlFor={htmlFor}>{label}</Label>
      {children}
      {hint ? (
        <p id={`${htmlFor}-hint`} className="text-muted-foreground text-sm">
          {hint}
        </p>
      ) : null}
    </div>
  );
}
```

- [ ] Characterize Action History and Node Logs with tests for desktop layout, mobile drawer, GET submission, active-count badge, Apply, Clear, hidden values, Enter submission, labels, and focus order.
- [ ] Add failing primitive tests for search, select, checkbox, hidden multi-values, custom children, and no-JavaScript GET behavior.
- [ ] Run `npm test -- src/components/filters/filter-controls.test.tsx` and confirm RED.
- [ ] Implement `FilterPanel` as a server-compatible composition by default. Isolate only drawer toggling in a small Client Component; do not turn the entire filter form into a Client Component.
- [ ] Implement fields using `src/components/ui/input.tsx`, `label.tsx`, and existing button styles. Forward native attributes and refs; do not create parallel styling props that conflict with the UI layer.
- [ ] Refactor `FilterDrawer`/`CollectionToolbar` to compose the canonical panel or remove superseded code after `rg` proves no callers remain.
- [ ] Run focused tests, typecheck, and lint.
- [ ] Commit: `git add src/components/filters src/components/collection && git commit -m "refactor(frontend): add shared filter system"`.

## Task 2: Migrate the reference record filters

**Files:**

- Modify: `src/app/(authenticated)/action-history/_components/ActionLogFilters.tsx`
- Modify: `src/app/(authenticated)/action-history/_lib/filters.ts`
- Modify: `src/app/(authenticated)/action-history/page.tsx`
- Modify: `src/app/(authenticated)/node-logs/_components/NodeLogFilters.tsx`
- Modify: `src/app/(authenticated)/node-logs/page.tsx`
- Modify: `src/app/(authenticated)/telemetry/_components/TelemetryFilters.tsx`
- Modify: `src/app/(authenticated)/telemetry/page.tsx`
- Create: `src/app/(authenticated)/action-history/_components/ActionLogFilters.test.tsx`
- Create: `src/app/(authenticated)/node-logs/_components/NodeLogFilters.test.tsx`
- Create: `src/app/(authenticated)/telemetry/_components/TelemetryFilters.test.tsx`

**Test shape:**

```tsx
it.each([
  [ActionLogFilters, ["action_id", "node_id", "status", "from", "to"]],
  [NodeLogFilters, ["node_id", "level", "search", "from", "to"]],
  [TelemetryFilters, ["node_id", "from", "to"]],
])("preserves canonical GET fields", (Filters, names) => {
  render(<Filters {...fixtureProps} />);
  for (const name of names)
    expect(document.querySelector(`[name="${name}"]`)).not.toBeNull();
});
```

**Implementation shape:**

```tsx
<FilterPanel
  action="/node-logs"
  resetHref="/node-logs"
  activeCount={activeCount}
>
  <SearchFilter name="search" label="Message" defaultValue={filters.search} />
  <TimeRangeFilter from={filters.from} to={filters.to} />
</FilterPanel>
```

- [ ] Write characterization assertions for every current query key and default: action/node IDs, device ID, severity/status, start/end time, search/window, and page reset behavior. Treat the rendered Action History/Node Logs layout as the visual contract.
- [ ] Run the three focused tests before refactoring and save the passing baseline.
- [ ] Replace duplicated form shell, grid, labels, search/select/checkbox fields, Apply, and Clear markup with shared filter primitives. Keep permission-aware selector-vs-raw-ID behavior from Phase 1.
- [ ] Move Action History's duplicate first-value/time/status parsing onto `src/lib/query.ts` and `src/lib/record-filters.ts` without renaming its URL keys.
- [ ] Ensure applying any filter submits hidden `page=1`, while Clear navigates to the same collection pathname and preserves only intentional page-size settings.
- [ ] Rerun focused tests and inspect the three routes at desktop and mobile widths.
- [ ] Commit: `git add src/app/'(authenticated)'/{action-history,node-logs,telemetry} && git commit -m "refactor(frontend): unify record filters"`.

## Task 3: Migrate every card and administrative filter

**Files:**

- Modify: `src/app/(authenticated)/nodes/_components/NodeFilters.tsx`
- Modify: `src/app/(authenticated)/node-classes/page.tsx`
- Modify: `src/app/(authenticated)/firmware/page.tsx`
- Modify: `src/app/(authenticated)/actions/page.tsx`
- Modify: `src/app/(authenticated)/admin/users/page.tsx`
- Modify: `src/app/(authenticated)/admin/payload-schemas/page.tsx`
- Modify: `src/app/(authenticated)/admin/api-keys/_components/ApiKeyFilters.tsx`
- Create: `src/components/filters/filter-route-contracts.test.tsx`

**Test shape:**

```ts
const contracts = [
  { route: "/nodes", names: ["search", "node_class_id"] },
  { route: "/firmware", names: ["search", "node_class_id"] },
  { route: "/admin/api-keys", names: ["search", "status"] },
] as const;

it.each(contracts)("keeps $route query contract", async ({ route, names }) => {
  const view = await renderRoute(route);
  for (const name of names)
    expect(view.container.querySelector(`[name="${name}"]`)).not.toBeNull();
});
```

**Implementation shape:**

```tsx
<FilterPanel action="/firmware" resetHref="/firmware" activeCount={activeCount}>
  <SearchFilter name="search" label="Search firmware" defaultValue={search} />
  <NodeClassSearchCombobox name="node_class_id" defaultValue={nodeClassId} />
</FilterPanel>
```

- [ ] Build a route contract table in the test containing the exact pathname, query keys, default values, reset href, and field labels for all seven collections.
- [ ] Add failing assertions that each route renders the same canonical panel/chrome used by the record routes, even when its results are cards.
- [ ] Run `npm test -- src/components/filters/filter-route-contracts.test.tsx` and confirm RED.
- [ ] Migrate each route to `FilterPanel` and typed field components. Keep domain-specific option lists and permissions in the route component.
- [ ] Use a consistent field order: primary search first, domain selectors second, state flags third, Apply/Clear last. Preserve URL keys and server-side filtering behavior.
- [ ] Remove duplicate label/select/input class strings only after all routes are migrated and `rg` finds no use.
- [ ] Run route-contract tests, typecheck, lint, and a production build.
- [ ] Commit: `git add src/app/'(authenticated)' src/components/filters && git commit -m "refactor(frontend): standardize collection filters"`.

## Task 4: Replace six selector copies with one accessible async combobox

**Files:**

- Create: `src/components/combobox/AsyncEntityCombobox.tsx`
- Create: `src/hooks/use-paginated-options.ts`
- Create: `src/hooks/use-paginated-options.test.tsx`
- Create: `src/components/combobox/async-entity-combobox.test.tsx`
- Modify: `src/components/actions/ActionSearchCombobox.tsx`
- Modify: `src/components/node-classes/NodeClassSearchCombobox.tsx`
- Modify: `src/components/nodes/NodeSearchCombobox.tsx`
- Modify: `src/components/nodes/NodeDeviceIdSearchCombobox.tsx`
- Modify: `src/components/roles/RoleSearchCombobox.tsx`
- Modify: `src/components/users/UserSearchCombobox.tsx`

**Interfaces:**

```tsx
type LoadOptions = (request: {
  query: string;
  page: number;
}) => Promise<SearchOptionsPage>;

type AsyncEntityComboboxProps = {
  name: string;
  label: string;
  defaultValue?: string;
  defaultLabel?: string;
  placeholder?: string;
  loadOptions: LoadOptions;
  emptyMessage: string;
  disabled?: boolean;
};
```

**Test shape:**

```tsx
it("ignores a stale slower response", async () => {
  const requests = deferredOptionRequests();
  render(<AsyncEntityCombobox {...props} loadOptions={requests.load} />);
  await typeQuery("old");
  await typeQuery("new");
  requests.resolve("new", optionPage("New result"));
  requests.resolve("old", optionPage("Old result"));
  expect(
    await screen.findByRole("option", { name: "New result" }),
  ).toBeVisible();
  expect(screen.queryByText("Old result")).toBeNull();
});
```

**Implementation shape:**

```ts
useEffect(() => {
  const sequence = ++requestSequence.current;
  const timer = window.setTimeout(async () => {
    const page = await loadOptions({ query, page: requestedPage });
    if (sequence === requestSequence.current) setResult(page);
  }, debounceMs);
  return () => {
    window.clearTimeout(timer);
    requestSequence.current += 1;
  };
}, [debounceMs, loadOptions, query, requestedPage]);
```

- [ ] Write failing fake-timer tests reproducing the current stale-response bug: query A resolves after query B and must not overwrite B. Also cover unmount, page changes, failed loads, and clearing.
- [ ] Write accessibility tests for combobox/listbox roles, `aria-expanded`, `aria-controls`, active descendant, Arrow Up/Down, Home/End, Enter, Escape, Tab, live loading/result announcements, selected label, and 44px paging controls.
- [ ] Run the focused test and confirm RED.
- [ ] Implement `usePaginatedOptions` with a monotonically increasing request sequence. Because Server Actions do not accept an AbortSignal, ignore any result whose sequence is not current; clear timers in the actual effect cleanup.
- [ ] Implement the headless-enough shared combobox with stable generated IDs and no resource-specific imports.
- [ ] Reduce each existing resource component to a thin adapter supplying its Phase 1 action, label, placeholder, empty message, default option, and hidden input name.
- [ ] Preserve form submission semantics and existing selected IDs. Verify no raw API wrapper is imported by a Client Component.
- [ ] Run focused tests, typecheck, lint, and `npm test`.
- [ ] Commit: `git add src/components/combobox src/hooks/use-paginated-options* src/components/{actions,node-classes,nodes,roles,users} && git commit -m "refactor(frontend): share accessible entity selector"`.

## Task 5: Build and migrate table structure

**Files:**

- Create: `src/components/table/TableFrame.tsx`
- Create: `src/components/table/TableHeaderCell.tsx`
- Create: `src/components/table/TableCell.tsx`
- Create: `src/components/table/ExpandableTableRow.tsx`
- Create: `src/components/table/data-table.test.tsx`
- Modify: `src/app/(authenticated)/action-history/_components/ActionHistoryTable.tsx`
- Modify: `src/app/(authenticated)/node-logs/_components/NodeLogTable.tsx`
- Modify: `src/app/(authenticated)/telemetry/_components/TelemetryTable.tsx`
- Modify: `src/app/(authenticated)/admin/api-keys/_components/ApiKeyTable.tsx`
- Modify: `src/components/records/RecordWindow.tsx`
- Modify: `src/components/records/JsonPayload.tsx`

**Interfaces:**

```tsx
type TableFrameProps = {
  caption: string;
  children: React.ReactNode;
  mobileCards?: React.ReactNode;
};
```

**Test shape:**

```tsx
it("provides semantic table structure inside a bounded scroller", () => {
  render(
    <TableFrame caption="Node logs">
      <thead>
        <tr>
          <TableHeaderCell>Time</TableHeaderCell>
        </tr>
      </thead>
    </TableFrame>,
  );
  expect(screen.getByRole("table", { name: "Node logs" })).toBeVisible();
  expect(screen.getByRole("columnheader", { name: "Time" })).toHaveAttribute(
    "scope",
    "col",
  );
  expect(screen.getByTestId("table-scroll-region")).toHaveClass(
    "overflow-x-auto",
  );
});
```

**Implementation shape:**

```tsx
export function TableFrame({ caption, children }: TableFrameProps) {
  return (
    <div data-testid="table-scroll-region" className="overflow-x-auto">
      <table className="w-full min-w-max border-separate border-spacing-0">
        <caption className="sr-only">{caption}</caption>
        {children}
      </table>
    </div>
  );
}
```

- [ ] Characterize the four tables: semantic captions/headers, timestamp/status formatting, JSON expansion, row actions, empty state, horizontal overflow, and existing mobile representation.
- [ ] Write failing primitive tests for semantic markup, visually hidden caption option, header scope, numeric alignment, overflow container, custom cell content, and mobile-card slot.
- [ ] Run `npm test -- src/components/table/data-table.test.tsx` and confirm RED.
- [ ] Implement structural primitives only. Keep column definitions explicit in each domain table so permission checks and responsive prioritization remain readable.
- [ ] Migrate Action History and Node Logs first; compare them visually to the pre-change reference. Then migrate Telemetry and API Keys.
- [ ] Share record payload/window presentation where truly identical. Do not force API-key action menus into record-specific components.
- [ ] Run focused table/route tests, typecheck, lint, and build.
- [ ] Commit: `git add src/components/table src/components/records src/app/'(authenticated)' && git commit -m "refactor(frontend): share table structure"`.

## Task 6: Build and migrate card collection structure

**Files:**

- Create: `src/components/collection/CollectionGrid.tsx`
- Create: `src/components/collection/CollectionSummary.tsx`
- Create: `src/components/collection/CollectionEmptyState.tsx`
- Create: `src/components/collection/ResourceDetailList.tsx`
- Create: `src/components/collection/ResourceDetail.tsx`
- Create: `src/components/collection/ResourceCardActions.tsx`
- Create: `src/components/collection/ResourceCardLink.tsx`
- Create: `src/components/collection/collection-primitives.test.tsx`
- Modify: `src/components/collection/ResourceCard.tsx`
- Modify: `src/app/(authenticated)/nodes/_components/NodeCard.tsx`
- Modify: `src/app/(authenticated)/nodes/_components/NodeCollectionEmptyState.tsx`
- Modify: `src/app/(authenticated)/node-classes/_components/NodeClassCard.tsx`
- Modify: `src/app/(authenticated)/firmware/_components/FirmwareCard.tsx`
- Modify: `src/app/(authenticated)/actions/_components/ActionCard.tsx`
- Modify: `src/app/(authenticated)/admin/users/_components/UserCard.tsx`
- Modify: `src/app/(authenticated)/admin/payload-schemas/_components/PayloadSchemaCard.tsx`
- Modify: `src/app/(authenticated)/admin/access-control/_components/RoleCard.tsx`
- Modify: `src/app/(authenticated)/admin/access-control/_components/PermissionCard.tsx`
- Modify: `src/app/(authenticated)/broadcast-sessions/_components/BroadcastSessionCard.tsx`

**Test shape:**

```tsx
it("keeps actions outside the stretched card link", () => {
  render(<FixtureCard />);
  expect(screen.getByRole("link", { name: /open node/i })).not.toContainElement(
    screen.getByRole("button", { name: /delete/i }),
  );
});
```

**Implementation shape:**

```tsx
export function CollectionGrid({ children }: { children: React.ReactNode }) {
  return (
    <ul className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">{children}</ul>
  );
}

export function ResourceCardLink(
  props: LinkProps & { children: React.ReactNode },
) {
  return (
    <Link
      className="after:absolute after:inset-0 focus-visible:outline-none"
      {...props}
    />
  );
}
```

- [ ] Write primitive tests for responsive grids, result counts, arbitrary empty-state content, whole-card links without nested interactive controls, separate action areas, focus rings, and card semantics.
- [ ] Add characterization tests for one operational card (Node), one administrative card (User), and one schema card so domain metadata/actions cannot disappear during migration.
- [ ] Run focused tests and confirm primitive tests RED.
- [ ] Implement compositional card/collection primitives. `ResourceCardLink` must use a stretched-link pattern that does not wrap buttons, menus, checkboxes, or dialog triggers.
- [ ] Migrate each card shell while leaving its domain metadata, status badges, permission gates, and actions explicit.
- [ ] Standardize summary/pagination placement across card routes: filter, summary, grid/empty state, pagination.
- [ ] Retire `ResourceCard` only if all uses are migrated; otherwise make it compose the new primitives.
- [ ] Run focused tests, typecheck, lint, and build.
- [ ] Commit: `git add src/components/collection src/app/'(authenticated)' && git commit -m "refactor(frontend): share card collection structure"`.

## Task 7: Standardize page and detail composition

**Files:**

- Create: `src/components/layout/PageContainer.tsx`
- Create: `src/components/layout/DetailSection.tsx`
- Create: `src/components/layout/MetadataGrid.tsx`
- Create: `src/components/layout/page-composition.test.tsx`
- Modify: `src/app/(authenticated)/dashboard/page.tsx`
- Modify: `src/app/(authenticated)/nodes/page.tsx`
- Modify: `src/app/(authenticated)/nodes/[id]/page.tsx`
- Modify: `src/app/(authenticated)/node-classes/page.tsx`
- Modify: `src/app/(authenticated)/node-classes/[id]/page.tsx`
- Modify: `src/app/(authenticated)/firmware/page.tsx`
- Modify: `src/app/(authenticated)/firmware/[id]/page.tsx`
- Modify: `src/app/(authenticated)/actions/page.tsx`
- Modify: `src/app/(authenticated)/actions/[id]/page.tsx`
- Modify: `src/app/(authenticated)/action-history/page.tsx`
- Modify: `src/app/(authenticated)/telemetry/page.tsx`
- Modify: `src/app/(authenticated)/node-logs/page.tsx`
- Modify: `src/app/(authenticated)/broadcast-sessions/page.tsx`
- Modify: `src/app/(authenticated)/admin/users/page.tsx`
- Modify: `src/app/(authenticated)/admin/users/[id]/page.tsx`
- Modify: `src/app/(authenticated)/admin/access-control/page.tsx`
- Modify: `src/app/(authenticated)/admin/payload-schemas/page.tsx`
- Modify: `src/app/(authenticated)/admin/payload-schemas/[id]/page.tsx`
- Modify: `src/app/(authenticated)/admin/api-keys/page.tsx`
- Modify: `src/components/ui/page-header.tsx`
- Modify: `src/components/ui/states.tsx`

**Test shape:**

```tsx
it("renders detail metadata as a labelled definition list", () => {
  render(<MetadataGrid items={[{ label: "Device ID", value: "mate-01" }]} />);
  expect(screen.getByText("Device ID").tagName).toBe("DT");
  expect(screen.getByText("mate-01").tagName).toBe("DD");
});
```

**Implementation shape:**

```tsx
export function MetadataGrid({ items }: MetadataGridProps) {
  return (
    <dl className="grid gap-4 sm:grid-cols-2">
      {items.map(({ label, value }) => (
        <div key={label}>
          <dt>{label}</dt>
          <dd>{value}</dd>
        </div>
      ))}
    </dl>
  );
}
```

- [ ] Write tests for consistent max width/gutters, header title/description/actions, section heading associations, definition-list metadata, loading/error/empty state placement, and responsive wrapping.
- [ ] Run the focused test and confirm RED.
- [ ] Implement layout-only primitives with `children`; keep data fetching and permissions in Server Component routes.
- [ ] Migrate pages in bounded groups: fleet (`nodes`, `node-classes`, `firmware`), operations (`actions`, `action-history`, `telemetry`, `node-logs`, `broadcast-sessions`), then administration.
- [ ] Preserve each route's semantic heading hierarchy and contextual actions. Do not flatten feature-specific tabs/workspaces into a generic detail renderer.
- [ ] Run focused tests, typecheck, lint, `npm test`, and build.
- [ ] Commit: `git add src/components/layout src/components/ui/page-header.tsx src/components/ui/states.tsx src/app/'(authenticated)' && git commit -m "refactor(frontend): standardize page composition"`.

## Phase Gate

- [ ] Run `npm run typecheck`, `npm run lint`, `npm test`, and `npm run build`.
- [ ] Run `npm run format:check` and `git diff --check`.
- [ ] At 375px, 768px, and 1440px, manually compare Nodes card filters and Action History/Node Logs table filters; they must use the same control hierarchy and spacing.
- [ ] Keyboard-test one card collection, each table mode, and each of the six selector adapters.
- [ ] Confirm no public route or query parameter name changed and no modified path is outside `frontend/`.
