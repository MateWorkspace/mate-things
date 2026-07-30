# Mate Things design system

## Direction

An operations cockpit for an ESP32 fleet: warm, calm, approachable, and highly legible under operational pressure. “Cute” comes from rounded geometry, the mug/hat mark, and the brown/cream palette—not novelty decoration. Interfaces stay flat, spacious, and professional.

## Foundations

- Light background: `#ffffff`; dark background: `#1a1210`.
- Light foreground/ink: `#231f20`; dark foreground: `#ffeac5`.
- Primary brown: light `#603f26`, dark `#e0b988`.
- Surface cream: light `#ffeac5`, dark `#2a1f19`.
- Accent tan: `#c49a6c`.
- Highlight peach: `#fedbb5`.
- Body: Inter. Display headings: Anton.
- Rounded corners: 12px controls, 16–24px cards/dialogs.
- Borders: quiet 1px ink at 10–18% opacity. Shadows are subtle and rare.
- Mobile-first responsive layout; automatic OS light/dark modes.

## Semantic status

Operational status must combine icon, label, and color:

- Healthy/connected/success: warm-compatible green.
- Attention/warning/pending: amber.
- Critical/disconnected/failed: red.
- Informational/running: muted blue.
- Unknown/disabled: neutral ink/tan.

Status hues are supporting semantics only; the warm brand palette remains visually dominant.

## Authenticated shell

- Sticky top app bar with sidebar toggle at left.
- App mark/title follows the toggle when space allows.
- Right-aligned profile trigger combines the user name and anonymous human icon; either click target opens one profile dialog.
- Desktop sidebar is collapsible between labeled and icon-only states.
- Mobile/tablet sidebar is an overlay drawer with focus trapping and backdrop.
- Navigation is grouped as Overview, Fleet, Operations, Observability, and Administration.
- Entire groups and individual entries are omitted when the user lacks the required read permission.
- Content begins with breadcrumb, page title, supporting copy, freshness indicator, and contextual actions.

## Cards and collections

- Cards are the default collection primitive for nodes, firmware, classes, actions, users, roles, schemas, and history/log records.
- Cards expose the most decision-relevant fields, status, freshness, and one primary next action.
- Clicking the card body opens details; explicit action buttons remain separate targets.
- Responsive grid: one column on mobile, two on tablet, three or four on wide desktop depending on information density.
- Search, filters, sort, result count, refresh, and view-state controls sit in a compact toolbar above cards.
- Pagination is URL-backed and uses clear Previous/Next plus page position.
- Empty, error, loading, permission-denied, and no-filter-results states receive purpose-built cards.

## Dashboard

Fleet health leads. The first viewport contains:

1. Fleet-health summary and freshness/pause controls.
2. Urgent attention queue with disconnected nodes, failed actions, warnings, and OTA problems.
3. Connection health and action/OTA activity cards.
4. Recent node warnings and quick links into filtered views.

Avoid vanity metrics and decorative charts. Every summary should support a decision or drill-down.

## Node detail

The node is the core operational workspace:

- Header: connection status, identity, class, firmware, last update, and permitted actions.
- Tabs/sections: Overview, Configuration, Firmware & OTA, Actions, Telemetry, Node Logs.
- Context is preserved while switching sections.
- Mutations use focused dialogs/drawers with confirmation for destructive or high-impact operations.

## Profile dialog

- Opens from the name or anonymous human icon in the app bar.
- Summary header plus tabs: Profile and Security.
- Profile supports view/edit/cancel/save without leaving the current page.
- Security supports password change.
- Logout is available as a clearly separated secondary action.

## Motion and accessibility

- Motion is restrained to drawer/dialog transitions, card hover feedback, refresh indicators, and toasts.
- Honor reduced-motion preferences.
- Visible focus rings, semantic HTML, keyboard-complete dialogs/drawers, comfortable touch targets, and WCAG-oriented contrast are mandatory.
