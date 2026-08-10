import { beforeEach, describe, expect, it, vi } from "vitest";

import { listAllActions, type ActionResponse } from "./actions";
import { apiFetch } from "./client";
import {
  listAllNodeClasses,
  listAllNodeClassActions,
  type NodeClassActionDetailResponse,
  type NodeClassResponse,
} from "./node-classes";
import { listAllNodes, type NodeResponse } from "./nodes";
import {
  listAllPayloadSchemas,
  type PayloadSchemaResponse,
} from "./payload-schemas";
import { listAllPermissions, type PermissionResponse } from "./permissions";
import { listAllRoles, type RoleResponse } from "./roles";

vi.mock("server-only", () => ({}));
vi.mock("./client", async (importOriginal) => {
  const original = await importOriginal<typeof import("./client")>();
  return { ...original, apiFetch: vi.fn() };
});

const AUDIT = { created_at: "2026-08-10T00:00:00Z" };

const ACTIONS: ActionResponse[] = [
  {
    ...AUDIT,
    id: "action-1",
    name: "Restart",
    description: "Restart node",
    payload_schema_name: "empty",
    payload_schema_version: 1,
    preferences: {},
  },
  {
    ...AUDIT,
    id: "action-2",
    name: "Ping",
    description: "Ping node",
    payload_schema_name: "empty",
    payload_schema_version: 1,
    preferences: {},
  },
];

const NODE_CLASSES: NodeClassResponse[] = [
  {
    ...AUDIT,
    id: "class-1",
    name: "Sensors",
    description: "Sensors",
    preferences: {},
  },
  {
    ...AUDIT,
    id: "class-2",
    name: "Relays",
    description: "Relays",
    preferences: {},
  },
];

const NODES: NodeResponse[] = [
  {
    ...AUDIT,
    id: "node-1",
    node_class_id: "class-1",
    device_id: "esp32-1",
    device_info: "ESP32",
    name: "Sensor one",
    firmware_id: "firmware-1",
    description: "",
    is_connected: true,
    preferences: {},
  },
  {
    ...AUDIT,
    id: "node-2",
    node_class_id: "class-1",
    device_id: "esp32-2",
    device_info: "ESP32",
    name: "Sensor two",
    firmware_id: "firmware-1",
    description: "",
    is_connected: true,
    preferences: {},
  },
];

const ROLES: RoleResponse[] = [
  {
    ...AUDIT,
    id: "role-1",
    name: "Operator",
    description: "Operator",
    is_default: false,
    preferences: {},
  },
  {
    ...AUDIT,
    id: "role-2",
    name: "Admin",
    description: "Admin",
    is_default: true,
    preferences: {},
  },
];

const PERMISSIONS: PermissionResponse[] = [
  {
    ...AUDIT,
    id: "permission-1",
    name: "node:get",
    description: "Read nodes",
    preferences: {},
  },
  {
    ...AUDIT,
    id: "permission-2",
    name: "node:set",
    description: "Edit nodes",
    preferences: {},
  },
];

const SCHEMAS: PayloadSchemaResponse[] = [
  {
    ...AUDIT,
    id: "schema-1",
    name: "empty",
    version: 1,
    definition: {},
    valid_from: "2026-01-01T00:00:00Z",
    preferences: {},
  },
  {
    ...AUDIT,
    id: "schema-2",
    name: "restart",
    version: 1,
    definition: {},
    valid_from: "2026-01-01T00:00:00Z",
    preferences: {},
  },
];

function mockTwoPages<T>(items: readonly [T, T]) {
  vi.mocked(apiFetch)
    .mockResolvedValueOnce({
      data: [items[0]],
      page: { page: 1, limit: 1, total_items: 2 },
    })
    .mockResolvedValueOnce({
      data: [items[1]],
      page: { page: 2, limit: 1, total_items: 2 },
    });
}

describe("complete resource collections", () => {
  beforeEach(() => {
    vi.resetAllMocks();
  });

  it("collects actions beyond page one while preserving server filters", async () => {
    mockTwoPages(ACTIONS as [ActionResponse, ActionResponse]);

    await expect(
      listAllActions({ search: "restart", node_class_id: "class-1" }),
    ).resolves.toEqual(ACTIONS);
    expect(apiFetch).toHaveBeenNthCalledWith(
      1,
      "/actions?search=restart&node_class_id=class-1&page=1",
    );
    expect(apiFetch).toHaveBeenNthCalledWith(
      2,
      "/actions?search=restart&node_class_id=class-1&page=2",
    );
  });

  it("collects node classes beyond page one without a fixed option cap", async () => {
    mockTwoPages(NODE_CLASSES as [NodeClassResponse, NodeClassResponse]);

    await expect(listAllNodeClasses({ search: "sensor" })).resolves.toEqual(
      NODE_CLASSES,
    );
    expect(apiFetch).toHaveBeenNthCalledWith(
      1,
      "/node-classes?search=sensor&page=1",
    );
    expect(apiFetch).toHaveBeenNthCalledWith(
      2,
      "/node-classes?search=sensor&page=2",
    );
  });

  it("collects roles and permissions beyond page one", async () => {
    mockTwoPages(ROLES as [RoleResponse, RoleResponse]);
    await expect(listAllRoles()).resolves.toEqual(ROLES);

    mockTwoPages(PERMISSIONS as [PermissionResponse, PermissionResponse]);
    await expect(listAllPermissions()).resolves.toEqual(PERMISSIONS);

    expect(apiFetch).toHaveBeenNthCalledWith(1, "/admin/roles?page=1");
    expect(apiFetch).toHaveBeenNthCalledWith(2, "/admin/roles?page=2");
    expect(apiFetch).toHaveBeenNthCalledWith(3, "/admin/permissions?page=1");
    expect(apiFetch).toHaveBeenNthCalledWith(4, "/admin/permissions?page=2");
  });

  it("collects action-compatible nodes beyond page one and preserves filters", async () => {
    mockTwoPages(NODES as [NodeResponse, NodeResponse]);

    await expect(
      listAllNodes({ search: "sensor", node_class_id: "class-1" }),
    ).resolves.toEqual(NODES);
    expect(apiFetch).toHaveBeenNthCalledWith(
      1,
      "/nodes?search=sensor&node_class_id=class-1&page=1",
    );
    expect(apiFetch).toHaveBeenNthCalledWith(
      2,
      "/nodes?search=sensor&node_class_id=class-1&page=2",
    );
  });

  it("collects compatible class assignments beyond page one with action filtering", async () => {
    const details: [
      NodeClassActionDetailResponse,
      NodeClassActionDetailResponse,
    ] = NODE_CLASSES.map((nodeClass, index) => ({
      node_class_action: {
        id: `assignment-${index + 1}`,
        node_class_id: nodeClass.id,
        action_id: "action-1",
        created_at: AUDIT.created_at,
      },
      node_class: nodeClass,
      action: ACTIONS[0],
    })) as [NodeClassActionDetailResponse, NodeClassActionDetailResponse];
    mockTwoPages(details);

    await expect(
      listAllNodeClassActions({ action_id: "action-1" }),
    ).resolves.toEqual(details);
    expect(apiFetch).toHaveBeenNthCalledWith(
      1,
      "/node-class-actions?action_id=action-1&page=1",
    );
    expect(apiFetch).toHaveBeenNthCalledWith(
      2,
      "/node-class-actions?action_id=action-1&page=2",
    );
  });

  it("collects payload schemas beyond page one while preserving validity filters", async () => {
    mockTwoPages(SCHEMAS as [PayloadSchemaResponse, PayloadSchemaResponse]);

    await expect(
      listAllPayloadSchemas({ valid_at: "2026-08-10T00:00:00Z" }),
    ).resolves.toEqual(SCHEMAS);
    expect(apiFetch).toHaveBeenNthCalledWith(
      1,
      "/admin/payload-schemas?valid_at=2026-08-10T00%3A00%3A00Z&page=1",
    );
    expect(apiFetch).toHaveBeenNthCalledWith(
      2,
      "/admin/payload-schemas?valid_at=2026-08-10T00%3A00%3A00Z&page=2",
    );
  });
});
