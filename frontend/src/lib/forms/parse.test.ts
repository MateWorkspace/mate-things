import { describe, expect, it } from "vitest";

import {
  formText,
  optionalDateTime,
  optionalString,
  parseJsonObject,
  requiredString,
} from "./parse";

describe("form parsing", () => {
  it("preserves a raw text value while exposing its trimmed value", () => {
    const data = new FormData();
    data.set("name", "  Espresso lab  ");

    expect(requiredString(data, "name")).toEqual({
      ok: true,
      raw: "  Espresso lab  ",
      value: "Espresso lab",
    });
    expect(formText(data, "name", { trim: false })).toBe(
      "  Espresso lab  ",
    );
  });

  it("rejects non-text and blank required values without losing raw input", () => {
    const data = new FormData();
    data.set("name", "   ");
    data.set("file", new File(["firmware"], "firmware.bin"));

    expect(requiredString(data, "name")).toEqual({
      ok: false,
      raw: "   ",
      error: "This field is required.",
    });
    expect(formText(data, "file")).toBe("");
  });

  it("normalizes optional strings and datetimes", () => {
    const data = new FormData();
    data.set("blank", "   ");
    data.set("expires_at", "2026-08-10T12:30");

    expect(optionalString(data, "blank")).toBeUndefined();
    expect(optionalDateTime(data, "blank")).toEqual({
      ok: true,
      raw: "   ",
      value: undefined,
    });
    expect(optionalDateTime(data, "expires_at")).toEqual({
      ok: true,
      raw: "2026-08-10T12:30",
      value: new Date("2026-08-10T12:30").toISOString(),
    });
  });

  it("parses JSON objects and reports arrays or malformed input", () => {
    expect(parseJsonObject('{"enabled":true}')).toEqual({
      ok: true,
      raw: '{"enabled":true}',
      value: { enabled: true },
    });
    expect(parseJsonObject("[]")).toMatchObject({
      ok: false,
      error: "Enter a JSON object.",
    });
    expect(parseJsonObject("{")) .toMatchObject({
      ok: false,
      error: "Enter valid JSON.",
    });
  });
});
