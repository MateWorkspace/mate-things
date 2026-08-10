# Frontend Consistency and Reuse Refactor Design

**Date:** 2026-08-10

**Status:** Approved design

## Purpose

Refactor `mate-things/frontend` into a more consistent, reusable, and
maintainable operations UI without changing its public routes, query parameter
names, permission model, or intended workflows. The work also fixes correctness
issues discovered during the whole-frontend audit.

All implementation changes are confined to `mate-things/frontend`. Existing or
future changes elsewhere in the repository are out of scope.

## Goals

- Make Action History and Node Logs the reference filter-and-table layout.
- Use the same filter panel, fields, and actions for card and table collections.
- Standardize card grids, collection summaries, empty states, pagination, and
  table chrome.
- Replace six duplicated async entity comboboxes with one accessible engine and
  thin resource adapters.
- Separate raw backend transport wrappers from public Server Actions.
- Standardize query parsing, action state, field errors, dialog lifecycle, and
  mutation feedback.
- Remove silent option limits, stale request races, misleading client-side
  filtering, and permission-inconsistent ancillary reads.
- Split oversized feature files into focused units without hiding route-level
  permission and data orchestration.
- Improve rendering and refresh behavior, especially for operational records
  and BLE Direct.

## Non-goals

- No backend API, database, migration, MQTT, or deployment changes.
- No URL or query parameter renaming.
- No visual rebrand or broad workflow redesign.
- No conversion of record-dense tables into cards.
- No generic configuration-driven resource-page framework.
- No relocation of domain validation or authorization authority from the
  backend to the frontend.

## Audit Summary

The audit found four main categories of work.

### Repeated presentation patterns

- Action History, Telemetry, and Node Logs duplicate the same filter panel,
  grid, divider, Apply button, and Clear link.
- Nodes, API Keys, Firmware, and record pages repeat labeled inputs and native
  select styling.
- Action History, Telemetry, Node Logs, and API Keys repeat table frames,
  header/cell classes, row hover behavior, and responsive overflow handling.
- Telemetry and Node Logs duplicate client-side record windowing.
- Resource cards share `ResourceCard` but still duplicate detail rows, footer
  spacing, CTA links, focus behavior, card grids, summaries, and empty states.
- Page containers and metadata grids are repeated throughout detail routes.

### Repeated stateful behavior

- Six async entity comboboxes duplicate debounce, paging, selection, loading,
  outside-click, and listbox markup. Their cancellation cleanup is ineffective,
  so stale requests can replace current results.
- Action-state shapes, API error conversion, inline messages, toast effects,
  refresh behavior, and dialog reset behavior are duplicated under different
  names.
- Many dialogs retain successful state after close, cannot be opened again, or
  remain dismissible while a mutation is pending.
- API-key generation/regeneration dialogs retain show-once secrets after close.

### Data and permission inconsistencies

- Raw API modules use `"use server"`, making transport functions public Server
  Functions instead of server-only wrappers. Client comboboxes call these
  wrappers directly rather than permission-checked selector actions.
- Navigation, proxy protection, and route-access policies are separate and have
  drifted for BLE Direct, Broadcast Sessions, and API Keys.
- Record filters fetch action/node labels without consistently checking the
  corresponding read permission.
- Several selectors silently stop at 48 or 100 options.
- Nodes apply connection filtering after backend pagination, producing
  incorrect counts, pages, and empty states.
- Query scalar, enum, boolean, repeated-value, and out-of-range behavior varies
  across routes.
- Firmware binary replacement does not expose the backend contract's distinct
  keep, replace, and clear schema intentions.

### Maintainability and performance issues

- `FirmwareForm.tsx`, node-class detail, firmware actions, and the BLE client
  mix several responsibilities in large files.
- Dashboard and shared record refresh controls implement similar visibility and
  timing logic differently.
- Telemetry and Node Logs transfer full result sets before client-side slicing.
- BLE log updates replace broad connection state and rerender unrelated panels;
  log rows use unstable index keys.
- Expected 404 handling and unexpected authentication/network/server failures
  are frequently collapsed into the same `null` fallback.
- Local loading/error/not-found boundaries are incomplete on several features.

## Architecture

The refactor uses small compositional primitives. Shared components own
presentation and generic interaction mechanics. Route components retain domain
copy, query names, permissions, data reads, mutations, table columns, card
content, and validation.

### Filter system

Create shared filter primitives under `src/components/filters/`:

- `FilterPanel` owns the canonical warm bordered panel, responsive field grid,
  GET form behavior, hidden `page=1`, preserved values such as `limit`, and the
  Apply/Clear footer.
- `FilterField` associates a visible label and optional hint/error with one
  control and owns field spacing.
- `SearchFilter`, `SelectFilter`, and `CheckboxFilter` provide reusable input
  implementations using the existing UI primitives and theme tokens.
- `FilterActions` is extracted only if a route needs footer composition beyond
  the default Apply/Clear pair.

Action History and Node Logs define the exact visual contract. Telemetry is
migrated next. All filtered card collections then adopt the same panel and
fields. `CollectionToolbar` may continue to provide the mobile sheet, but it
must wrap the same `FilterPanel` content rather than a separate form design.

The filter system preserves every existing public query name and clears filters
to the route's canonical URL while preserving intentional page-size settings.

### Collection system

Extend `src/components/collection/` with:

- `CollectionGrid` for accessible responsive card grids with named density
  variants.
- `CollectionSummary` for consistent visible/total counts.
- `CollectionEmptyState` for distinct filtered and truly-empty messaging,
  including a canonical clear action.
- `ResourceDetailList` and `ResourceDetail` for card metadata.
- `ResourceCardLink` and `ResourceCardActions` for consistent footer spacing,
  touch targets, and focus rings.

Domain cards remain separate components. They compose these primitives rather
than passing domain schemas into a generic resource renderer.

### Table system

Create structural table primitives under `src/components/table/`:

- `TableFrame`
- `TableHeaderCell`
- `TableCell`
- `ExpandableTableRow`
- `RecordWindowControl` or a shared `useRecordWindow` hook

Action History and Node Logs remain the reference table pages. Telemetry and API
Keys adopt the same frame, typography, row focus/hover behavior, and responsive
column priorities. Domain columns and expanded content stay route-local to
avoid render-callback complexity across Server/Client Component boundaries.

Tables remain horizontally scrollable when the information is inherently
tabular. Secondary columns may move into expanded content or hide at narrow
breakpoints, while primary identity/status columns remain visible.

### Async entity selection

Create `src/components/combobox/AsyncEntityCombobox.tsx` and
`src/hooks/use-paginated-options.ts`.

The engine owns:

- debounced search;
- page navigation;
- request cancellation or sequence-based stale-response rejection;
- open/close and outside interaction;
- selected option state;
- loading, empty, and error announcements;
- WAI-ARIA combobox/listbox semantics;
- Arrow, Home, End, Enter, and Escape behavior;
- focus restoration; and
- touch targets of at least 44 CSS pixels.

Thin resource adapters supply field name, labels, placeholders, value/label
mapping, empty-option copy, and a dedicated selector Server Action. Node ID and
device-ID selectors may share the same engine while retaining different stored
values.

### Page and detail composition

Create a narrow `PageContainer` and shared metadata presentation primitives.
Do not create a generic resource-detail page. Each route continues to expose its
permission-aware reads, tabs, relationships, action buttons, and error mapping
directly.

Every collection page follows this composition when applicable:

1. `PageContainer`
2. `PageHeader`
3. `FilterPanel`
4. `CollectionSummary`
5. card grid or data table
6. filtered or truly-empty state
7. pagination or bounded record window

Broadcast Sessions has no filters and omits step 3. Access Control retains its
tabs but adopts shared collection/card behavior inside each tab.

## Data and Permission Boundaries

### Raw transport versus Server Actions

- Raw modules in `src/lib/api/` import `server-only` and never use
  `"use server"` merely to make client invocation possible.
- Public Server Actions live in explicit feature/shared action modules and
  recheck the exact permission before accepting direct-call input.
- Entity selectors call dedicated actions such as `searchNodesAction` and
  `searchRolesAction`; they never import raw transport wrappers into a Client
  Component.

### Route policy

Define one server-safe `ROUTE_POLICIES` registry containing protected path,
navigation group/label, and required permission alternatives. Derive navigation
visibility and proxy protected-prefix matching from it where runtime boundaries
allow. Page guards and mutation guards remain authoritative and explicit.

The registry must include Dashboard, Fleet, Operations, Observability,
Administration, BLE Direct, Broadcast Sessions, and API Keys.

### Permission-aware ancillary reads

Filter label resolution and selector rendering occur only when the user owns
the resource read permission. Without it, the page preserves a raw selected ID
as non-searchable context or omits that filter. Only an expected not-found
response may degrade to a missing label. Authentication, authorization,
network, and server failures reach the appropriate route boundary.

### Query normalization

Add shared query utilities for:

- first scalar value;
- positive safe page/limit values;
- typed enums;
- booleans;
- ISO local/UTC time conversion;
- canonical filtered URLs; and
- out-of-range redirects that preserve active filters.

Repeated query parameters follow one policy: the first scalar value wins.
Existing parameter names remain stable for bookmarks and links.

### Complete option sets

Replace magic `48` and `100` limits with a tested generic all-page collector
for bounded sets or the paginated combobox for large sets. The collector
deduplicates by ID, validates response progress, and terminates safely on an
empty page, no new IDs, or the validated total page count.

### Nodes connection filter

The frontend must not filter connection status after backend pagination. The
current frontend API contract has no server-side connection query, so remove
the Nodes connection filter and its filtered Dashboard links in this refactor.
They may return only through a separately scoped backend/API capability. No
backend change is part of this refactor.

### Firmware schema replacement

The binary replacement UI exposes three explicit modes:

- **Keep:** omit `config_schema` from multipart data.
- **Replace:** send the edited non-empty array.
- **Clear:** send an explicit empty array.

The API wrapper accepts an optional schema value and preserves this distinction.

## Forms, Actions, and Dialogs

Create a shared generic action state under `src/lib/forms/`:

```ts
type ActionStatus = "idle" | "success" | "error";

interface ActionState<Field extends string = string> {
  status: ActionStatus;
  title?: string;
  message?: string;
  fieldErrors?: Partial<Record<Field, string>>;
}
```

Feature states extend it only for real additional data, such as a newly issued
API-key secret or scoped-deletion count.

Shared helpers cover safe form text extraction, optional trimming, JSON object
parsing, optional datetime parsing, permission-denied state, and normalized API
failures. Domain validation remains feature-local.

Create UI primitives for field errors and action messages, plus hooks for:

- action feedback/toasts;
- refresh after success or partial failure;
- first-invalid-field focus; and
- generation-key dialog reset.

The shared `Dialog` gains a dismissal contract. When `dismissible` is false,
Escape, overlay click, native cancel, and the close button are all suppressed.
Mutation dialogs pass `dismissible={!pending}`.

Reopening a dialog mounts a fresh form/action-state instance. API-key secret
results are unmounted and cleared on close so they cannot be re-viewed. Partial
assignment failures refresh authoritative data instead of claiming a reload
that did not occur.

Typed destructive confirmation remains a UX interlock unless the backend
supports authoritative expected-name validation. The refactor must not add
read-permission dependencies that break legitimate remove-only roles.

The payload-schema update action is always an update: it rejects a missing ID
and always requires `payload_schema:set`; it may not fall through to creation
based on client input.

## Refresh and Runtime Optimization

- Unify Dashboard and record-page refresh controls around the existing smart
  refresh behavior.
- Use `startTransition` to expose real refresh pending state.
- Pause automatic refresh while the document is hidden.
- Update the displayed refresh timestamp only from new server props, not
  immediately when `router.refresh()` is called.
- Keep operational queries bounded by time and/or server pagination. Client
  record windows are rendering aids, not a substitute for bounded transfer.
- Remove the authenticated-shell role-name read waterfall when the role label
  is not already available without another blocking request.

## BLE Direct Decomposition

Split BLE responsibilities into focused modules:

- GATT transport;
- protocol codec;
- connection state machine/reconnect policy;
- state store/subscriptions; and
- UI hooks.

Log notifications update only the log slice so System Info, Wi-Fi, and Settings
do not rerender for each line. Log entries receive monotonic stable IDs rather
than array-index keys. The existing bounded log capacity remains.

Normalize BLE directories/files to the documented kebab-case convention after
imports and tests protect the move.

## Route Boundaries and Error Handling

Add meaningful feature-local loading/error/not-found states where currently
missing, including BLE Direct, Broadcast Sessions, API Keys, Node Classes, and
Node detail as applicable.

Expected absence maps to `notFound()` or an explicit unavailable state. Broad
`.catch(() => null)` handling is removed where it masks authorization,
connectivity, or backend failures.

## File Decomposition

- Split firmware schema fields, upload, metadata edit, binary replacement, and
  deletion into separate focused components.
- Move node-class relationship sections out of the route page into colocated
  components.
- Split large action modules only by stable responsibility: parsing/state
  conversion versus mutation orchestration.
- Split BLE transport, connection, codec, and store as described above.
- Remove confirmed dead route helpers and unused assets only after import and
  behavior checks.

## Testing Strategy

### Characterization first

Before extraction, protect existing routes, query names, permissions, result
counts, empty states, and pagination. These tests distinguish intended behavior
from defects explicitly changed by this design.

### Shared component tests

- FilterPanel GET behavior, reset-to-page-one, preservation, clear URL,
  responsive sheet reuse, labels, and keyboard focus.
- Query parser scalar/enum/boolean/date and canonical redirect behavior.
- Collection summaries, filtered/truly-empty states, and accessible card-grid
  labels.
- Table frame, expansion, responsive priorities, focus styles, and record
  window behavior.
- Dialog pending dismissal, generation reset, focus restoration, and secret
  disposal.
- Action state, error conversion, field messages, and invalid-field focus.

### Combobox tests

Cover debounce, stale response suppression, abort/unmount, paging, selection,
clear selection, loading/error announcements, complete keyboard interaction,
ARIA relationships, and 44-pixel controls.

### Route and contract tests

- Raw API wrapper/server-only boundary.
- Exact permission checks in selector and mutation actions.
- Suppression of unauthorized ancillary reads.
- Complete option loading and safe all-page termination.
- Nodes connection-filter and misleading Dashboard-link removal.
- Firmware Keep/Replace/Clear transport behavior.
- Canonical out-of-range redirects and stable existing query names.
- Local error/not-found boundaries.
- API-key show-once secret lifecycle.

### Browser tests

Use Playwright for desktop/mobile filters, table expansion, dialog focus and
pending dismissal, navigation/session refresh, API-key show-once behavior, and
BLE unsupported-state UX.

## Implementation Sequence

1. Restore a clean dependency and typecheck/lint/test/build baseline.
2. Fix raw API versus Server Action boundaries, route policy drift,
   permission-aware ancillary reads, payload-schema update semantics, API-key
   secret lifecycle, and misleading connection filtering.
3. Add shared query, all-page, action-state, form, feedback, and dialog
   foundations.
4. Build shared filters from Action History and Node Logs; migrate Telemetry,
   then filtered card/table routes without changing query names.
5. Build and migrate the accessible async combobox engine and resource
   selector actions.
6. Consolidate table structure, expansion, record windowing, and responsive
   priorities.
7. Consolidate collection grids, summaries, empty states, card metadata, and
   card actions.
8. Remove option caps and normalize all collection query/redirect behavior.
9. Implement firmware schema intent and decompose oversized Firmware and Node
   Class files.
10. Unify refresh behavior, optimize/decompose BLE Direct, add route
    boundaries, remove confirmed dead code, and normalize names.
11. Run complete frontend verification and browser workflows.

Each stage must remain independently reviewable, type-safe, and buildable.

## Success Criteria

- Every filtered card and table collection uses the same filter primitives and
  visual contract.
- Action History and Node Logs retain their intended dense table workflows.
- Shared card/table/filter/combobox mechanics have one implementation each.
- No raw mutation wrapper is exposed as a public Server Action.
- Navigation/proxy route policy covers every authenticated route.
- Ancillary reads obey exact permissions and unexpected failures remain
  visible.
- No selector silently truncates resources at a magic limit.
- No stale combobox response can replace newer results; keyboard and screen
  reader behavior follows the combobox contract.
- Mutation dialogs reset correctly, block dismissal while pending, retain
  error input, and dispose show-once secrets.
- Existing public routes and query parameter names continue to work.
- No code or documentation outside `mate-things/frontend` is changed.
- Frontend typecheck, lint, unit tests, production build, and scoped browser
  tests pass.
