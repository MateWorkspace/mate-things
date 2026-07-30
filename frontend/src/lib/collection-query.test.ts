import { describe, expect, it } from "vitest";

import { parsePageQuery } from "./collection-query";

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
});
