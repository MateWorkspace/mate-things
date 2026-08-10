import { describe, expect, it } from "vitest";

import {
  buildCollectionUrl,
  buildOutOfRangeRedirect,
  firstQueryValue,
  parseBooleanQuery,
  parseEnumQuery,
  parseLocalDateTime,
  parsePositiveSafeInteger,
  toUtcQueryValue,
} from "./query";

describe("query helpers", () => {
  it("selects only the first value from a repeated query parameter", () => {
    expect(firstQueryValue(["first", "second"])).toBe("first");
    expect(firstQueryValue("only")).toBe("only");
    expect(firstQueryValue(undefined)).toBeUndefined();
  });

  it.each([
    ["0"],
    ["-1"],
    ["1.5"],
    [String(Number.MAX_SAFE_INTEGER + 1)],
    [0],
    [-1],
    [Number.MAX_SAFE_INTEGER + 1],
    [undefined],
  ])("uses the fallback for a non-positive or unsafe integer %j", (value) => {
    expect(parsePositiveSafeInteger(value, 12)).toBe(12);
  });

  it.each([
    ["24", 24],
    [24, 24],
    [String(Number.MAX_SAFE_INTEGER), Number.MAX_SAFE_INTEGER],
  ])("accepts a positive safe integer %j", (value, expected) => {
    expect(parsePositiveSafeInteger(value, 12)).toBe(expected);
  });

  it("matches enum values exactly", () => {
    const values = ["roles", "permissions"] as const;

    expect(parseEnumQuery("permissions", values)).toBe("permissions");
    expect(parseEnumQuery("Permissions", values)).toBeUndefined();
    expect(parseEnumQuery(" permissions ", values)).toBeUndefined();
    expect(parseEnumQuery(["roles"], values)).toBeUndefined();
  });

  it("parses booleans strictly", () => {
    expect(parseBooleanQuery("true")).toBe(true);
    expect(parseBooleanQuery("false")).toBe(false);
    expect(parseBooleanQuery("1")).toBeUndefined();
    expect(parseBooleanQuery(true)).toBeUndefined();
  });

  it("converts a valid local datetime and rejects invalid input", () => {
    const parsed = parseLocalDateTime("2026-08-10T14:30");

    expect(parsed?.getTime()).toBe(new Date(2026, 7, 10, 14, 30).getTime());
    expect(parseLocalDateTime("not-a-date")).toBeUndefined();
    expect(parseLocalDateTime(123)).toBeUndefined();
  });

  it("formats a date as a canonical UTC query value", () => {
    expect(toUtcQueryValue(new Date("2026-08-10T14:30:45.123Z"))).toBe(
      "2026-08-10T14:30:45.123Z",
    );
  });

  it("builds an encoded collection URL while filtering only undefined values", () => {
    expect(
      buildCollectionUrl("/nodes", {
        search: "cold storage & lab",
        node_class_id: undefined,
        tag: "",
      }),
    ).toBe("/nodes?search=cold+storage+%26+lab&tag=");
    expect(buildCollectionUrl("/nodes", { search: undefined })).toBe("/nodes");
  });

  it("preserves active repeated filters and meaningful empty values in redirects", () => {
    expect(
      buildOutOfRangeRedirect(
        "/nodes",
        {
          page: "99",
          search: "cold storage",
          status: ["", "connected"],
        },
        4,
      ),
    ).toBe("/nodes?search=cold+storage&status=&status=connected&page=4");
  });
});
