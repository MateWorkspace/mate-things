import { describe, expect, it } from "vitest";

import { canVisit, requiredPermissions } from "./route-access";

describe("route access", () => {
  it("maps collection and detail routes to the same read permission", () => {
    expect(requiredPermissions("/nodes")).toEqual(["node:get"]);
    expect(requiredPermissions("/nodes/node-42?tab=firmware")).toEqual([
      "node:get",
    ]);
  });

  it("accepts any access-control read permission", () => {
    expect(canVisit("/admin/access-control", new Set(["permission:get"]))).toBe(
      true,
    );
    expect(canVisit("/admin/access-control", new Set(["node:get"]))).toBe(
      false,
    );
  });

  it("keeps the fleet overview available to every authenticated user", () => {
    expect(canVisit("/dashboard", new Set())).toBe(true);
  });
});
