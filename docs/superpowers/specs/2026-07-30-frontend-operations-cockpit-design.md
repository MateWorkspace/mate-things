# Mate Things Frontend Operations Cockpit Design

**Date:** 2026-07-30

**Status:** Approved

**Scope:** `mate-things/frontend`

## Summary

Mate Things will become a responsive, permission-aware operations cockpit for
managing ESP32-class IoT fleets. Fleet health and urgent operational issues lead
the experience. Resource collections use cards rather than tables, and the
selected node becomes a focused workspace for configuration, firmware/OTA,
actions, telemetry, and logs.

The design retains the existing warm brown/cream brand, automatic light/dark
themes, Anton/Inter typography, Server Component architecture, and server-only
JWT handling. The frontend remains presentation plus thin typed API access;
business rules stay in the Go backend.

Approved Superdesign project:

- Project: [Mate Things Operations Cockpit](https://superdesign.dev/teams/c3cdfe5c-365f-46b8-ab93-37b3fc20fff7/projects/d9386ec2-8f4e-4649-91d6-5f04b84111d9)
- [Fleet-health dashboard](https://p.superdesign.dev/draft/cfc9a362-39f6-4841-a29c-da35cba6b2f5)
- [Node card collection](https://p.superdesign.dev/draft/0a404f29-2900-44ce-9146-5cea1deda19f)
- [Node detail workspace](https://p.superdesign.dev/draft/05d16694-190f-4d16-a287-4ed937c6aa03)
- [Access Control](https://p.superdesign.dev/draft/688ce067-b131-406e-b630-5cc84f351659)
- [Profile dialog](https://p.superdesign.dev/draft/df694efb-e4c0-49a0-964f-fe9ebaa5fee6)

## Product goals

1. Make urgent fleet problems visible and actionable from the first viewport.
2. Expose every supported backend feature through task-oriented navigation.
3. Ensure users see only navigation and actions their effective permissions
   allow.
4. Make nodes easy to scan through attractive, responsive card pagination.
5. Use the same card-first collection system for firmware, classes, actions,
   users, access-control resources, schemas, telemetry, and logs.
6. Preserve context by putting node-specific work inside one node detail
   workspace.
7. Keep operational views current with controlled smart refresh.
8. Remain fully usable on mobile, tablet, and desktop in light and dark modes.
9. Keep tokens server-only and all backend mutations behind Server Actions.

## Non-goals

- Reimplementing backend validation or business rules in the browser.
- Storing access or refresh tokens in client-readable storage.
- Creating nodes through HTTP. Nodes self-register through MQTT.
- Adding a manual light/dark theme toggle.
- Introducing real-time browser sockets in the first implementation. Smart
  polling is sufficient.
- Creating decorative charts without an operational decision they support.
- Turning preference storage into a standalone user-facing resource page.

## Experience principles

### Operations first

The dashboard answers:

- Is the fleet healthy?
- What requires attention now?
- Which nodes, actions, firmware operations, or logs explain the problem?
- What is the next supported action?

Every summary card must link to a filtered or scoped detail view. Unsupported
claims such as vulnerability scanning must not appear.

### Card-first collections

Cards are the default representation for list and pagination views. A card
shows only decision-relevant information, status, freshness, and a clear next
action. Dense metadata belongs in a detail page, drawer, or expandable region.

Tables are permitted only if a future feature genuinely requires matrix-style
comparison and a card representation would harm comprehension. None of the
currently planned resource pages require a table.

### Progressive operational depth

Global pages support cross-fleet discovery and filtering. A node detail
workspace exposes the same data pre-filtered to the chosen node. Users can move
from fleet signal to node diagnosis without losing context.

## Information architecture

### Public routes

| Route | Purpose |
|---|---|
| `/` | Redirect to `/dashboard` when authenticated, otherwise `/login`. |
| `/login` | Existing branded username/password login. |

### Authenticated routes

| Route | Page |
|---|---|
| `/dashboard` | Fleet-health overview |
| `/nodes` | Node card collection |
| `/nodes/[id]` | Node detail workspace |
| `/node-classes` | Node-class collection |
| `/node-classes/[id]` | Node-class details and related resources |
| `/firmware` | Firmware collection |
| `/firmware/[id]` | Firmware details, binary operations, related nodes |
| `/actions` | Action collection |
| `/actions/[id]` | Action details and dispatch |
| `/action-history` | Action-log records |
| `/telemetry` | Telemetry records |
| `/node-logs` | Node log records |
| `/admin/users` | User administration |
| `/admin/users/[id]` | User details, role, permissions, password reset |
| `/admin/access-control` | Roles and Permissions tabs |
| `/admin/payload-schemas` | Payload-schema collection |
| `/admin/payload-schemas/[id]` | Schema version details/editor |

The authenticated pages should live under an App Router route group sharing the
same protected layout without changing their public URLs.

### Sidebar groups and permissions

| Group | Item | Visibility permission |
|---|---|---|
| Overview | Fleet Overview | Authenticated; individual widgets are permission-aware |
| Fleet | Nodes | `node:get` |
| Fleet | Node Classes | `node_class:get` |
| Fleet | Firmware | `firmware:get` |
| Operations | Actions | `action:get` |
| Operations | Action History | `action_log:get` |
| Observability | Telemetry | `telemetry_record:get` |
| Observability | Node Logs | `node_log:get` |
| Administration | Users | `user:get` |
| Administration | Access Control | Any relevant `role:*`, `permission:*`, or `role_permission:*` read capability |
| Administration | Payload Schemas | `payload_schema:get` |

An empty group is omitted. Mutation buttons are independently gated by their
specific add/set/remove/dispatch permissions. Unauthorized actions are normally
omitted rather than disabled. A disabled action is appropriate only when its
explanation helps the user understand a temporary resource state.

## Authenticated app shell

### App bar

- Sticky at the top of the viewport.
- Left side: accessible burger button, brand mark, and product name where space
  permits.
- Right side: one profile trigger composed of the current profile name and a
  rounded anonymous human icon.
- Clicking either the name or icon opens the same profile dialog.
- Icon-only controls have an accessible name and visible focus treatment.

### Sidebar

- Desktop: expanded labeled navigation by default; burger toggles a persistent
  icon-only collapsed state.
- Tablet/mobile: overlay drawer with backdrop, focus trap, Escape dismissal,
  and focus restoration.
- Active route is communicated by shape, icon, text treatment, and
  `aria-current`, not color alone.
- Collapsed icons provide accessible labels/tooltips.

### Page header

Authenticated pages begin with:

- breadcrumb;
- page title and concise supporting copy;
- result/freshness context where relevant;
- contextual actions permitted to the user.

## Profile dialog

The profile experience stays available without leaving the current workflow.

### Profile tab

- Anonymous human icon, name, username, role, and bio.
- View mode by default.
- Edit mode for name, username, and bio.
- Explicit Edit, Cancel, and Save Changes controls.
- Effective-permission summary.
- Successful saves update the app-bar name.

### Security tab

- Current password.
- New password and confirmation.
- Clear password requirements driven by backend errors rather than duplicated
  frontend policy.
- Save feedback without closing the entire profile dialog unexpectedly.

### Session action

Logout is visually separated from profile mutation actions and clears the
server-managed session cookies before redirecting to login.

On small screens, the dialog becomes a full-width bottom sheet or full-height
panel while retaining dialog semantics.

## Page designs

### Fleet Overview

The first desktop viewport contains:

1. fleet title, current freshness, Pause/Resume, and manual Refresh;
2. operational health summary cards;
3. urgent-attention queue;
4. recent warning/error node logs and useful drill-downs.

Possible widgets, gated by their data permissions:

- connected versus disconnected nodes;
- failed/unresponded action count;
- nodes with firmware attention;
- recent error/warning node logs;
- action and OTA activity;
- links to filtered Nodes, Action History, and Node Logs.

Existing endpoints do not provide a dedicated aggregate dashboard API. Initial
summary values may be composed from available list/filter endpoints. Requests
must be bounded and parallelized. If this becomes inefficient, a backend
aggregate endpoint should be proposed separately rather than simulating
unsupported metrics.

### Nodes

Toolbar:

- URL-backed search;
- filters for connection state, node class, and firmware where supported;
- sort;
- result count;
- smart-refresh controls;
- mobile filter drawer.

Each card shows:

- name;
- connected/disconnected status label and icon;
- device ID;
- node class;
- firmware;
- last relevant update/freshness;
- short description;
- View Node action;
- permitted edit/OTA contextual actions.

Cards paginate from the backend `PageDataResponse`. The URL contains page,
limit, search, and filters. No Create Node action exists.

### Node detail workspace

Header:

- name and connection state;
- device ID, class, firmware, and freshness;
- Edit, Dispatch Action, View Logs, and OTA actions when permitted.

Tabs:

1. **Overview** — identity, health, configuration completeness, firmware,
   recent actions, telemetry highlights, and warning logs.
2. **Configuration** — firmware config parameters and current node values.
3. **Firmware & OTA** — current/compatible firmware and OTA dispatch.
4. **Actions** — compatible actions and dispatch form.
5. **Telemetry** — device-filtered telemetry.
6. **Node Logs** — device-filtered logs.

Tab state is URL-addressable so refresh/back navigation preserves context.
High-impact mutations use confirmation dialogs and suppress background refresh
while the user is editing.

### Node Classes

- Card CRUD collection.
- Name, description, node count when derivable, firmware count when derivable,
  and audit information.
- Detail page links associated nodes, firmware, and actions.
- Deletion confirmation warns that backend dependency constraints may reject
  the operation.

### Firmware

- Card collection filterable by node class.
- Upload form uses multipart data for node class, name, and binary.
- Details show size, checksum, binary status, class, preferences, and audit
  information.
- Download uses the backend redirect/Location flow.
- Replace binary, update metadata, and delete are permission-gated.
- OTA begins from node context because target compatibility is node-specific.

### Actions

- Card CRUD collection filterable by node class and payload schema.
- Details show description, class, referenced schema/version, preferences,
  audit information, and recent results.
- Dispatch chooses a compatible node, collects payload JSON, and optionally
  supplies `executed_at`.
- Payload editing uses the referenced schema to improve the form and catch
  obvious format errors, but backend validation remains authoritative.
- Success links directly to the returned action-log record.

### Action History

- Filter cards/controls for supported time, action, and node filters.
- Status cards visually distinguish `UNEXECUTED`, `UNRESPONDED`, `FAILED`, and
  `SUCCESS`.
- Payload and messages expand in place.
- Filter-scoped deletion previews the active scope and requires confirmation.

### Telemetry

- Filters: time range, node device ID, metric, schema name, schema version.
- Record cards show metric, device, recorded time, schema, and a compact payload
  preview.
- Payload can expand and be copied.
- Lightweight trends appear only for payloads that can be represented honestly
  as numeric series.
- Filter-scoped deletion previews scope and affected count.

### Node Logs

- Filters: time range, node device ID, and level.
- Levels: `NONE`, `ERROR`, `WARN`, `INFO`, `DEBUG`.
- Compact cards show timestamp, level, tag, device, and expandable message.
- Smart refresh is enabled.
- Filter-scoped deletion previews scope and affected count.

The frontend currently lacks `src/lib/api/node-logs.ts`; implementation must
add a typed wrapper matching the backend counted list/delete responses.

### Users

- Card CRUD collection.
- Cards show name, username, bio summary, role, and audit context.
- Details support role assignment, effective-permission inspection, profile
  fields, preferences, administrative password reset, and deletion.
- Additional confirmation protects deletion or mutation of the currently
  authenticated account.

### Access Control

One workspace contains:

- **Roles tab:** role cards, default-role badge, details, mutation, default-role
  assignment, and grouped permission assignment.
- **Permissions tab:** permission cards and CRUD.

Role permissions are grouped by Profile, Fleet, Operations, Observability, and
Administration. Read-only users see assignments without interactive toggles.
There is no separate sidebar page for role-permission join records.

### Payload Schemas

- Versioned schema cards with validity windows.
- Search and validity-at-time filtering.
- Structured JSON editor with formatting, syntax validation, and readable
  preview.
- Latest-version and exact name/version lookup.
- CRUD and contextual preferences.
- Actions link to their referenced schema version.

### Preferences

Preferences are edited from the owning Action, Firmware, Node, Node Class,
Payload Schema, Permission, Role, or User interface. The shared editor accepts
structured key/value JSON and sends it through the generic backend preference
endpoint. It is never a standalone sidebar page.

## Shared collection behavior

### Search, filters, and pagination

- Paginated filters are stored in URL search parameters.
- Changing search, filter, sort, or page size resets the page to one.
- Search submission is debounced only where it does not impede explicit form
  behavior.
- Pagination shows Previous, Next, current page, total pages, total items, and
  page size.
- Out-of-range pages redirect or normalize to the last available page.

### Counted non-paginated endpoints

Action logs, telemetry, and node logs return counted data without page/limit.
The frontend must:

- encourage bounded time filters;
- avoid rendering an unbounded DOM;
- use incremental rendering/windowing when result sizes are large;
- preserve backend response semantics rather than inventing server pagination.

### Required visual states

Every collection provides:

- skeleton card loading;
- initial empty state;
- no-results state with Clear Filters;
- recoverable error with Retry;
- stale-data state;
- permission-limited/read-only state;
- mutation-pending state;
- purpose-built delete confirmation.

## Smart refresh

Enabled for:

- dashboard operational cards;
- nodes/connectivity;
- action history;
- telemetry;
- node logs.

Behavior:

- show Live/Paused/Stale plus “updated N seconds ago”;
- provide Pause/Resume and manual Refresh;
- pause when `document.visibilityState` is hidden;
- resume on visibility without discarding URL filters;
- avoid overlapping requests;
- temporarily pause during editing, dialogs, and confirmations;
- keep existing data on transient background failure;
- announce refresh failures non-disruptively;
- prevent refresh from resetting keyboard focus or scroll position.

The interval is centralized and resource-aware rather than duplicated by each
page. Automated requests stop on authentication failure.

## Data and security architecture

### Reads

- Server Components fetch initial data through `src/lib/api/*`.
- API wrappers remain one file per backend resource area.
- Typed request/response shapes are declared once beside their wrappers.
- Protected layouts fetch the current profile and effective permissions.
- Initial independent datasets are fetched in parallel.

### Mutations

- Server Actions live in shared API/resource modules or route-private `_lib`
  modules as appropriate.
- Each mutation rechecks session and relevant effective permission server-side.
- Forms expose field-level failures where the backend response permits.
- Successful mutations refresh only affected UI/data.

### Authentication

- Access and refresh JWTs remain `httpOnly` cookies.
- No token crosses the Server/Client Component boundary.
- Protected routing expands from the current `/dashboard` prefix to all
  authenticated routes.
- A `401` clears/ends invalid session state and returns to login safely.
- A `403` renders an access-denied state without leaking restricted data.

### Frontend API gaps

Implementation must add or complete wrappers for:

- node logs;
- firmware configuration parameters;
- node configuration read/write.

All existing wrappers must be checked against current backend routes before UI
work relies on them.

## Component architecture

### Shared layout

- `AppShell`
- `AppBar`
- `Sidebar`
- `SidebarGroup`
- `SidebarItem`
- `MobileNavigationDrawer`
- `Breadcrumbs`
- `ProfileDialog`
- `PermissionGate`

### Shared primitives

- `Button` and `IconButton`
- `Input`, `Textarea`, `Select`, `Checkbox`, `Label`
- `Dialog`, `Drawer`, `Tabs`
- `Card`, `ResourceCard`, `MetricCard`
- `StatusBadge`
- `SkeletonCard`
- `EmptyState`, `ErrorState`, `AccessDeniedState`
- `SearchFilterToolbar`
- `Pagination`
- `RefreshControl`
- `ConfirmationDialog`
- `JsonViewer` and `JsonEditor`
- existing toast system, extended only as needed

Route-specific components stay in private `_components`, `_hooks`, and `_lib`
folders until reused by another route.

## Theme and visual language

- Use the existing semantic Tailwind v4 tokens.
- Add semantic status, muted surface, border, destructive, and focus tokens in
  both light and dark modes.
- Never hardcode raw colors in component class names.
- Anton is limited to titles and short display labels.
- Inter handles navigation, forms, body copy, and data.
- Controls use approximately 12px radius; cards/dialogs use 16–24px.
- Borders are quiet and shadows minimal.
- Lucide icons use consistent rounded strokes.
- Avoid gradients, glassmorphism, excessive animation, and unrelated hues.

## Responsive behavior

- Mobile first.
- Card grid: one column mobile, two tablet, three or four wide desktop.
- Filters move into a drawer on constrained widths.
- Page actions wrap or move into a compact menu.
- Node tabs remain visible through horizontal scrolling.
- Dialogs become bottom sheets/full-width panels where appropriate.
- Touch targets remain approximately 44×44px or larger.
- No feature is desktop-only, even though desktop is the primary operations
  viewport.

## Accessibility

- Semantic landmarks, headings, navigation, forms, and buttons.
- Visible focus indicators.
- Keyboard-complete sidebar, menus, dialogs, drawers, tabs, cards, and
  pagination.
- Dialog focus trapping/restoration.
- `aria-current` for navigation.
- Live announcements for mutation results, refresh state, and validation.
- Status always pairs color with icon and text.
- Reduced-motion support.
- Contrast reviewed in both automatic theme modes.
- Card body links do not swallow nested action-button interactions.

## Error and confirmation model

- Initial page failure: full error-state card with Retry.
- Background refresh failure: preserve data and mark it stale.
- Field validation: inline near the field plus a concise form summary.
- Global mutation success/failure: toast plus persistent inline context when
  recovery requires it.
- Destructive resource deletion: explicit resource identity.
- Filter-scoped log/telemetry deletion: summarize active filters and affected
  count.
- High-risk or broad deletion may require typed confirmation.
- Backend error titles/messages are displayed safely; raw stack traces/details
  are not exposed.

## Verification strategy

### Unit tests

- permission-to-navigation mapping;
- group omission;
- URL query parsing/normalization;
- pagination calculations;
- status mappings;
- refresh pause/resume and visibility behavior;
- API query/response helpers.

### Component tests

- expanded/collapsed sidebar;
- mobile navigation focus behavior;
- profile view/edit/security flows;
- card nested interactions;
- search/filter toolbar;
- pagination;
- confirmation scope;
- loading, empty, error, stale, and read-only states.

### Route and integration tests

- authenticated-route protection;
- unauthorized page/action behavior;
- Server Action permission rechecks;
- validation error rendering;
- successful mutation refresh behavior.

### End-to-end journeys

1. Login, token refresh, and logout.
2. Edit profile and change password.
3. Browse/filter nodes and open a node.
4. Edit node metadata/configuration.
5. Select firmware and dispatch OTA.
6. Dispatch an action and inspect its result.
7. Filter telemetry and node logs.
8. Manage a user and inspect effective permissions.
9. Modify a role's permission assignment.
10. Create/edit a payload schema version.

### Visual verification

- Mobile, tablet, and desktop.
- Light and dark mode.
- Long names/messages, empty data, large result counts, and reduced motion.
- `npm run lint`, strict type checking, tests, and `npm run build`.

## Approved decisions

- Fleet health and urgent issues lead the dashboard.
- Operational data uses controlled smart auto-refresh.
- Roles, permissions, and assignments share one Access Control workspace.
- The Operations Cockpit information architecture is selected.
- Card collections are preferred across the application.
- The four reviewed design sections and linked Superdesign drafts are approved.
