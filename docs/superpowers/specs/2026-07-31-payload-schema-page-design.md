# Payload schema page redesign

## Context

The payload schemas admin page (`frontend/src/app/(authenticated)/admin/payload-schemas/`)
has two usability problems:

1. The list page's filter toolbar has a redundant "Open exact/latest" lookup
   form (exact name + optional version → server-side redirect straight to the
   detail page) alongside the normal search bar, plus a `valid_at`
   datetime-local filter that's fiddly to use just to answer "is this schema
   currently valid?".
2. The create/edit form's `definition` field (the JSON Schema-like shape that
   describes a payload's fields — see below) is a raw JSON textarea
   (`JsonEditor`). Authoring or editing a schema requires knowing the exact
   hand-rolled JSON shape the backend expects, which is not discoverable and
   error-prone by hand.

This spec covers both: simplifying the list page's filters, and replacing the
raw JSON editor with a guided, row/block-based builder.

No backend changes are needed for either — the builder only ever emits and
reads the exact same JSON shape the API already accepts and returns.

## Background: the `PayloadSchemaDefinition` shape

The backend (`backend/internal/domain/models/payload_schema.go`) defines a
single flat, recursive struct — not real JSON-Schema:

```go
type PayloadSchemaDefinition struct {
	Type          PayloadSchemaDefinitionType        `json:"type"`
	Required      []string                           `json:"required"`
	Options       []string                           `json:"options"`
	Unit          *string                            `json:"unit"`
	Minimum       *float64                           `json:"minimum"`
	Maximum       *float64                           `json:"maximum"`
	MinimumLength *int64                             `json:"minimum_length"`
	MaximumLength *int64                             `json:"maximum_length"`
	MinimumItem   *int64                             `json:"minimum_item"`
	MaximumItem   *int64                             `json:"maximum_item"`
	Properties    map[string]PayloadSchemaDefinition `json:"properties"`
	Items         *PayloadSchemaDefinition           `json:"items"`
}
```

12 type strings: `string`, `float`, `integer`, `boolean`, `enum`, `object`,
and their `[]`-prefixed array variants (`[]string` … `[]object`) — arrays are
a distinct type string, not a `type: array` + separate `items.type`. For
array types, if no distinct `items` sub-schema is given, the array node's own
scalar fields (`minimum`/`maximum`/`options`/`properties`/`required`/etc,
everything except `minimum_item`/`maximum_item`) are reused as the per-item
schema (`arrayItemSchema()` in
`backend/internal/infrastructure/utility/payload_schema_validator/validator.go`).
The guided builder only ever emits this simpler shape — it never generates a
redundant nested `items` object when the array's own fields already say the
same thing.

`unit` is free-text, never validated — purely a documentation hint (e.g. `°C`).
There is no `default`, `nullable`, or regex/`pattern` support anywhere in the
backend.

Root schemas are always `{"type": "object", "properties": {...}}` in every
schema created so far — payloads are JSON objects.

## Part 1: List page filter simplification

File: `frontend/src/app/(authenticated)/admin/payload-schemas/page.tsx`

- **Remove** the "Open exact/latest" lookup form (`lookup_name` +
  `lookup_version` inputs, its button, and the server-side redirect handling
  that calls `getLatestPayloadSchema`/`getPayloadSchemaByNameAndVersion`
  before the list even renders).
- **Remove** the `valid_at` datetime-local filter input.
- **Add** a "Still valid" checkbox next to the existing search bar. When
  checked, the list request passes `valid_at=<ISO timestamp captured at
  submit time>` to `listPayloadSchemas` — reusing the exact backend filter
  that already exists, just always "now" instead of an arbitrary picked date,
  and exposed as a boolean instead of a raw datetime.
- Pagination/query-string plumbing: drop `valid_at`, add `still_valid`
  (boolean) as a preserved query param alongside `search`/`limit`.
- **Unchanged**: the plain `search` input, and each `PayloadSchemaCard`'s own
  `valid_from`/`valid_to` display and Current/Future/Expired status badge —
  those are per-item info, not part of the filter bar, and stay exactly as
  they are today.

## Part 2: Guided schema definition builder

### Interaction model

Each nesting level (the root, or any Object/List-of-objects field) renders as
an ordered list of **field rows**, plus one trailing plain text input for
adding the next field, always present at the end of the list.

To add a field: type a snake_case name into that trailing input and press
**Enter or Space**. Typing a space auto-converts it to `_` as you type (so
`sample rate` becomes `sample_rate` mid-typing — it reads as "finishing a
word"). On commit, that input becomes a **field block**:

```
┌─────────────────────────────────────────────────────┐
│ sample_rate            [Whole number ▾]   ☐ Required │
│   Minimum ____   Maximum ____   Unit ____            │
└─────────────────────────────────────────────────────┘
```

- The name becomes a small pill/label (click to rename; renaming re-validates
  uniqueness among siblings).
- Type defaults to **Text**, and is a dropdown with 12 friendly entries (see
  table below) — one entry per backend type string, not a
  base-type-plus-"multiple values"-checkbox scheme.
- A **Required** checkbox sits next to the type. Checking it does not touch
  the row's own JSON — the parent object collects all checked rows into its
  own `required: [names]` array at serialization time (Data model section).
- Below the name/type/required line, that type's property inputs
  (table below) render immediately — no extra click.
- A small ✕ button removes the whole block (and everything nested under it).

### Per-type properties

| Friendly type | Backend type | Extra controls shown |
|---|---|---|
| Text | `string` | Min length, Max length |
| Number | `float` | Min, Max, Unit |
| Whole number | `integer` | Min, Max, Unit |
| Yes/No | `boolean` | *(none)* |
| Choice | `enum` | Options: tag-style add/remove list, ≥1 required |
| Object | `object` | *(none — see Nesting)* |
| List of text | `[]string` | Min items, Max items; item constraints: Min length, Max length |
| List of numbers | `[]float` | Min items, Max items; item constraints: Min, Max, Unit |
| List of whole numbers | `[]integer` | Min items, Max items; item constraints: Min, Max, Unit |
| List of yes/no | `[]boolean` | Min items, Max items |
| List of choices | `[]enum` | Min items, Max items; item constraints: Options (tag list) |
| List of objects | `[]object` | Min items, Max items; item constraints: nested property rows (Nesting) |

All numeric/length/item-count fields are optional — blank means "not sent",
matching the backend's `*float64`/`*int64` pointer fields; no invented
defaults. "Unit" is shown as a small suffix hint next to Min/Max, never
validated.

For array types, the "item constraints" controls are the *same widgets* the
scalar type would show — e.g. "List of numbers" shows the same Min/Max/Unit
inputs "Number" would — because they serialize onto the same array-node
fields the backend already falls back to as the per-item schema (see
Background). The builder never emits a distinct nested `items` object.

### Nesting

When a field's type is **Object** or **List of objects**, an indented
sub-list of field rows renders directly beneath it — the exact same row
anatomy and the same trailing "add field" input, indented one level with a
vertical connector line (file-tree style), always expanded (no
collapse/expand control). This nests arbitrarily deep. This is the simplest
option and keeps a whole schema glanceable in one view; payload schemas in
this system are shallow telemetry/action payload shapes, not deeply nested
API contracts, so an always-expanded view doesn't become unwieldy in
practice.

The root level *is* the top field list directly — no root wrapper row, no
root type picker. The form opens straight into the root object's fields.

### Internal data model

A recursive TypeScript type mirrors `PayloadSchemaDefinition` field-for-field
(`type`, `required`, `options`, `unit`, `minimum`, `maximum`,
`minimum_length`, `maximum_length`, `minimum_item`, `maximum_item`,
`properties`, `items`), plus one builder-only field: a stable local `key`
(random id) per row, for React list reconciliation.

Because `properties` is a JSON *map*, not an ordered list, but the UI needs a
stable insertion order for editing, each object level's fields are held in
builder state as an **ordered array of `{key, name, ...definition}`**, and
only collapsed into the backend's `Record<string, Definition>` +
`required: string[]` shape at submit time. Loading an existing schema for
edit does the reverse (map → ordered array) using the parsed JSON object's own
key insertion order (which `JSON.parse` preserves).

### Live "Preview JSON" panel

A collapsed-by-default panel below the builder shows the current definition
as pretty-printed, read-only JSON, regenerated live from builder state on
every change — for confidence/debugging. It is never editable; there is no
"advanced mode" toggle back into raw JSON entry.

### Validation & error handling

Client-side, live, advisory (inline red text under the specific control,
doesn't block typing but blocks Save):

- Field name: non-empty, matches `^[a-z0-9]+(_[a-z0-9]+)*$`, unique among
  siblings at that nesting level.
- Choice (enum): at least one option.
- Any Min/Max-style pair: error if both filled and Min > Max.
- Save is disabled while any row has an active error; a summary line ("3
  fields need attention") appears if an errored row is scrolled out of view.

Server-side (submit time): `saveSchema` in
`frontend/src/app/(authenticated)/admin/payload-schemas/_lib/actions.ts`
already returns `fieldErrors: Record<string, string>`. Backend structural
validation errors reference dotted paths like
`definition.readings.items.sample_rate` — the action's error handling is
extended to walk that path and highlight/scroll to the exact matching row,
falling back to a top-of-form banner if the path doesn't resolve to a
mounted row (defensive; client-side validation should already catch anything
that would produce this).

### Editing a schema the builder can't represent

On load, an existing schema's `definition` JSON is checked against what the
guided model can represent (12 known type strings, no unrecognized keys).
If it doesn't fit — e.g. created via direct API access — the entire edit
form (not just the definition field) is replaced with a read-only `<pre>`
block showing the raw JSON, plus a message that the guided editor can't
represent this definition and to use the API directly or contact an admin.
This never silently drops or reshapes data it doesn't fully understand.

## Component architecture

New files under
`frontend/src/app/(authenticated)/admin/payload-schemas/_components/`:

- **`schema-definition-types.ts`** — the recursive TS type, the
  map↔ordered-array conversion helpers (`definitionToRows` /
  `rowsToDefinition`), and the "can the builder represent this?" checker used
  by the edit-mode fallback.
- **`SchemaFieldRow.tsx`** — one field block: name pill, type dropdown,
  Required checkbox, remove button, its own type-specific property inputs;
  renders `SchemaFieldList` recursively for Object/List-of-objects children.
- **`SchemaFieldList.tsx`** — an ordered list of `SchemaFieldRow`s at one
  nesting level, plus the trailing "add field" input with the Enter/Space
  commit behavior. Owns add/remove/rename for its own level.
- **`SchemaDefinitionBuilder.tsx`** — top-level wrapper: the root
  `SchemaFieldList`, the collapsed Preview JSON panel, the unparseable-schema
  fallback view, and a hidden `<input type="hidden" name="definition">`
  carrying the serialized JSON so it still posts through the existing
  `<form action={...}>` / server-action pipeline unchanged.

Changed files:

- `PayloadSchemaForm.tsx`: swap `JsonEditor` for
  `<SchemaDefinitionBuilder defaultValue={schema?.definition} />` inside the
  existing `SchemaEditor` dialog; no other change to the dialog/action
  wiring.
- `page.tsx`: remove the lookup form, remove the `valid_at` input, add the
  `still_valid` checkbox (Part 1).
- `JsonEditor.tsx` is untouched — still used elsewhere (e.g. preferences) —
  this only stops using it on this page.

No backend changes.
