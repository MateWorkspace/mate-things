import { describe, expect, it } from "vitest";

import { canAccessAny, hasPermission } from "./permissions";

describe("permission helpers", () => {
  it("checks an exact effective permission", () => {
    const permissions = new Set(["node:get"]);

    expect(hasPermission(permissions, "node:get")).toBe(true);
    expect(hasPermission(permissions, "node:set")).toBe(false);
  });

  it("allows public requirements and any matching required permission", () => {
    const permissions = new Set(["firmware:get"]);

    expect(canAccessAny(permissions, [])).toBe(true);
    expect(canAccessAny(permissions, ["node:get", "firmware:get"])).toBe(true);
    expect(canAccessAny(permissions, ["node:get", "node_class:get"])).toBe(
      false,
    );
  });
});
