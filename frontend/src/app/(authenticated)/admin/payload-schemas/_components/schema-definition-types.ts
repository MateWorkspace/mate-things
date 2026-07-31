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
