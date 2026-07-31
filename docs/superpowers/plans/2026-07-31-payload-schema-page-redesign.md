# Payload Schema Page Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the payload schemas list page's redundant lookup/date filters with a single "Still valid" checkbox, and replace the raw-JSON `definition` textarea in the create/edit form with a guided, row-based field builder that mirrors the backend's `PayloadSchemaDefinition` shape.

**Architecture:** A small recursive TS module (`schema-definition-types.ts`) converts between the backend's map-shaped JSON (`{type, properties, required, ...}`) and an ordered-array "row" model the UI can render and edit. Two mutually-recursive React components (`SchemaFieldList` renders a list of rows at one nesting level + an "add field" input; `SchemaFieldRow` renders one row's controls and, for Object/List-of-objects, nests another `SchemaFieldList`) drive the builder. A wrapper (`SchemaDefinitionBuilder`) owns the root row list, a live JSON preview, the unparseable-schema fallback, and a hidden form input carrying the serialized JSON so the existing `<form action={...}>` server-action pipeline needs no changes. No backend changes.

**Tech Stack:** Next.js App Router (React 19, Server Actions), TypeScript, Tailwind, Vitest + @testing-library/react + @testing-library/user-event for tests (matching this repo's existing test files, e.g. `src/components/ui/dialog.test.tsx`).

## Global Constraints

- Field names must match `^[a-z0-9]+(_[a-z0-9]+)*$` (same pattern the backend's `RequiredSnakeCaseName` already enforces server-side for schema names) — reuse this exact regex client-side.
- The 12 backend type strings are exactly: `string`, `float`, `integer`, `boolean`, `enum`, `object`, `[]string`, `[]float`, `[]integer`, `[]boolean`, `[]enum`, `[]object`. Never invent new ones; never emit a distinct `items` sub-schema (array item constraints reuse the array node's own fields, matching `arrayItemSchema()` in `backend/internal/infrastructure/utility/payload_schema_validator/validator.go`).
- The builder must never silently drop or reshape a definition it doesn't fully understand — unrepresentable existing schemas get a read-only fallback view, never a lossy round-trip.
- No backend changes; no changes to `frontend/src/lib/api/payload-schemas.ts` or `frontend/src/components/json/JsonEditor.tsx` (still used elsewhere).
- Design spec: `docs/superpowers/specs/2026-07-31-payload-schema-page-design.md` — every task below implements a part of it; consult it for the "why" behind any control.

---

### Task 1: Recursive definition types and row⇄JSON conversion helpers

**Files:**
- Create: `frontend/src/app/(authenticated)/admin/payload-schemas/_components/schema-definition-types.ts`
- Test: `frontend/src/app/(authenticated)/admin/payload-schemas/_components/schema-definition-types.test.ts`

**Interfaces:**
- Produces: `FieldType` (union of the 12 type strings), `FIELD_TYPES: FieldType[]`, `FIELD_TYPE_LABELS: Record<FieldType, string>`, `isObjectType`, `isArrayType`, `isEnumType`, `RawDefinition` (backend JSON shape), `FieldRow` (builder row shape), `createEmptyRow(name: string): FieldRow`, `isValidFieldName(name: string): boolean`, `rowsFromRootDefinition(def: RawDefinition): FieldRow[]`, `rootDefinitionFromRows(rows: FieldRow[]): RawDefinition`, `canRepresentDefinition(value: unknown): value is RawDefinition`.

- [ ] **Step 1: Write the failing tests**

```ts
// frontend/src/app/(authenticated)/admin/payload-schemas/_components/schema-definition-types.test.ts
import { describe, expect, it } from "vitest";

import {
  canRepresentDefinition,
  isValidFieldName,
  rootDefinitionFromRows,
  rowsFromRootDefinition,
  type RawDefinition,
} from "./schema-definition-types";

describe("isValidFieldName", () => {
  it("accepts lowercase snake_case names", () => {
    expect(isValidFieldName("sample_rate")).toBe(true);
    expect(isValidFieldName("a")).toBe(true);
  });

  it("rejects empty, uppercase, leading-digit-only-ok, and malformed names", () => {
    expect(isValidFieldName("")).toBe(false);
    expect(isValidFieldName("SampleRate")).toBe(false);
    expect(isValidFieldName("sample__rate")).toBe(false);
    expect(isValidFieldName("_sample")).toBe(false);
    expect(isValidFieldName("sample_")).toBe(false);
  });
});

describe("canRepresentDefinition", () => {
  it("accepts a plain object root with no properties", () => {
    expect(canRepresentDefinition({ type: "object", properties: {} })).toBe(
      true,
    );
  });

  it("rejects an unknown type string", () => {
    expect(canRepresentDefinition({ type: "date", properties: {} })).toBe(
      false,
    );
  });

  it("rejects any definition (at any depth) that carries a distinct items schema", () => {
    const def = {
      type: "object",
      properties: {
        readings: { type: "[]float", items: { type: "float" } },
      },
    };
    expect(canRepresentDefinition(def)).toBe(false);
  });

  it("rejects a non-object value", () => {
    expect(canRepresentDefinition(null)).toBe(false);
    expect(canRepresentDefinition([])).toBe(false);
    expect(canRepresentDefinition("object")).toBe(false);
  });

  it("recurses into nested object properties", () => {
    const def = {
      type: "object",
      properties: {
        location: {
          type: "object",
          properties: { lat: { type: "not_a_real_type" } },
        },
      },
    };
    expect(canRepresentDefinition(def)).toBe(false);
  });
});

describe("rowsFromRootDefinition / rootDefinitionFromRows round trip", () => {
  it("round-trips a definition with scalar, enum, nested object, and array-of-object fields", () => {
    const original: RawDefinition = {
      type: "object",
      required: ["sample_rate"],
      properties: {
        sample_rate: {
          type: "float",
          minimum: 0,
          maximum: 1000,
          unit: "Hz",
        },
        mode: {
          type: "enum",
          options: ["auto", "manual"],
        },
        location: {
          type: "object",
          properties: {
            lat: { type: "float" },
            lon: { type: "float" },
          },
        },
        readings: {
          type: "[]object",
          minimum_item: 1,
          properties: {
            value: { type: "float" },
          },
        },
      },
    };

    const rows = rowsFromRootDefinition(original);
    const rebuilt = rootDefinitionFromRows(rows);

    expect(rebuilt).toEqual(original);
  });

  it("omits unset optional numeric fields entirely rather than writing null", () => {
    const original: RawDefinition = {
      type: "object",
      properties: { name: { type: "string" } },
    };

    const rebuilt = rootDefinitionFromRows(rowsFromRootDefinition(original));

    expect(rebuilt.properties?.name).toEqual({ type: "string" });
    expect(Object.keys(rebuilt.properties?.name ?? {})).toEqual(["type"]);
  });

  it("marks a row required based on the parent's required list, not a stored flag", () => {
    const original: RawDefinition = {
      type: "object",
      required: ["name"],
      properties: {
        name: { type: "string" },
        bio: { type: "string" },
      },
    };

    const rows = rowsFromRootDefinition(original);
    const name = rows.find((row) => row.name === "name");
    const bio = rows.find((row) => row.name === "bio");

    expect(name?.required).toBe(true);
    expect(bio?.required).toBe(false);
  });
});
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd frontend && npx vitest run src/app/\(authenticated\)/admin/payload-schemas/_components/schema-definition-types.test.ts`
Expected: FAIL with "Cannot find module './schema-definition-types'" (file doesn't exist yet).

- [ ] **Step 3: Write the implementation**

```ts
// frontend/src/app/(authenticated)/admin/payload-schemas/_components/schema-definition-types.ts

export type FieldType =
  | "string"
  | "float"
  | "integer"
  | "boolean"
  | "enum"
  | "object"
  | "[]string"
  | "[]float"
  | "[]integer"
  | "[]boolean"
  | "[]enum"
  | "[]object";

export const FIELD_TYPE_LABELS: Record<FieldType, string> = {
  string: "Text",
  float: "Number",
  integer: "Whole number",
  boolean: "Yes/No",
  enum: "Choice",
  object: "Object",
  "[]string": "List of text",
  "[]float": "List of numbers",
  "[]integer": "List of whole numbers",
  "[]boolean": "List of yes/no",
  "[]enum": "List of choices",
  "[]object": "List of objects",
};

export const FIELD_TYPES = Object.keys(FIELD_TYPE_LABELS) as FieldType[];

const FIELD_TYPE_SET = new Set<string>(FIELD_TYPES);

export function isArrayType(type: FieldType): boolean {
  return type.startsWith("[]");
}

export function baseType(type: FieldType): string {
  return isArrayType(type) ? type.slice(2) : type;
}

export function isObjectType(type: FieldType): boolean {
  return baseType(type) === "object";
}

export function isEnumType(type: FieldType): boolean {
  return baseType(type) === "enum";
}

export const FIELD_NAME_PATTERN = /^[a-z0-9]+(_[a-z0-9]+)*$/;

export function isValidFieldName(name: string): boolean {
  return FIELD_NAME_PATTERN.test(name);
}

/**
 * The backend's PayloadSchemaDefinition JSON shape (payload_schema.go). The
 * Go struct has no `omitempty` tags, so a definition that was ever
 * re-marshaled by Go (as opposed to round-tripped as raw bytes, which is
 * what the create/update path actually does today) could carry explicit
 * `null`s on unset optional fields rather than omitting the key — every
 * optional field here allows `| null` so that shape is never misread.
 */
export interface RawDefinition {
  type: string;
  required?: string[] | null;
  options?: string[] | null;
  unit?: string | null;
  minimum?: number | null;
  maximum?: number | null;
  minimum_length?: number | null;
  maximum_length?: number | null;
  minimum_item?: number | null;
  maximum_item?: number | null;
  properties?: Record<string, RawDefinition> | null;
  items?: RawDefinition | null;
}

/** One field's editable state at a single nesting level. */
export interface FieldRow {
  key: string;
  name: string;
  type: FieldType;
  required: boolean;
  unit: string;
  minimum: string;
  maximum: string;
  minimumLength: string;
  maximumLength: string;
  minimumItem: string;
  maximumItem: string;
  options: string[];
  children: FieldRow[];
}

function randomKey(): string {
  return typeof crypto !== "undefined" && crypto.randomUUID
    ? crypto.randomUUID()
    : `${Date.now()}-${Math.random()}`;
}

export function createEmptyRow(name: string): FieldRow {
  return {
    key: randomKey(),
    name,
    type: "string",
    required: false,
    unit: "",
    minimum: "",
    maximum: "",
    minimumLength: "",
    maximumLength: "",
    minimumItem: "",
    maximumItem: "",
    options: [],
    children: [],
  };
}

function numberToText(value: number | null | undefined): string {
  return value === undefined || value === null ? "" : String(value);
}

function textToNumber(value: string): number | undefined {
  const trimmed = value.trim();
  if (!trimmed) return undefined;
  const parsed = Number(trimmed);
  return Number.isFinite(parsed) ? parsed : undefined;
}

function rowFromDefinition(
  name: string,
  def: RawDefinition,
  requiredNames: ReadonlySet<string>,
): FieldRow {
  const type = def.type as FieldType;
  const properties = def.properties ?? {};
  const childRequired = new Set(def.required ?? []);
  const children = isObjectType(type)
    ? Object.entries(properties).map(([childName, childDef]) =>
        rowFromDefinition(childName, childDef, childRequired),
      )
    : [];

  return {
    key: randomKey(),
    name,
    type,
    required: requiredNames.has(name),
    unit: def.unit ?? "",
    minimum: numberToText(def.minimum),
    maximum: numberToText(def.maximum),
    minimumLength: numberToText(def.minimum_length),
    maximumLength: numberToText(def.maximum_length),
    minimumItem: numberToText(def.minimum_item),
    maximumItem: numberToText(def.maximum_item),
    options: def.options ?? [],
    children,
  };
}

export function rowsFromRootDefinition(def: RawDefinition): FieldRow[] {
  const properties = def.properties ?? {};
  const requiredNames = new Set(def.required ?? []);
  return Object.entries(properties).map(([name, childDef]) =>
    rowFromDefinition(name, childDef, requiredNames),
  );
}

function rowToDefinition(row: FieldRow): RawDefinition {
  const def: RawDefinition = { type: row.type };

  if (row.unit.trim()) def.unit = row.unit.trim();

  const minimum = textToNumber(row.minimum);
  if (minimum !== undefined) def.minimum = minimum;
  const maximum = textToNumber(row.maximum);
  if (maximum !== undefined) def.maximum = maximum;
  const minimumLength = textToNumber(row.minimumLength);
  if (minimumLength !== undefined) def.minimum_length = minimumLength;
  const maximumLength = textToNumber(row.maximumLength);
  if (maximumLength !== undefined) def.maximum_length = maximumLength;
  const minimumItem = textToNumber(row.minimumItem);
  if (minimumItem !== undefined) def.minimum_item = minimumItem;
  const maximumItem = textToNumber(row.maximumItem);
  if (maximumItem !== undefined) def.maximum_item = maximumItem;

  if (isEnumType(row.type) && row.options.length) def.options = row.options;

  if (isObjectType(row.type)) {
    const { properties, required } = rowsToProperties(row.children);
    def.properties = properties;
    if (required.length) def.required = required;
  }

  return def;
}

export function rowsToProperties(rows: readonly FieldRow[]): {
  properties: Record<string, RawDefinition>;
  required: string[];
} {
  const properties: Record<string, RawDefinition> = {};
  const required: string[] = [];
  for (const row of rows) {
    if (!row.name) continue;
    properties[row.name] = rowToDefinition(row);
    if (row.required) required.push(row.name);
  }
  return { properties, required };
}

export function rootDefinitionFromRows(rows: readonly FieldRow[]): RawDefinition {
  const { properties, required } = rowsToProperties(rows);
  const def: RawDefinition = { type: "object", properties };
  if (required.length) def.required = required;
  return def;
}

/**
 * Whether the guided builder can fully represent `value` without dropping
 * or reshaping anything. Rejects unknown type strings and any use of a
 * distinct `items` sub-schema (the builder only ever emits array item
 * constraints as the array node's own fields, per arrayItemSchema() in the
 * Go validator).
 */
export function canRepresentDefinition(value: unknown): value is RawDefinition {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    return false;
  }
  const def = value as Record<string, unknown>;

  if (typeof def.type !== "string" || !FIELD_TYPE_SET.has(def.type)) {
    return false;
  }
  if (def.items !== undefined && def.items !== null) return false;

  const type = def.type as FieldType;

  if (isEnumType(type) && def.options !== undefined && def.options !== null) {
    if (
      !Array.isArray(def.options) ||
      def.options.some((option) => typeof option !== "string")
    ) {
      return false;
    }
  }

  if (isObjectType(type)) {
    if (def.properties !== undefined && def.properties !== null) {
      if (typeof def.properties !== "object" || Array.isArray(def.properties)) {
        return false;
      }
      for (const child of Object.values(
        def.properties as Record<string, unknown>,
      )) {
        if (!canRepresentDefinition(child)) return false;
      }
    }
    if (def.required !== undefined && def.required !== null) {
      if (
        !Array.isArray(def.required) ||
        def.required.some((requiredName) => typeof requiredName !== "string")
      ) {
        return false;
      }
    }
  }

  return true;
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd frontend && npx vitest run src/app/\(authenticated\)/admin/payload-schemas/_components/schema-definition-types.test.ts`
Expected: PASS (9 tests).

- [ ] **Step 5: Typecheck**

Run: `cd frontend && npx tsc --noEmit`
Expected: no new errors from this file.

- [ ] **Step 6: Commit**

```bash
git add frontend/src/app/\(authenticated\)/admin/payload-schemas/_components/schema-definition-types.ts frontend/src/app/\(authenticated\)/admin/payload-schemas/_components/schema-definition-types.test.ts
git commit -m "feat(payload-schemas): add recursive definition⇄row conversion helpers"
```

---

### Task 2: Backend field-error path parser

**Files:**
- Create: `frontend/src/app/(authenticated)/admin/payload-schemas/_components/schema-definition-errors.ts`
- Test: `frontend/src/app/(authenticated)/admin/payload-schemas/_components/schema-definition-errors.test.ts`

**Interfaces:**
- Consumes: nothing from Task 1 (pure string parsing).
- Produces: `parseDefinitionFieldError(message: string): { path: string[]; message: string } | null`.

The backend's structural validator (`validatePayloadSchemaDefinitionShape` in
`backend/internal/application/shared/validation.go`) returns errors whose
message is `<dotted-field-path> <english suffix>`, e.g.
`"definition.readings.sample_rate has an invalid or missing type"`. This
parser recovers the path so the UI can highlight the exact row instead of
only showing a top-of-form banner.

- [ ] **Step 1: Write the failing tests**

```ts
// frontend/src/app/(authenticated)/admin/payload-schemas/_components/schema-definition-errors.test.ts
import { describe, expect, it } from "vitest";

import { parseDefinitionFieldError } from "./schema-definition-errors";

describe("parseDefinitionFieldError", () => {
  it("extracts a nested path from an invalid-type error", () => {
    expect(
      parseDefinitionFieldError(
        "definition.readings.sample_rate has an invalid or missing type",
      ),
    ).toEqual({
      path: ["readings", "sample_rate"],
      message: "Choose a valid type.",
    });
  });

  it("extracts a single-segment path", () => {
    expect(
      parseDefinitionFieldError("definition.mode has an invalid or missing type"),
    ).toEqual({ path: ["mode"], message: "Choose a valid type." });
  });

  it("handles the root-level (no nested path) case", () => {
    expect(
      parseDefinitionFieldError(
        "definition enum type requires at least one option",
      ),
    ).toEqual({ path: [], message: "Add at least one option." });
  });

  it("returns null for messages it doesn't recognize", () => {
    expect(parseDefinitionFieldError("Something went wrong")).toBeNull();
    expect(parseDefinitionFieldError("name is required")).toBeNull();
  });
});
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd frontend && npx vitest run src/app/\(authenticated\)/admin/payload-schemas/_components/schema-definition-errors.test.ts`
Expected: FAIL with "Cannot find module './schema-definition-errors'".

- [ ] **Step 3: Write the implementation**

```ts
// frontend/src/app/(authenticated)/admin/payload-schemas/_components/schema-definition-errors.ts

const SUFFIXES: Array<{ suffix: string; message: string }> = [
  { suffix: " has an invalid or missing type", message: "Choose a valid type." },
  {
    suffix: " enum type requires at least one option",
    message: "Add at least one option.",
  },
];

const PREFIX = "definition";

export function parseDefinitionFieldError(
  message: string,
): { path: string[]; message: string } | null {
  for (const { suffix, message: friendly } of SUFFIXES) {
    if (!message.endsWith(suffix)) continue;
    const head = message.slice(0, -suffix.length);
    if (head === PREFIX) return { path: [], message: friendly };
    if (head.startsWith(`${PREFIX}.`)) {
      return { path: head.slice(PREFIX.length + 1).split("."), message: friendly };
    }
  }
  return null;
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd frontend && npx vitest run src/app/\(authenticated\)/admin/payload-schemas/_components/schema-definition-errors.test.ts`
Expected: PASS (4 tests).

- [ ] **Step 5: Commit**

```bash
git add frontend/src/app/\(authenticated\)/admin/payload-schemas/_components/schema-definition-errors.ts frontend/src/app/\(authenticated\)/admin/payload-schemas/_components/schema-definition-errors.test.ts
git commit -m "feat(payload-schemas): parse backend definition validation errors to a field path"
```

---

### Task 3: `SchemaFieldRow` and `SchemaFieldList` (the recursive field editor)

**Files:**
- Create: `frontend/src/app/(authenticated)/admin/payload-schemas/_components/SchemaFieldRow.tsx`
- Create: `frontend/src/app/(authenticated)/admin/payload-schemas/_components/SchemaFieldList.tsx`
- Test: `frontend/src/app/(authenticated)/admin/payload-schemas/_components/SchemaFieldList.test.tsx`

**Interfaces:**
- Consumes: everything from Task 1 (`FieldRow`, `FieldType`, `FIELD_TYPES`, `FIELD_TYPE_LABELS`, `isArrayType`, `isObjectType`, `isEnumType`, `baseType`, `isValidFieldName`, `createEmptyRow`).
- Produces:
  - `SchemaFieldList(props: { rows: FieldRow[]; onChange: (rows: FieldRow[]) => void; path: string[]; errors: Record<string, string> }): JSX.Element` — default export.
  - `SchemaFieldRow(props: { row: FieldRow; path: string[]; errors: Record<string, string>; siblingNames: string[]; onChange: (row: FieldRow) => void; onRemove: () => void }): JSX.Element` — default export.

These two files import each other (`SchemaFieldList` renders `SchemaFieldRow`
per row; `SchemaFieldRow` renders `SchemaFieldList` for Object/List-of-objects
children). This is a legal ES module cycle since neither is invoked at
module-evaluation time — only inside render functions — so import order
between the two files below doesn't matter.

- [ ] **Step 1: Write the failing tests**

```tsx
// frontend/src/app/(authenticated)/admin/payload-schemas/_components/SchemaFieldList.test.tsx
import { cleanup, render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it } from "vitest";
import { useState } from "react";

import SchemaFieldList from "./SchemaFieldList";
import type { FieldRow } from "./schema-definition-types";

function Harness({ initial = [] as FieldRow[] }) {
  const [rows, setRows] = useState<FieldRow[]>(initial);
  return <SchemaFieldList rows={rows} onChange={setRows} path={[]} errors={{}} />;
}

describe("SchemaFieldList", () => {
  afterEach(cleanup);

  it("adds a field row when a valid name is typed and Enter is pressed", async () => {
    const user = userEvent.setup();
    render(<Harness />);

    const input = screen.getByPlaceholderText("variable_name");
    await user.type(input, "sample_rate{Enter}");

    expect(screen.getByText("sample_rate")).toBeInTheDocument();
    expect(screen.getByRole("combobox", { name: "sample_rate type" })).toHaveValue(
      "string",
    );
  });

  it("converts a typed space into an underscore as the user types", async () => {
    const user = userEvent.setup();
    render(<Harness />);

    const input = screen.getByPlaceholderText("variable_name");
    await user.type(input, "sample rate{Enter}");

    expect(screen.getByText("sample_rate")).toBeInTheDocument();
  });

  it("rejects a duplicate sibling name without adding a second row", async () => {
    const user = userEvent.setup();
    render(<Harness initial={[]} />);

    const input = screen.getByPlaceholderText("variable_name");
    await user.type(input, "name{Enter}");
    await user.type(input, "name{Enter}");

    expect(screen.getAllByText("name")).toHaveLength(1);
    expect(
      screen.getByText("A field with this name already exists here."),
    ).toBeInTheDocument();
  });

  it("reveals a nested add-field input when a row's type becomes Object", async () => {
    const user = userEvent.setup();
    render(<Harness />);

    await user.type(
      screen.getByPlaceholderText("variable_name"),
      "location{Enter}",
    );
    await user.selectOptions(
      screen.getByRole("combobox", { name: "location type" }),
      "object",
    );

    const nestedInputs = screen.getAllByPlaceholderText("variable_name");
    expect(nestedInputs).toHaveLength(2);
  });

  it("removes a row when its remove button is clicked", async () => {
    const user = userEvent.setup();
    render(<Harness />);

    await user.type(
      screen.getByPlaceholderText("variable_name"),
      "temp{Enter}",
    );
    expect(screen.getByText("temp")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Remove temp" }));
    expect(screen.queryByText("temp")).not.toBeInTheDocument();
  });

  it("shows Min/Max item and length fields for a List of text row", async () => {
    const user = userEvent.setup();
    render(<Harness />);

    await user.type(screen.getByPlaceholderText("variable_name"), "tags{Enter}");
    await user.selectOptions(
      screen.getByRole("combobox", { name: "tags type" }),
      "[]string",
    );

    expect(screen.getByLabelText("Min items")).toBeInTheDocument();
    expect(screen.getByLabelText("Max items")).toBeInTheDocument();
    expect(screen.getByLabelText("Min length")).toBeInTheDocument();
    expect(screen.getByLabelText("Max length")).toBeInTheDocument();
  });

  it("adds and removes enum options as tags", async () => {
    const user = userEvent.setup();
    render(<Harness />);

    await user.type(screen.getByPlaceholderText("variable_name"), "mode{Enter}");
    await user.selectOptions(
      screen.getByRole("combobox", { name: "mode type" }),
      "enum",
    );

    const optionsInput = screen.getByPlaceholderText(
      "Type an option and press Enter",
    );
    await user.type(optionsInput, "auto{Enter}");
    await user.type(optionsInput, "manual{Enter}");

    expect(screen.getByText("auto")).toBeInTheDocument();
    expect(screen.getByText("manual")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Remove option auto" }));
    expect(screen.queryByText("auto")).not.toBeInTheDocument();
    expect(screen.getByText("manual")).toBeInTheDocument();
  });

  it("toggles the Required checkbox for a row", async () => {
    const user = userEvent.setup();
    render(<Harness />);

    await user.type(screen.getByPlaceholderText("variable_name"), "name{Enter}");
    const row = screen.getByText("name").closest("div")!;
    const checkbox = within(row.parentElement as HTMLElement).getByRole(
      "checkbox",
      { name: "Required" },
    );

    expect(checkbox).not.toBeChecked();
    await user.click(checkbox);
    expect(checkbox).toBeChecked();
  });
});
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd frontend && npx vitest run src/app/\(authenticated\)/admin/payload-schemas/_components/SchemaFieldList.test.tsx`
Expected: FAIL with "Cannot find module './SchemaFieldList'".

- [ ] **Step 3: Write `SchemaFieldRow.tsx`**

```tsx
// frontend/src/app/(authenticated)/admin/payload-schemas/_components/SchemaFieldRow.tsx
"use client";

import { useId, useState } from "react";

import Input from "@/components/ui/input";
import Label from "@/components/ui/label";

import SchemaFieldList from "./SchemaFieldList";
import {
  FIELD_TYPES,
  FIELD_TYPE_LABELS,
  baseType,
  isArrayType,
  isEnumType,
  isObjectType,
  isValidFieldName,
  type FieldRow,
  type FieldType,
} from "./schema-definition-types";

interface SchemaFieldRowProps {
  row: FieldRow;
  path: string[];
  errors: Record<string, string>;
  siblingNames: string[];
  onChange: (row: FieldRow) => void;
  onRemove: () => void;
}

function NumberField({
  id,
  label,
  value,
  onChange,
}: {
  id: string;
  label: string;
  value: string;
  onChange: (value: string) => void;
}) {
  return (
    <div>
      <Label htmlFor={id}>{label}</Label>
      <Input
        id={id}
        type="number"
        value={value}
        onChange={(event) => onChange(event.target.value)}
      />
    </div>
  );
}

function OptionsField({
  id,
  options,
  onChange,
}: {
  id: string;
  options: string[];
  onChange: (options: string[]) => void;
}) {
  const [draft, setDraft] = useState("");
  return (
    <div className="sm:col-span-2">
      <Label htmlFor={id}>Options</Label>
      {options.length ? (
        <div className="mb-2 flex flex-wrap gap-2">
          {options.map((option) => (
            <span
              key={option}
              className="bg-muted flex items-center gap-1.5 rounded-full px-3 py-1 text-sm"
            >
              {option}
              <button
                type="button"
                aria-label={`Remove option ${option}`}
                onClick={() => onChange(options.filter((o) => o !== option))}
                className="text-muted-foreground hover:text-critical"
              >
                ✕
              </button>
            </span>
          ))}
        </div>
      ) : null}
      <Input
        id={id}
        value={draft}
        placeholder="Type an option and press Enter"
        onChange={(event) => setDraft(event.target.value)}
        onKeyDown={(event) => {
          if (event.key !== "Enter") return;
          event.preventDefault();
          const value = draft.trim();
          if (!value || options.includes(value)) return;
          onChange([...options, value]);
          setDraft("");
        }}
      />
    </div>
  );
}

export default function SchemaFieldRow({
  row,
  path,
  errors,
  siblingNames,
  onChange,
  onRemove,
}: SchemaFieldRowProps) {
  const id = useId();
  const [renaming, setRenaming] = useState(false);
  const [nameDraft, setNameDraft] = useState(row.name);
  const [nameError, setNameError] = useState<string>();
  const rowError = errors[path.join(".")];
  const type = baseType(row.type);

  function commitRename() {
    const name = nameDraft.trim();
    if (!name) {
      setNameError("Name is required.");
      return;
    }
    if (!isValidFieldName(name)) {
      setNameError(
        "Use lowercase letters, numbers, and underscores only (e.g. sample_rate).",
      );
      return;
    }
    if (name !== row.name && siblingNames.includes(name)) {
      setNameError("A field with this name already exists here.");
      return;
    }
    onChange({ ...row, name });
    setRenaming(false);
    setNameError(undefined);
  }

  return (
    <div className="border-control-border rounded-xl border p-3">
      <div className="flex flex-wrap items-center gap-3">
        {renaming ? (
          <Input
            autoFocus
            value={nameDraft}
            onChange={(event) =>
              setNameDraft(event.target.value.replace(/ +/g, "_"))
            }
            onBlur={commitRename}
            onKeyDown={(event) => {
              if (event.key === "Enter") {
                event.preventDefault();
                commitRename();
              }
            }}
            className="max-w-48"
          />
        ) : (
          <button
            type="button"
            onClick={() => {
              setNameDraft(row.name);
              setRenaming(true);
            }}
            className="bg-muted rounded-lg px-2.5 py-1 font-mono text-sm font-semibold"
          >
            {row.name}
          </button>
        )}

        <select
          aria-label={`${row.name} type`}
          value={row.type}
          onChange={(event) =>
            onChange({ ...row, type: event.target.value as FieldType })
          }
          className="border-control-border bg-background rounded-lg border px-2.5 py-1.5 text-sm"
        >
          {FIELD_TYPES.map((option) => (
            <option key={option} value={option}>
              {FIELD_TYPE_LABELS[option]}
            </option>
          ))}
        </select>

        <label className="flex items-center gap-1.5 text-sm">
          <input
            type="checkbox"
            checked={row.required}
            onChange={(event) =>
              onChange({ ...row, required: event.target.checked })
            }
            className="accent-primary size-4"
          />
          Required
        </label>

        <button
          type="button"
          aria-label={`Remove ${row.name}`}
          onClick={onRemove}
          className="text-critical hover:bg-critical/10 ml-auto rounded-lg px-2 py-1 text-sm transition-colors"
        >
          ✕
        </button>
      </div>

      {nameError ? (
        <p className="text-critical mt-1 text-sm">{nameError}</p>
      ) : null}
      {rowError ? <p className="text-critical mt-1 text-sm">{rowError}</p> : null}

      <div className="mt-3 grid gap-3 sm:grid-cols-2">
        {isArrayType(row.type) ? (
          <>
            <NumberField
              id={`${id}-min-items`}
              label="Min items"
              value={row.minimumItem}
              onChange={(value) => onChange({ ...row, minimumItem: value })}
            />
            <NumberField
              id={`${id}-max-items`}
              label="Max items"
              value={row.maximumItem}
              onChange={(value) => onChange({ ...row, maximumItem: value })}
            />
          </>
        ) : null}

        {type === "string" ? (
          <>
            <NumberField
              id={`${id}-min-length`}
              label="Min length"
              value={row.minimumLength}
              onChange={(value) => onChange({ ...row, minimumLength: value })}
            />
            <NumberField
              id={`${id}-max-length`}
              label="Max length"
              value={row.maximumLength}
              onChange={(value) => onChange({ ...row, maximumLength: value })}
            />
          </>
        ) : null}

        {type === "float" || type === "integer" ? (
          <>
            <NumberField
              id={`${id}-min`}
              label="Min"
              value={row.minimum}
              onChange={(value) => onChange({ ...row, minimum: value })}
            />
            <NumberField
              id={`${id}-max`}
              label="Max"
              value={row.maximum}
              onChange={(value) => onChange({ ...row, maximum: value })}
            />
            <div>
              <Label htmlFor={`${id}-unit`}>Unit</Label>
              <Input
                id={`${id}-unit`}
                value={row.unit}
                placeholder="e.g. °C"
                onChange={(event) =>
                  onChange({ ...row, unit: event.target.value })
                }
              />
            </div>
          </>
        ) : null}

        {isEnumType(row.type) ? (
          <OptionsField
            id={`${id}-options`}
            options={row.options}
            onChange={(options) => onChange({ ...row, options })}
          />
        ) : null}
      </div>

      {isObjectType(row.type) ? (
        <div className="border-control-border mt-3 space-y-3 border-l-2 pl-4">
          <SchemaFieldList
            rows={row.children}
            path={path}
            errors={errors}
            onChange={(children) => onChange({ ...row, children })}
          />
        </div>
      ) : null}
    </div>
  );
}
```

- [ ] **Step 4: Write `SchemaFieldList.tsx`**

```tsx
// frontend/src/app/(authenticated)/admin/payload-schemas/_components/SchemaFieldList.tsx
"use client";

import { useId, useState } from "react";

import Input from "@/components/ui/input";

import SchemaFieldRow from "./SchemaFieldRow";
import { createEmptyRow, isValidFieldName, type FieldRow } from "./schema-definition-types";

interface SchemaFieldListProps {
  rows: FieldRow[];
  onChange: (rows: FieldRow[]) => void;
  path: string[];
  errors: Record<string, string>;
}

export default function SchemaFieldList({
  rows,
  onChange,
  path,
  errors,
}: SchemaFieldListProps) {
  const [draftName, setDraftName] = useState("");
  const [draftError, setDraftError] = useState<string>();
  const inputId = useId();

  function commitDraft() {
    const name = draftName.trim();
    if (!name) return;
    if (!isValidFieldName(name)) {
      setDraftError(
        "Use lowercase letters, numbers, and underscores only (e.g. sample_rate).",
      );
      return;
    }
    if (rows.some((row) => row.name === name)) {
      setDraftError("A field with this name already exists here.");
      return;
    }
    onChange([...rows, createEmptyRow(name)]);
    setDraftName("");
    setDraftError(undefined);
  }

  return (
    <div className="space-y-3">
      {rows.map((row) => (
        <SchemaFieldRow
          key={row.key}
          row={row}
          path={[...path, row.name]}
          errors={errors}
          siblingNames={rows
            .filter((sibling) => sibling.key !== row.key)
            .map((sibling) => sibling.name)}
          onChange={(next) =>
            onChange(rows.map((sibling) => (sibling.key === row.key ? next : sibling)))
          }
          onRemove={() => onChange(rows.filter((sibling) => sibling.key !== row.key))}
        />
      ))}
      <div>
        <Input
          id={inputId}
          value={draftName}
          placeholder="variable_name"
          aria-label="New field name"
          onChange={(event) => {
            setDraftName(event.target.value.replace(/ +/g, "_"));
            setDraftError(undefined);
          }}
          onKeyDown={(event) => {
            if (event.key === "Enter") {
              event.preventDefault();
              commitDraft();
            }
          }}
        />
        {draftError ? (
          <p className="text-critical mt-1 text-sm">{draftError}</p>
        ) : null}
      </div>
    </div>
  );
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd frontend && npx vitest run src/app/\(authenticated\)/admin/payload-schemas/_components/SchemaFieldList.test.tsx`
Expected: PASS (8 tests). If the "Required checkbox" test fails to locate the
checkbox via `closest("div")!.parentElement`, adjust the query to
`screen.getByText("name").closest(".border-control-border")` instead (the
row's outermost wrapper) — the important assertion is that the checkbox
toggles, not the exact traversal.

- [ ] **Step 6: Typecheck and lint**

Run: `cd frontend && npx tsc --noEmit && npx eslint src/app/\(authenticated\)/admin/payload-schemas/_components/SchemaFieldRow.tsx src/app/\(authenticated\)/admin/payload-schemas/_components/SchemaFieldList.tsx`
Expected: no errors.

- [ ] **Step 7: Commit**

```bash
git add frontend/src/app/\(authenticated\)/admin/payload-schemas/_components/SchemaFieldRow.tsx frontend/src/app/\(authenticated\)/admin/payload-schemas/_components/SchemaFieldList.tsx frontend/src/app/\(authenticated\)/admin/payload-schemas/_components/SchemaFieldList.test.tsx
git commit -m "feat(payload-schemas): add recursive guided field row/list editor"
```

---

### Task 4: `SchemaDefinitionBuilder` (root wrapper, preview, unparseable fallback)

**Files:**
- Create: `frontend/src/app/(authenticated)/admin/payload-schemas/_components/SchemaDefinitionBuilder.tsx`
- Test: `frontend/src/app/(authenticated)/admin/payload-schemas/_components/SchemaDefinitionBuilder.test.tsx`

**Interfaces:**
- Consumes: `rowsFromRootDefinition`, `rootDefinitionFromRows`, `canRepresentDefinition`, `RawDefinition` (Task 1); `SchemaFieldList` (Task 3).
- Produces: `SchemaDefinitionBuilder(props: { name: string; defaultValue?: Record<string, unknown>; errors?: Record<string, string>; onRepresentableChange?: (representable: boolean) => void }): JSX.Element` — default export. Renders a hidden `<input type="hidden" name={name}>` carrying the live-serialized JSON.

- [ ] **Step 1: Write the failing tests**

```tsx
// frontend/src/app/(authenticated)/admin/payload-schemas/_components/SchemaDefinitionBuilder.test.tsx
import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";

import SchemaDefinitionBuilder from "./SchemaDefinitionBuilder";

describe("SchemaDefinitionBuilder", () => {
  afterEach(cleanup);

  it("starts empty with a hidden input carrying an empty object definition", () => {
    const { container } = render(<SchemaDefinitionBuilder name="definition" />);
    const hidden = container.querySelector(
      'input[type="hidden"][name="definition"]',
    ) as HTMLInputElement;
    expect(JSON.parse(hidden.value)).toEqual({ type: "object", properties: {} });
  });

  it("pre-populates rows from an existing representable definition", () => {
    render(
      <SchemaDefinitionBuilder
        name="definition"
        defaultValue={{
          type: "object",
          properties: { sample_rate: { type: "float" } },
        }}
      />,
    );
    expect(screen.getByText("sample_rate")).toBeInTheDocument();
  });

  it("updates the hidden input's JSON as fields are added", async () => {
    const user = userEvent.setup();
    const { container } = render(<SchemaDefinitionBuilder name="definition" />);

    await user.type(screen.getByPlaceholderText("variable_name"), "name{Enter}");

    const hidden = container.querySelector(
      'input[type="hidden"][name="definition"]',
    ) as HTMLInputElement;
    expect(JSON.parse(hidden.value)).toEqual({
      type: "object",
      properties: { name: { type: "string" } },
    });
  });

  it("shows a read-only fallback and reports unrepresentable for a definition it can't parse", () => {
    const onRepresentableChange = vi.fn();
    render(
      <SchemaDefinitionBuilder
        name="definition"
        defaultValue={{ type: "date" }}
        onRepresentableChange={onRepresentableChange}
      />,
    );

    expect(
      screen.getByText(/guided editor can't represent/i),
    ).toBeInTheDocument();
    expect(screen.queryByPlaceholderText("variable_name")).not.toBeInTheDocument();
    expect(onRepresentableChange).toHaveBeenCalledWith(false);
  });

  it("reports representable for a normal definition", () => {
    const onRepresentableChange = vi.fn();
    render(
      <SchemaDefinitionBuilder
        name="definition"
        onRepresentableChange={onRepresentableChange}
      />,
    );
    expect(onRepresentableChange).toHaveBeenCalledWith(true);
  });
});
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd frontend && npx vitest run src/app/\(authenticated\)/admin/payload-schemas/_components/SchemaDefinitionBuilder.test.tsx`
Expected: FAIL with "Cannot find module './SchemaDefinitionBuilder'".

- [ ] **Step 3: Write the implementation**

```tsx
// frontend/src/app/(authenticated)/admin/payload-schemas/_components/SchemaDefinitionBuilder.tsx
"use client";

import { useEffect, useMemo, useState } from "react";

import SchemaFieldList from "./SchemaFieldList";
import {
  canRepresentDefinition,
  rootDefinitionFromRows,
  rowsFromRootDefinition,
  type FieldRow,
  type RawDefinition,
} from "./schema-definition-types";

const EMPTY_DEFINITION: RawDefinition = { type: "object", properties: {} };

interface SchemaDefinitionBuilderProps {
  name: string;
  defaultValue?: Record<string, unknown>;
  errors?: Record<string, string>;
  onRepresentableChange?: (representable: boolean) => void;
}

export default function SchemaDefinitionBuilder({
  name,
  defaultValue,
  errors = {},
  onRepresentableChange,
}: SchemaDefinitionBuilderProps) {
  const initial = (defaultValue as RawDefinition | undefined) ?? EMPTY_DEFINITION;
  const representable = useMemo(() => canRepresentDefinition(initial), [initial]);
  const [rows, setRows] = useState<FieldRow[]>(() =>
    representable ? rowsFromRootDefinition(initial) : [],
  );

  useEffect(() => {
    onRepresentableChange?.(representable);
    // Only re-run if the representability verdict itself changes, not on
    // every keystroke inside `rows` (onRepresentableChange isn't expected to
    // be referentially stable across parent re-renders).
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [representable]);

  if (!representable) {
    return (
      <div className="space-y-2">
        <p className="text-critical text-sm">
          This schema&apos;s definition uses a shape the guided editor can&apos;t
          represent. Edit it directly via the API, or contact an admin.
        </p>
        <pre className="bg-muted overflow-auto rounded-xl p-4 text-xs">
          {JSON.stringify(initial, null, 2)}
        </pre>
      </div>
    );
  }

  const definition = rootDefinitionFromRows(rows);

  return (
    <div className="space-y-3">
      <SchemaFieldList rows={rows} onChange={setRows} path={[]} errors={errors} />
      <input
        type="hidden"
        name={name}
        value={JSON.stringify(definition)}
        readOnly
      />
      <details className="text-sm">
        <summary className="text-muted-foreground cursor-pointer select-none">
          Preview JSON
        </summary>
        <pre className="bg-muted mt-2 overflow-auto rounded-xl p-4 text-xs">
          {JSON.stringify(definition, null, 2)}
        </pre>
      </details>
    </div>
  );
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd frontend && npx vitest run src/app/\(authenticated\)/admin/payload-schemas/_components/SchemaDefinitionBuilder.test.tsx`
Expected: PASS (5 tests).

- [ ] **Step 5: Typecheck and lint**

Run: `cd frontend && npx tsc --noEmit && npx eslint src/app/\(authenticated\)/admin/payload-schemas/_components/SchemaDefinitionBuilder.tsx`
Expected: no errors.

- [ ] **Step 6: Commit**

```bash
git add frontend/src/app/\(authenticated\)/admin/payload-schemas/_components/SchemaDefinitionBuilder.tsx frontend/src/app/\(authenticated\)/admin/payload-schemas/_components/SchemaDefinitionBuilder.test.tsx
git commit -m "feat(payload-schemas): add schema definition builder wrapper with preview and fallback"
```

---

### Task 5: Wire the builder into `PayloadSchemaForm.tsx`

**Files:**
- Modify: `frontend/src/app/(authenticated)/admin/payload-schemas/_components/PayloadSchemaForm.tsx`

**Interfaces:**
- Consumes: `SchemaDefinitionBuilder` (Task 4), `parseDefinitionFieldError` (Task 2).

The `SchemaEditor` dialog currently renders `JsonEditor` for the `definition`
field (lines 137-143 of the current file) and has no mechanism to disable
the rest of the form. This task replaces the JSON editor and adds the
representable/disabled wiring described in the spec's Validation & error
handling and Editing-an-unrepresentable-schema sections.

- [ ] **Step 1: Replace the import**

In `PayloadSchemaForm.tsx`, replace:

```tsx
import JsonEditor from "@/components/json/JsonEditor";
```

with:

```tsx
import { useMemo, useState } from "react";

import { parseDefinitionFieldError } from "./schema-definition-errors";
import SchemaDefinitionBuilder from "./SchemaDefinitionBuilder";
```

(`useState`/`useMemo` are new imports alongside the existing
`useActionState, useEffect` from `"react"` — merge them into the single
existing `react` import rather than adding a second one.)

- [ ] **Step 2: Compute nested field errors and representability inside `SchemaEditor`**

Immediately after the existing `useEffect` that calls `router.refresh()` on
success (around line 94), add:

```tsx
const definitionErrors = useMemo(() => {
  if (state.status !== "error" || !state.message) return {};
  const parsed = parseDefinitionFieldError(state.message);
  return parsed ? { [parsed.path.join(".")]: parsed.message } : {};
}, [state]);
const [definitionRepresentable, setDefinitionRepresentable] = useState(true);
```

- [ ] **Step 3: Replace the `JsonEditor` usage and wrap the disable-able fields**

Replace:

```tsx
      <form action={action} className="space-y-4">
        {schema ? (
          <input type="hidden" name="payload_schema_id" value={schema.id} />
        ) : null}
        <div>
          <Label htmlFor="schema-name">Name</Label>
          <Input
            id="schema-name"
            name="name"
            defaultValue={schema?.name}
            required
          />
          {state.fieldErrors?.name ? (
            <p className="text-critical text-sm">{state.fieldErrors.name}</p>
          ) : null}
        </div>
        <div>
          <Label htmlFor="schema-version">Version</Label>
          <Input
            id="schema-version"
            name="version"
            type="number"
            min={1}
            step={1}
            defaultValue={schema?.version ?? 1}
            required
          />
          {state.fieldErrors?.version ? (
            <p className="text-critical text-sm">{state.fieldErrors.version}</p>
          ) : null}
        </div>
        <JsonEditor
          name="definition"
          label="Definition"
          defaultValue={
            schema?.definition ?? { type: "object", properties: {} }
          }
        />
        <div className="grid gap-4 sm:grid-cols-2">
          <div>
            <Label htmlFor="schema-valid-from">Valid from</Label>
            <Input
              id="schema-valid-from"
              name="valid_from"
              type="datetime-local"
              defaultValue={localDate(schema?.valid_from)}
            />
          </div>
          <div>
            <Label htmlFor="schema-valid-to">Valid to</Label>
            <Input
              id="schema-valid-to"
              name="valid_to"
              type="datetime-local"
              defaultValue={localDate(schema?.valid_to)}
            />
            {state.fieldErrors?.valid_to ? (
              <p className="text-critical text-sm">
                {state.fieldErrors.valid_to}
              </p>
            ) : null}
          </div>
        </div>
```

with:

```tsx
      <form action={action} className="space-y-4">
        {schema ? (
          <input type="hidden" name="payload_schema_id" value={schema.id} />
        ) : null}
        <fieldset disabled={!definitionRepresentable} className="space-y-4">
          <div>
            <Label htmlFor="schema-name">Name</Label>
            <Input
              id="schema-name"
              name="name"
              defaultValue={schema?.name}
              required
            />
            {state.fieldErrors?.name ? (
              <p className="text-critical text-sm">{state.fieldErrors.name}</p>
            ) : null}
          </div>
          <div>
            <Label htmlFor="schema-version">Version</Label>
            <Input
              id="schema-version"
              name="version"
              type="number"
              min={1}
              step={1}
              defaultValue={schema?.version ?? 1}
              required
            />
            {state.fieldErrors?.version ? (
              <p className="text-critical text-sm">
                {state.fieldErrors.version}
              </p>
            ) : null}
          </div>
          <div>
            <Label>Definition</Label>
            <SchemaDefinitionBuilder
              name="definition"
              defaultValue={schema?.definition}
              errors={definitionErrors}
              onRepresentableChange={setDefinitionRepresentable}
            />
          </div>
          <div className="grid gap-4 sm:grid-cols-2">
            <div>
              <Label htmlFor="schema-valid-from">Valid from</Label>
              <Input
                id="schema-valid-from"
                name="valid_from"
                type="datetime-local"
                defaultValue={localDate(schema?.valid_from)}
              />
            </div>
            <div>
              <Label htmlFor="schema-valid-to">Valid to</Label>
              <Input
                id="schema-valid-to"
                name="valid_to"
                type="datetime-local"
                defaultValue={localDate(schema?.valid_to)}
              />
              {state.fieldErrors?.valid_to ? (
                <p className="text-critical text-sm">
                  {state.fieldErrors.valid_to}
                </p>
              ) : null}
            </div>
          </div>
        </fieldset>
```

- [ ] **Step 4: Disable Save while the loaded schema can't be represented**

Replace:

```tsx
          <Button type="submit" disabled={pending}>
            {pending ? "Saving…" : "Save schema"}
          </Button>
```

with:

```tsx
          <Button type="submit" disabled={pending || !definitionRepresentable}>
            {pending ? "Saving…" : "Save schema"}
          </Button>
```

- [ ] **Step 5: Typecheck and lint**

Run: `cd frontend && npx tsc --noEmit && npx eslint src/app/\(authenticated\)/admin/payload-schemas/_components/PayloadSchemaForm.tsx`
Expected: no errors. (`JsonEditor` import removal must not leave it referenced
anywhere else in this file — grep to confirm: `grep -n JsonEditor
frontend/src/app/\(authenticated\)/admin/payload-schemas/_components/PayloadSchemaForm.tsx`
should return nothing.)

- [ ] **Step 6: Commit**

```bash
git add frontend/src/app/\(authenticated\)/admin/payload-schemas/_components/PayloadSchemaForm.tsx
git commit -m "feat(payload-schemas): use the guided builder in the create/edit dialog"
```

---

### Task 6: List page filter simplification

**Files:**
- Modify: `frontend/src/app/(authenticated)/admin/payload-schemas/page.tsx`

- [ ] **Step 1: Remove the lookup form and its redirect handling**

Remove the `getLatestPayloadSchema`/`getPayloadSchemaByNameAndVersion` import
names from the `@/lib/api/payload-schemas` import (keep `listPayloadSchemas`),
and delete this whole block:

```tsx
  const lookupName = String(raw.lookup_name ?? "").trim();
  const lookupVersion = String(raw.lookup_version ?? "").trim();
  if (lookupName) {
    const found = lookupVersion
      ? await getPayloadSchemaByNameAndVersion(
          lookupName,
          Number(lookupVersion),
        )
      : await getLatestPayloadSchema(lookupName);
    redirect(`/admin/payload-schemas/${found.id}`);
  }
```

and delete the second `<form>` (the "Open exact/latest" one):

```tsx
        <form
          action="/admin/payload-schemas"
          className="border-border mt-4 grid gap-3 border-t pt-4 md:grid-cols-[1fr_10rem_auto]"
        >
          <Input
            name="lookup_name"
            placeholder="Exact schema name"
            aria-label="Schema lookup name"
            required
          />
          <Input
            name="lookup_version"
            type="number"
            min={1}
            placeholder="Latest"
            aria-label="Schema version"
          />
          <Button type="submit" variant="secondary">
            Open exact/latest
          </Button>
        </form>
```

- [ ] **Step 2: Replace the `valid_at` datetime filter with a "Still valid" checkbox**

Replace:

```tsx
  const base = parsePageQuery(raw);
  const rawValidAt = String(raw.valid_at ?? "").trim();
  const validAt =
    rawValidAt && !Number.isNaN(new Date(rawValidAt).valueOf())
      ? new Date(rawValidAt).toISOString()
      : undefined;
  const result = await listPayloadSchemas({ ...base, valid_at: validAt });
```

with:

```tsx
  const base = parsePageQuery(raw);
  const stillValid = String(raw.still_valid ?? "") === "1";
  const validAt = stillValid ? new Date().toISOString() : undefined;
  const result = await listPayloadSchemas({ ...base, valid_at: validAt });
```

Replace the first `<form>`'s contents:

```tsx
        <form
          action="/admin/payload-schemas"
          className="grid gap-3 md:grid-cols-[1fr_1fr_auto]"
        >
          <Input
            name="search"
            type="search"
            defaultValue={base.search}
            placeholder="Search schema names"
            aria-label="Search schemas"
          />
          <Input
            name="valid_at"
            type="datetime-local"
            defaultValue={rawValidAt}
            aria-label="Valid at"
          />
          <Button type="submit" className="gap-2">
            <Search className="size-4" aria-hidden="true" />
            Filter
          </Button>
        </form>
```

with:

```tsx
        <form
          action="/admin/payload-schemas"
          className="flex flex-wrap items-center gap-3"
        >
          <Input
            name="search"
            type="search"
            defaultValue={base.search}
            placeholder="Search schema names"
            aria-label="Search schemas"
            className="min-w-0 flex-1"
          />
          <label className="flex items-center gap-1.5 text-sm whitespace-nowrap">
            <input
              type="checkbox"
              name="still_valid"
              value="1"
              defaultChecked={stillValid}
              className="accent-primary size-4"
            />
            Still valid
          </label>
          <Button type="submit" className="gap-2">
            <Search className="size-4" aria-hidden="true" />
            Filter
          </Button>
        </form>
```

- [ ] **Step 3: Update empty-state and pagination query-string references**

Replace every remaining use of `validAt` used purely as a "were filters
applied" boolean, and the `rawValidAt` pagination param, so they refer to
`stillValid` instead:

```tsx
      {result.data.length ? (
        <section className="grid gap-5 sm:grid-cols-2 xl:grid-cols-3">
          {result.data.map((schema) => (
            <PayloadSchemaCard key={schema.id} schema={schema} />
          ))}
        </section>
      ) : (
        <EmptyState
          title={
            base.search || stillValid
              ? "No matching schemas"
              : "No payload schemas yet"
          }
          description={
            base.search || stillValid
              ? "No schema version matches these filters."
              : "Create a versioned schema for validated action payloads."
          }
          action={
            base.search || stillValid ? (
              <Link
                href="/admin/payload-schemas"
                className="text-primary font-semibold"
              >
                Clear filters
              </Link>
            ) : undefined
          }
        />
      )}
      <Pagination
        page={result.page}
        pathname="/admin/payload-schemas"
        searchParams={{
          limit: String(base.limit),
          search: base.search,
          still_valid: stillValid ? "1" : undefined,
        }}
      />
```

- [ ] **Step 4: Typecheck and lint**

Run: `cd frontend && npx tsc --noEmit && npx eslint src/app/\(authenticated\)/admin/payload-schemas/page.tsx`
Expected: no errors. Confirm `redirect` is still used elsewhere in the file
(the out-of-range-page redirect) so its import isn't now unused; confirm
`getLatestPayloadSchema`/`getPayloadSchemaByNameAndVersion` are no longer
imported here.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/app/\(authenticated\)/admin/payload-schemas/page.tsx
git commit -m "feat(payload-schemas): replace lookup/date filters with a Still valid checkbox"
```

---

### Task 7: Full-suite verification and a live smoke test

**Files:** none (verification only).

- [ ] **Step 1: Run the full frontend test suite**

Run: `cd frontend && npx vitest run`
Expected: all tests pass, including the new ones from Tasks 1-4 and every
pre-existing test file.

- [ ] **Step 2: Full typecheck and lint**

Run: `cd frontend && npx tsc --noEmit && npx eslint src`
Expected: no errors (warnings pre-existing elsewhere in the repo, e.g.
`RoleCard.tsx`'s unused var, are fine and unrelated to this change — only
check that nothing new appears in the files this plan touched).

- [ ] **Step 3: Build the Docker image and start it**

Run: `docker compose build app && docker compose up -d --force-recreate app`
Wait ~5s, then confirm it's serving:
Run: `curl -s -o /dev/null -w "%{http_code}\n" http://127.0.0.1:<mapped-port>/login`
Expected: `200`. (Check `compose.yml`'s `ports:` mapping for the current host
port — it has changed between sessions, e.g. `60001:80`.)

- [ ] **Step 4: Smoke-test the guided builder end-to-end with Playwright**

Using this repo's established Docker+Playwright pattern (a `mate-browser-test`
container with `node`/`playwright` already installed, reachable via
`172.17.0.1:<port>` from inside that container), log in as the seeded admin
(`admin` / whatever the current seeded password is — check
`backend/database/seeder/user.json`, and if it's been changed via direct DB
update in a prior session, reset it or ask the user), then:

1. Navigate to `/admin/payload-schemas`, confirm no "Open exact/latest" form
   and no date input are present, and that a "Still valid" checkbox is
   present next to the search field.
2. Click "Create schema", type a name/version, then in the guided builder:
   type `sample_rate` + Enter, change its type to "Number", fill Min=0
   Max=100, set Unit=`Hz`, check Required; type `mode` + Enter, change type
   to "Choice", add two options via the tag input; type `location` + Enter,
   change type to "Object", then inside its nested list type `lat` + Enter
   and `lon` + Enter. Expand "Preview JSON" and confirm it shows the expected
   nested structure with `sample_rate`/`mode`/`location.lat`/`location.lon`
   and `required: ["sample_rate"]`.
3. Submit, confirm the dialog closes and a success toast appears (per the
   already-fixed `router.refresh()`-based closing behavior from the prior
   session's dialog fixes — do not reintroduce the old JSON-hang bug).
4. Reopen the same schema for edit, confirm the guided builder repopulates
   with `sample_rate`/`mode`/`location` exactly as entered (round-trip
   check).
5. Check the "Still valid" checkbox and submit the filter form; confirm the
   URL gains `?still_valid=1` and the list only shows currently-valid
   versions.

If any step fails, fix the specific task's code before moving on — do not
patch around it with an unrelated workaround.

- [ ] **Step 5: Final commit (if any smoke-test fixes were needed)**

```bash
git add -A
git commit -m "fix(payload-schemas): address issues found during builder smoke test"
```

(Skip this step entirely if Step 4 required no code changes.)
