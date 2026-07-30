import { describe, expect, it } from "vitest";

import { isProtectedRoute } from "./proxy";

describe("isProtectedRoute", () => {
  it.each([
    "/dashboard",
    "/nodes",
    "/firmware/abc",
    "/actions",
    "/telemetry",
    "/node-logs",
    "/admin/users",
  ])("treats %s as protected", (pathname) => {
    expect(isProtectedRoute(pathname)).toBe(true);
  });

  it("does not protect login", () => {
    expect(isProtectedRoute("/login")).toBe(false);
  });
});
