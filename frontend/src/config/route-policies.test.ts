import { describe, expect, it } from "vitest";

import {
  canVisitRoute,
  findRoutePolicy,
  isProtectedRoute,
} from "./route-policies";

describe("route policies", () => {
  it.each([
    ["/ble-direct", true],
    ["/broadcast-sessions/live", true],
    ["/admin/api-keys", true],
    ["/login", false],
  ])("classifies %s", (pathname, expected) => {
    expect(isProtectedRoute(pathname)).toBe(expected);
  });

  it.each([
    ["/dashboard", "Fleet Overview", []],
    ["/nodes/node-123", "Nodes", ["node:get"]],
    ["/node-classes/class-123", "Node Classes", ["node_class:get"]],
    ["/firmware/firmware-123", "Firmware", ["firmware:get"]],
    ["/actions/action-123", "Actions", ["action:get"]],
    ["/action-history", "Action History", ["action_log:get"]],
    ["/ble-direct", "BLE Direct", []],
    ["/telemetry", "Telemetry", ["telemetry_record:get"]],
    ["/node-logs", "Node Logs", ["node_log:get"]],
    [
      "/broadcast-sessions/live",
      "Broadcast Sessions",
      ["broadcast_session:get"],
    ],
    ["/admin/users/user-123", "Users", ["user:get"]],
    ["/admin/api-keys", "API Keys", ["api_key:get"]],
    [
      "/admin/access-control",
      "Access Control",
      ["role:get", "permission:get", "role_permission:get"],
    ],
    [
      "/admin/payload-schemas/schema-123",
      "Payload Schemas",
      ["payload_schema:get"],
    ],
  ])(
    "resolves %s to its navigation permission rule",
    (pathname, label, requiredAny) => {
      expect(findRoutePolicy(pathname)).toMatchObject({
        navigation: { label },
        requiredAny,
      });
    },
  );

  it("does not treat similarly named paths as a policy match", () => {
    expect(findRoutePolicy("/nodes-old")).toBeUndefined();
  });

  it("allows a route when any required permission is present", () => {
    expect(
      canVisitRoute("/admin/access-control", new Set(["permission:get"])),
    ).toBe(true);
    expect(canVisitRoute("/admin/access-control", new Set(["user:get"]))).toBe(
      false,
    );
  });
});
