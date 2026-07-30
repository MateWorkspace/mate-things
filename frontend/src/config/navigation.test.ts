import { describe, expect, it } from "vitest";

import { visibleNavigation } from "./navigation";

describe("visibleNavigation", () => {
  it("omits empty navigation groups", () => {
    const visible = visibleNavigation(new Set(["node:get"]));

    expect(visible.map((group) => group.label)).toEqual(["Overview", "Fleet"]);
    expect(visible[1]?.items.map((item) => item.label)).toEqual(["Nodes"]);
  });

  it("shows Access Control for any access-control read permission", () => {
    const visible = visibleNavigation(new Set(["role:get"]));

    expect(
      visible.flatMap((group) => group.items).map((item) => item.label),
    ).toContain("Access Control");
  });
});
