import { describe, expect, it } from "vitest";

import {
  canRepresentDefinition,
  createEmptyRow,
  isRowTreeValid,
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

  it("rejects a non-enum type carrying a spurious options array", () => {
    const def = { type: "string", options: ["a", "b"] };
    expect(canRepresentDefinition(def)).toBe(false);
  });

  it("rejects a non-object type carrying spurious properties or required", () => {
    expect(
      canRepresentDefinition({ type: "string", properties: {} }),
    ).toBe(false);
    expect(
      canRepresentDefinition({ type: "string", required: ["name"] }),
    ).toBe(false);
  });

  it("rejects a definition with a non-numeric minimum", () => {
    const def = { type: "float", minimum: "0" };
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

describe("isRowTreeValid", () => {
  it("is valid for an empty tree and for rows with no constraints", () => {
    expect(isRowTreeValid([])).toBe(true);
    expect(isRowTreeValid([createEmptyRow("name")])).toBe(true);
  });

  it("is invalid when an enum row has zero options", () => {
    const row = { ...createEmptyRow("mode"), type: "enum" as const };
    expect(isRowTreeValid([row])).toBe(false);
  });

  it("is valid once the enum row has at least one option", () => {
    const row = {
      ...createEmptyRow("mode"),
      type: "enum" as const,
      options: ["auto"],
    };
    expect(isRowTreeValid([row])).toBe(true);
  });

  it("is invalid when minimum exceeds maximum", () => {
    const row = {
      ...createEmptyRow("rate"),
      type: "float" as const,
      minimum: "100",
      maximum: "0",
    };
    expect(isRowTreeValid([row])).toBe(false);
  });

  it("is invalid when a nested child row is invalid", () => {
    const child = { ...createEmptyRow("mode"), type: "enum" as const };
    const parent = {
      ...createEmptyRow("location"),
      type: "object" as const,
      children: [child],
    };
    expect(isRowTreeValid([parent])).toBe(false);
  });

  it("ignores stale children left behind after a row is switched away from object", () => {
    const invalidChild = { ...createEmptyRow("mode"), type: "enum" as const };
    const row = {
      ...createEmptyRow("location"),
      type: "string" as const,
      children: [invalidChild],
    };
    expect(isRowTreeValid([row])).toBe(true);
  });
});
