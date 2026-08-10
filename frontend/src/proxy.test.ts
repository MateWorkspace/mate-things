import { describe, expect, it } from "vitest";

import { isProtectedRoute } from "./proxy";

describe("proxy route protection", () => {
  it.each([
    "/dashboard",
    "/nodes/node-123",
    "/node-classes/class-123",
    "/firmware/firmware-123",
    "/actions/action-123",
    "/action-history",
    "/ble-direct",
    "/telemetry",
    "/node-logs",
    "/broadcast-sessions/live",
    "/admin",
    "/admin/users/user-123",
    "/admin/api-keys",
    "/admin/access-control",
    "/admin/payload-schemas/schema-123",
  ])("refreshes the session for protected path %s", (pathname) => {
    expect(isProtectedRoute(pathname)).toBe(true);
  });
});
