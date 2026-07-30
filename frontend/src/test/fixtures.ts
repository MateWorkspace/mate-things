import type { NodeResponse, UserResponse } from "@/lib/api";

export const USER: UserResponse = {
  id: "user-1",
  role_id: "role-1",
  name: "Alex Morgan",
  bio: "Fleet operator",
  username: "alex",
  preferences: {},
  created_at: "2026-07-30T00:00:00Z",
};

export function nodeFixture(
  overrides: Partial<NodeResponse> = {},
): NodeResponse {
  return {
    id: "node-1",
    node_class_id: "class-1",
    device_id: "AC276E5E030C",
    device_info: "ESP32",
    name: "Cold Storage Sensor 07",
    firmware_id: "firmware-1",
    description: "Freezer room sensor",
    is_connected: false,
    preferences: {},
    created_at: "2026-07-30T00:00:00Z",
    ...overrides,
  };
}

export const EMPTY_STATE = { status: "idle" } as const;
