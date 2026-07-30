import { describe, expect, it } from "vitest";

import { getOutOfRangePageRedirect, parsePageQuery } from "./collection-query";

describe("parsePageQuery", () => {
  it("normalizes invalid page values", () => {
    expect(
      parsePageQuery({ page: "-2", limit: "999", search: " sensor " }),
    ).toEqual({
      page: 1,
      limit: 24,
      search: "sensor",
    });
  });

  it("keeps supported limits and trims an empty search away", () => {
    expect(parsePageQuery({ page: "3", limit: "48", search: "   " })).toEqual({
      page: 3,
      limit: 48,
      search: undefined,
    });
  });

  it("normalizes repeated page and limit values instead of selecting the first", () => {
    expect(
      parsePageQuery({
        page: ["2", "3"],
        limit: ["24", "48"],
        search: ["sensor", "gateway"],
      }),
    ).toEqual({
      page: 1,
      limit: 12,
      search: "sensor,gateway",
    });
  });

  it.each([
    "0",
    "-1",
    "1.5",
    "Infinity",
    "9007199254740992",
    "1e2",
    "0x10",
    " 2 ",
    "+2",
  ])("accepts only positive safe integer pages and rejects %s", (page) => {
    expect(parsePageQuery({ page })).toMatchObject({ page: 1 });
  });

  it("accepts the largest positive safe integer page", () => {
    expect(parsePageQuery({ page: "9007199254740991" })).toMatchObject({
      page: Number.MAX_SAFE_INTEGER,
    });
  });

  it("rejects an array-shaped limit even when it contains one supported value", () => {
    expect(parsePageQuery({ limit: ["48"] })).toMatchObject({ limit: 12 });
  });

  it("rejects non-decimal limit syntax before applying supported-limit fallback", () => {
    expect(parsePageQuery({ limit: "0x18" })).toMatchObject({ limit: 12 });
  });
});

describe("getOutOfRangePageRedirect", () => {
  it("clamps to the last page while preserving active filters", () => {
    const target = getOutOfRangePageRedirect(
      "/nodes",
      {
        page: "9",
        limit: "12",
        search: "freezer",
        node_class_id: "class-1",
      },
      { page: 9, limit: 12, total_items: 25 },
    );

    const url = new URL(target ?? "", "https://example.test");
    expect(url.pathname).toBe("/nodes");
    expect(url.searchParams.get("page")).toBe("3");
    expect(url.searchParams.get("limit")).toBe("12");
    expect(url.searchParams.get("search")).toBe("freezer");
    expect(url.searchParams.get("node_class_id")).toBe("class-1");
  });

  it("does not redirect an initially empty collection", () => {
    expect(
      getOutOfRangePageRedirect(
        "/nodes",
        { page: "9", search: "freezer" },
        { page: 9, limit: 12, total_items: 0 },
      ),
    ).toBeUndefined();
  });
});
