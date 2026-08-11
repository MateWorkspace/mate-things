<!-- BEGIN:nextjs-agent-rules -->
# This is NOT the Next.js you know

This version has breaking changes — APIs, conventions, and file structure may all differ from your training data. Read the relevant guide in `node_modules/next/dist/docs/` before writing any code. Heed deprecation notices.
<!-- END:nextjs-agent-rules -->

# Project Context

This is the admin dashboard for `mate-things`, an IoT fleet-management
platform for ESP32-class devices (device registration, firmware OTA, action
dispatch, telemetry, RBAC — see the repo-root `AGENTS.md` for what the
product actually does). This app is a pure client of the Go backend's REST
API (`backend/`); it owns no database and no business logic of its own —
every rule of record (validation, permissions, MQTT orchestration) lives in
the backend. Treat this frontend as presentation + thin data-fetching only.

**Production topology**: the Go backend is the sole ingress. It serves
`/api/*` itself and reverse-proxies everything else to this Next.js app —
so in production, frontend and backend share the same origin, and
`fetch('/api/...')` works with no base URL. In local dev (`npm run dev`),
the backend is a separate process, so API calls need an absolute base URL
from an environment variable — never hardcode `http://localhost:8080` (or
any host) inline.

The backend authenticates with JWT access/refresh tokens. Never store
tokens in `localStorage`/`sessionStorage` or any place client JS can read —
that's an XSS-exfiltration risk against an admin panel. Set tokens as
`httpOnly` cookies from a Server Action or Route Handler, and read them
server-side (`cookies()`) when calling the backend from Server Components.

**One narrow, deliberate exception**: WebSocket handshakes can't carry an
`Authorization` header, so the backend's broadcast endpoints authenticate
via a `?token=` query param instead. `getBroadcastToken()`
(`src/app/(authenticated)/apps/infrared/record/[id]/_lib/broadcast-token.ts`)
is a Server Action that hands the short-lived access token to client JS
for exactly this purpose — the one sanctioned case where a token crosses
into the client bundle. Don't generalize this pattern; any other client
code that needs to call the backend should keep doing so through a Server
Action or Route Handler that reads the cookie server-side.

Stack: Next.js 16 (App Router, Turbopack), React 19, TypeScript (`strict`),
Tailwind CSS v4.

## Current product surface

Protected routes live under `src/app/(authenticated)/` and share the
permission-aware `AppShell`:

| Group | Routes | Primary permissions |
|---|---|---|
| Overview | `/dashboard` | Cards render only from reads the session may perform |
| Fleet | `/nodes`, `/nodes/[id]`, `/node-classes`, `/node-classes/[id]`, `/firmware`, `/firmware/[id]` | `node:*`, `node_config:*`, `node_class:*`, `firmware:*`, `ota:dispatch` |
| Operations | `/actions`, `/actions/[id]`, `/action-history` | `action:*`, `action_log:*` |
| Observability | `/telemetry`, `/node-logs`, `/broadcast-sessions` | `telemetry_record:*`, `node_log:*`, `broadcast_session:get` |
| Administration | `/admin/users`, `/admin/users/[id]`, `/admin/access-control`, `/admin/payload-schemas`, `/admin/payload-schemas/[id]`, `/admin/api-keys` | `user:*`, role/permission assignment permissions, `payload_schema:*`, `api_key:*` |

The node detail route is the contextual operations workspace: Overview,
Configuration, Firmware/OTA, Actions, Telemetry, and Logs are URL-selected
tabs and are individually permission-gated.

Navigation definitions live in `src/config/navigation.ts`, grouped by product
purpose and filtered through `visibleNavigation()`. Route authorization is
centralized in `src/config/route-policies.ts` (`ROUTE_POLICIES`,
`isProtectedRoute`, `canVisitRoute`, `findRoutePolicy`) — `isProtectedRoute`
is the actual authentication gate consumed by the root `proxy.ts` middleware,
not just a display concern, and a pathname with no matching entry fails
closed (treated as protected/not-visitable) rather than silently granting
access. These are presentation/routing conveniences, not substitutes for
page-level and Server Action permission checks.

# Code Rules

## Directory structure & colocation

- Shared assets, components, hooks, and lib code live directly in
  `src/assets/`, `src/components/`, `src/hooks/`, `src/lib/` — no `shared/`
  wrapper folder. Something belongs here the moment it's used by more than
  one route.
- API client functions (typed wrappers around backend endpoints) live in
  `src/lib/api/`, one file per backend resource area (e.g. `nodes.ts`,
  `actions.ts`, `auth.ts`), mirroring the backend's own resource grouping.
  Components never call `fetch` against the backend directly — they go
  through `src/lib/api/*`.
- Feature mutations live beside their route in `_lib/actions.ts`. Reusable
  session-level mutations may live in `src/lib/api/` or a shared component's
  action module. Keep raw API wrappers and Server Actions separate: wrappers
  express the backend transport contract; actions enforce session permissions,
  validate direct-call input, invoke wrappers, and revalidate or redirect only
  after a successful mutation.
- Anything used by exactly one route lives inside that route's folder in a
  private, non-routable subfolder prefixed with `_`:
  `the-page/_components`, `the-page/_hooks`, `the-page/_lib`,
  `the-page/_assets`. The leading underscore is a Next.js convention — it
  and everything under it are guaranteed excluded from routing, so this is
  always safe regardless of what you name the files inside.
- Standard Next.js routing files (`page.tsx`, `layout.tsx`, `loading.tsx`,
  `error.tsx`, `not-found.tsx`, `route.ts`, `template.tsx`) go directly in
  the route segment folder, never inside a private `_` folder. Use route
  groups `(group-name)` to organize routes or share a layout without
  affecting the URL.
- When something outgrows a single file, give it its own folder using the
  same `_components`/`_hooks`/`_lib` split one level down rather than
  piling unrelated concerns into one file.

## Naming

- Folders and `.ts` files: `kebab-case` (`node-status.ts`,
  `use-node-list.ts`).
- React `.tsx` component files: `PascalCase` (`NodeStatusBadge.tsx`),
  matching the component they default-export.
- Exception: primitive UI components in `src/components/ui/` (buttons,
  inputs, dialogs — the small reusable building blocks, not features) are
  named lowercase to match convention (`button.tsx`, `dialog.tsx`), even
  though they still default-export a `PascalCase` component.
- One default-exported component per `.tsx` file; the filename matches the
  export.

## Server and Client Components

Next.js App Router components are **Server Components by default**. Do not
add `'use client'` unless the component genuinely needs one of:

- React state or event handlers (`useState`, `onClick`, `onChange`, ...)
- Lifecycle/effects (`useEffect`) or browser-only APIs (`window`,
  `localStorage`, `navigator.geolocation`, ...)
- A custom hook that itself needs the above

Fetch data and read cookies/tokens in Server Components/Server Actions as
close to the source as possible, and pass only the resulting plain data
down as props to Client Components — never pass a token or fetch function
across the server/client boundary. Push `'use client'` as far down the
tree as possible (wrap just the interactive leaf, not the whole page) so
most of the tree stays server-rendered.

## Data fetching & mutations

- `fetch` is **not cached by default** in this Next.js version (this is a
  behavior change from older Next.js releases you may know — see
  `node_modules/next/dist/docs/01-app/02-guides/caching-without-cache-components.md`).
  Every request re-runs unless you explicitly opt in with
  `{ cache: 'force-cache' }` or `{ next: { revalidate: <seconds> } }`.
  This project's `next.config.ts` does not enable `cacheComponents`/
  `use cache`, so don't reach for the `'use cache'` directive here — use
  explicit `fetch` cache options instead, and only cache reads that are
  safe to serve stale for the given window (e.g. not live device status).
- Reads: fetch in Server Components (or a `src/lib/api/*` function they
  call), so data loads before any JS ships to the client.
- Mutations (dispatching an action, uploading firmware, editing a node):
  use Server Actions (`'use server'`), normally in the owning route's
  `_lib/actions.ts`, called from a `<form action={...}>` or from a Client
  Component's event handler. Always
  re-verify the caller's session/permissions inside the Server Action
  itself — it's a public POST endpoint regardless of which UI calls it,
  same as the backend does for its own handlers.
- Never call `revalidatePath`/`revalidateTag` speculatively — only after a
  mutation actually changes what a cached read would return.
- Parallelize independent reads with `Promise.all`, but only issue a read when
  the session owns its permission. A hidden panel that still performs a
  forbidden background request is broken.
- Use local `loading.tsx`, `error.tsx`, and `not-found.tsx` boundaries for
  meaningful transitions and recoverable failures. Detail pages map backend
  404 responses to `notFound()`.
- Dashboard, telemetry, action-history, and node-log data is operationally
  live. Do not cache it as static data. Shared smart refresh pauses while the
  tab is hidden, supports manual refresh, and keeps query windows bounded.

## API and device-operation contracts

- `src/lib/api/client.ts` owns authorization headers, JSON encoding, error
  normalization, and same-origin/backend-base resolution. Do not duplicate
  those concerns in resource wrappers.
- Collection wrappers accept typed queries and return the backend's paginated
  envelope. Complete selectors (`listAll*` functions) call the shared
  `collectAllPages` helper (`src/lib/api/collect-all-pages.ts`) for
  deduplication and safe termination (it throws past a 1000-page safety
  ceiling rather than looping forever) instead of hand-rolling a page-walk
  loop; never impose a silent per-caller item cap.
- Resolving an optional filter's display name from an id (e.g. a `role_id`
  query param → its role name for a combobox default) goes through
  `getOptionalById` (`src/lib/api/optional.ts`), which returns `null` on a
  backend 404 but rethrows everything else — never hand-roll
  `.catch(() => null)`, which would also swallow a transient 5xx as "not
  found".
- Node `device_id` is immutable. Show it as identity/context, never as an
  ordinary editable field.
- OTA dispatch sends `{ firmware_id }`. The backend authoritatively checks node
  compatibility, resolves the binary URL, and publishes checksum/size data.
  The frontend mutation requires `ota:dispatch` and must not introduce a
  `firmware:get` dependency.
- Firmware delete sends the typed confirmation as the expected name; never
  send a separate client-asserted authoritative name.
- Firmware binary replacement preserves the distinction between omitted
  `config_schema` (unchanged) and explicit `[]` (clear).
- String node-config values are passed verbatim, including empty and
  whitespace-only values. Do not `trim()` them into a different value or treat
  empty as absent.

## Collections, cards, and records

- Resource collections are card-first. Use
  `src/components/collection/` and URL-owned search/filter/page state; do not
  introduce a desktop table without a materially better small-screen
  representation. Telemetry, Node Logs, and API Keys are the established
  `<table>` exceptions (record-dense, tabular-by-nature data) — API Keys
  pairs that table with real `page`/`limit` pagination (`<Pagination>`,
  like Users/Nodes) rather than Telemetry/Node Logs' unpaginated
  batch-load, since it's a bounded per-user resource, not a time-series one.
- Every collection page's filter form uses the shared
  `src/components/collection/FilterBar.tsx` shell (a `<form>` with a
  `grid gap-3 sm:grid-cols-2 lg:grid-cols-3` field area and a `border-t`
  footer with a "Clear" link + "Apply" button) — filters are always
  visible; there is no collapsible mobile filter drawer. Entity-reference
  filters (node, node class, firmware, action, role, user) use the
  matching `*SearchCombobox` component, backed by a `search<Entity>Action`
  Server Action in `src/lib/actions/entity-search-actions.ts` (typed via
  `src/lib/actions/search-options.ts`) — never a native `<select>`
  populated from a full `listAll*()` fetch. Enum-valued filters (status,
  level) use `src/components/ui/select.tsx`, not a raw `<select>`, so the
  dropdown chevron doesn't collide with the browser's own arrow.
- Secret-bearing values (API keys) are shown in full exactly once,
  immediately after generate/regenerate, next to a copy-to-clipboard
  control (`src/components/ui/copy-button.tsx`); every list/table view
  after that only ever shows a masked suffix (`key_last_four`). Follow this
  same show-once pattern for any future credential-like resource — never
  make a raw secret re-viewable from a list.
- Parse pagination through `src/lib/collection-query.ts`. Page numbers must be
  positive safe integers; preserve filters when redirecting an out-of-range
  page.
- Distinguish a truly empty collection from “no results match these filters.”
  Their copy and recovery actions are different.
- Action history, telemetry, and node logs share
  `src/components/records/` controls for bounded time ranges, JSON inspection,
  record windows, and scoped deletion.
- Destructive and result-bearing dialogs (delete confirmations, API key
  generation, action dispatch) share `useActionDialog`
  (`src/hooks/use-action-dialog.ts`) for open/reset/remount-on-reopen
  lifecycle instead of hand-rolling `[open, setOpen]` +
  `[generation, setGeneration]` + `key={generation}`. It defaults to
  closing (and resetting the form) as soon as the action state reaches
  `"success"` — pass `closeOnSuccess: false` for dialogs that must stay
  open to show a result in place (a one-time secret, a "Done"
  confirmation) instead of auto-closing over it. Destructive actions
  require explicit confirmation, remain open while pending, surface
  backend failures, and cannot repeat from stale success state. Keep
  dialogs mounted and control `open` so focus restoration works.
- Permission/action-checklist assignment UI (role permissions, node-class
  actions) uses `useAssignmentSelection`
  (`src/hooks/use-assignment-selection.ts`) to reconcile optimistic
  checkbox overrides against the authoritative server-confirmed set across
  submissions. It keys its reconciliation off the authoritative `Set`'s
  reference identity, not its contents — memoize the `Set` you pass it
  (`useMemo`) rather than constructing a new one inline every render.

## TypeScript

- `strict` mode is on (see `tsconfig.json`) — keep it that way. No `any`;
  use `unknown` and narrow, or define a proper type/interface.
- Give exported functions and component props explicit types; let
  inference handle the rest.
- Import backend response/request shapes from a single typed source
  (colocate with the relevant `src/lib/api/*.ts` file) instead of
  redeclaring the same shape in multiple components.

## Images, fonts, environment

- Always use `next/image`, never a raw `<img>` — it's required for layout
  stability and automatic format/size optimization, and every image needs
  real `alt` text (decorative-only images get `alt=""`, not a missing
  attribute).
- Load every font — brand and body — through `next/font/local` or
  `next/font/google`, never a `<link>` tag or `@import` from a font CDN.
  Self-hosting is required for both privacy (no third-party font requests)
  and layout stability. See [Design Rules](#design-rules) for which fonts.
- `.env.local` (untracked) for local secrets/overrides. Only prefix a
  variable `NEXT_PUBLIC_` if its value is genuinely safe to ship to the
  browser — never a backend API secret, service token, or anything
  auth-related.

## Linting & verification

- Available verification commands are:

  ```bash
  npm run typecheck
  npm run lint
  npm test
  npm run build
  npm run test:e2e
  ```

- At minimum, `typecheck`, `lint`, and the production build must pass before
  considering a frontend change done. Run focused Vitest coverage while
  iterating and the full suite for shared/auth/API changes. Use Playwright for
  login, shell, responsive drawer, focus, and browser accessibility behavior.
- Turbopack is the default bundler for both `next dev` and `next build` in
  this version — don't add webpack-specific config unless there's a
  concrete reason, and note it if you do.
- Server Actions accept firmware uploads up to the explicit bounded limit in
  `next.config.ts`; do not raise it casually or bypass it with browser-visible
  backend credentials.

## Accessibility

Use semantic HTML elements (`<button>`, `<nav>`, headings in
order) over generic `<div>`s with click handlers. Icon-only controls need
an `aria-label`. Interactive elements must be keyboard-reachable and show a
visible focus state — don't strip default focus rings without replacing
them.

The shell includes a skip link, modal focus containment/restoration, and a
mobile navigation drawer. Preserve those behaviors when changing navigation,
the profile dialog, or shared `Dialog`; Escape/overlay close must be ignored
while a mutation is pending.

The sidebar and the main content pane (`#main-content`) are independent
scroll regions inside a fixed `h-screen` shell
(`src/components/layout/AppShell.tsx`) — the document itself never scrolls.
A `pathname`-keyed effect resets `#main-content`'s scroll to top on
navigation, since Next.js's default scroll-to-top targets `window`, which no
longer scrolls here. Preserve both properties (independent scroll regions,
scroll-reset-on-navigate) when touching the shell.

# Design Rules

## Color palette

Base palette: `#000000` (black), `#FFFFFF` (white), `#603f26`, `#c49a6c`,
`#fedbb5`, `#ffeac5` — plus `#231f20` (a near-black "ink" tone, not pure
black), which appears as the outline/stroke color throughout the brand
assets (`src/assets/hat.svg`, `horizontal.svg`, `vertical.svg`) alongside
the others. Treat it as part of the palette, distinct from `#000000`, for
outlines and fine detail where pure black would feel too harsh against the
warm palette.

Ordered by lightness, the four brand browns/tans form a natural tint ramp
from ink to background — use this ordering when deriving new shades
in-between, rather than picking arbitrary hex values:

| Token (suggested) | Hex | Role, as used in the brand assets |
|---|---|---|
| `ink` | `#231f20` | Near-black outline/stroke detail |
| `primary` | `#603f26` | Dark brown — logo fill, wordmark, primary ink on light surfaces |
| `primary-muted` / `accent` | `#c49a6c` | Mid tan — secondary accents, borders, muted text on light surfaces |
| `highlight` | `#fedbb5` | Soft peach — highlights, hover/active tints |
| `surface` | `#ffeac5` | Pale cream — the card/badge background behind the mark in `vertical.svg` |

Define these as Tailwind v4 `@theme` tokens in `src/app/globals.css`
(alongside the existing `--color-background`/`--color-foreground`
pattern already there) with semantic names, e.g. `--color-primary`,
`--color-surface`, so components reference `bg-primary`/`text-primary`
rather than raw hex — this keeps a future dark-mode or rebrand pass to one
file. You may expand the palette with tints/shades derived from these five
colors (for hover, disabled, focus-ring states, etc.) but don't introduce
unrelated hues — everything should read as part of the same warm,
brownish-pastel family. Every semantic token needs both a light and a dark
value — see [Dark & light mode](#dark--light-mode) below.

## Dark & light mode

`globals.css` already redefines `--background`/`--foreground` inside an
`@media (prefers-color-scheme: dark)` block; follow that exact pattern for
every semantic color token added above (`--color-primary`,
`--color-surface`, `--color-accent`, `--color-highlight`, `--color-ink`,
and any derived tint/shade) — never a hex value hardcoded directly in a
component's `className`. Concretely:

- Pick the light-mode value from the brand palette as given (it's
  designed for a light, cream surface). For dark mode, don't just invert —
  darken the `surface`/`background` toward near-black, keep `primary`
  legible against it (lighten it if `#603f26` reads too muddy on a dark
  background), and keep the warm brown/tan hue family in both modes so it
  still reads as the same brand.
- **No component is done until it's been checked in both modes.** When
  building or reviewing any UI, mentally (or actually, via OS/browser
  dark-mode toggle) render it against both `@media (prefers-color-scheme:
  light)` and `dark` before calling it finished — check text contrast,
  border visibility, and that icons/illustrations (many of which are flat
  SVG fills) don't disappear against a flipped background.
- This project has no manual theme toggle (`prefers-color-scheme` only) —
  don't build one unless asked; just make sure both automatic states look
  correct.

## Mood & shape language

Brownish-pastel, simplistic, cute-looking, but professional — derived from
`src/assets/{hat,horizontal,vertical}.svg`. Concretely:

- **Rounded, friendly shapes.** The brand mark uses thick, rounded strokes
  and a generously-rounded container (`vertical.svg`'s badge uses a 32px
  corner radius on a 265px square — proportionally very rounded). Prefer
  large-radius rounded corners on cards, buttons, and inputs
  (`rounded-xl`/`rounded-2xl` territory) over sharp corners.
  Illustrations/icons should favor thick, consistent stroke weights with
  rounded terminals, matching the mug/hat glyph.
- **Simplicity over ornamentation.** Flat fills, minimal-to-no drop
  shadows (soft/subtle if used at all), generous whitespace. "Cute" comes
  from the shapes and palette, not from decoration.
  Keep it professional: no skeuomorphism, no gratuitous animation, no
  playful iconography that undercuts an admin/ops tool's credibility.

## Typography

The wordmark uses a Coolvetica-style face — a bold, condensed, geometric
display sans. **No font file for it exists in this repo yet**
(`src/assets/` only has the SVGs, which have the letterforms pre-outlined
as paths, not a usable font). Before using it in real UI text, either
source a licensed Coolvetica font file and self-host it via
`next/font/local` (see
`node_modules/next/dist/docs/01-app/01-getting-started/13-fonts.md`), or
substitute a close, properly-licensed free alternative with the same
condensed/geometric/bold character (e.g. Archivo Black, Anton, or a
heavy-weight Poppins) — do not link to a Coolvetica CDN copy of unclear
license.

Use the rules below regardless of which face is ultimately chosen:

- **Display face** (Coolvetica or its substitute): logo/wordmark
  treatments, page titles, and primary CTAs only. It's a display face —
  condensed display fonts hurt readability at body-copy sizes and in
  longer runs of text, so don't use it for paragraphs, table cells, or
  form labels.
  - Load via `next/font/local` (or `next/font/google` if using a
    Google-hosted substitute) and expose it as a Tailwind font token
    (e.g. `--font-display`) in `globals.css`, not an inline `style`.
- **Body face**: a clean, humanist/geometric sans with good legibility at
  small sizes (e.g. Inter, Geist, or a similar Google font) for
  everything else — body text, labels, table data, nav. Load via
  `next/font/google`, expose as `--font-sans`, and keep it as the default
  `font-family` on `<body>`.

## Responsiveness

This is an admin/ops dashboard — desktop is the primary use case (device
collections, record inspection, dispatch forms), but it must stay usable on tablet and
mobile (an operator checking a device from their phone), not just avoid
visibly breaking:

- Build **mobile-first**: unprefixed Tailwind classes are the small-screen
  baseline; layer `sm:`/`md:`/`lg:`/`xl:` on top for wider viewports,
  rather than designing for desktop and bolting on a mobile fallback.
- Data-dense elements need an explicit small-screen plan. Prefer the
  established card/list and pagination patterns, hide secondary metadata when
  justified, and reserve horizontal scrolling for inherently tabular content.
- Never hardcode pixel widths/heights that break at common breakpoints;
  use relative units (`%`, `rem`, `fr`, `flex`/`grid`) and Tailwind's
  breakpoint scale so layouts reflow instead of overflowing or clipping.
- Touch targets (buttons, icon-only controls, table row actions) need to
  stay comfortably tappable at mobile widths — don't shrink interactive
  elements below a reasonable touch-target size just to fit more on
  screen.
- **No component is done until it's been checked across at least a
  mobile, tablet, and desktop viewport** (e.g. browser dev-tools device
  toolbar) — the same bar as the dark/light-mode check above.
