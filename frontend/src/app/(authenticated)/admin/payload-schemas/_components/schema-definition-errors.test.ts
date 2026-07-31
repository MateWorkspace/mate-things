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
