import { beforeEach, describe, expect, it, vi } from "vitest";

import { listActions } from "@/lib/api/actions";
import { listFirmwares } from "@/lib/api/firmwares";
import { listNodeClasses } from "@/lib/api/node-classes";
import { listNodes } from "@/lib/api/nodes";
import { listRoles } from "@/lib/api/roles";
import { listUsers } from "@/lib/api/users";
import { requirePermission } from "@/lib/session";

import {
  searchActionsAction,
  searchFirmwaresAction,
  searchNodeClassesAction,
  searchNodeDeviceIdsAction,
  searchNodesAction,
  searchRolesAction,
  searchUsersAction,
} from "./entity-search-actions";

vi.mock("server-only", () => ({}));

vi.mock("@/lib/api/actions", () => ({ listActions: vi.fn() }));
vi.mock("@/lib/api/firmwares", () => ({ listFirmwares: vi.fn() }));
vi.mock("@/lib/api/node-classes", () => ({ listNodeClasses: vi.fn() }));
vi.mock("@/lib/api/nodes", () => ({ listNodes: vi.fn() }));
vi.mock("@/lib/api/roles", () => ({ listRoles: vi.fn() }));
vi.mock("@/lib/api/users", () => ({ listUsers: vi.fn() }));
vi.mock("@/lib/session", () => ({ requirePermission: vi.fn() }));

const REQUEST = { query: "  lab  ", page: 2, limit: 6 };

const PAGE = {
  page: { page: 2, limit: 6, total_items: 13 },
};

const NODE = {
  id: "node-1",
  node_class_id: "class-1",
  device_id: "esp32-a1",
  device_info: "ESP32",
  name: "Lab sensor",
  firmware_id: "firmware-1",
  description: "North lab",
  is_connected: true,
  preferences: {},
  created_at: "2026-08-10T00:00:00Z",
};

const ACTION = {
  id: "action-1",
  name: "Restart",
  description: "Restart the node",
  payload_schema_name: "empty",
  payload_schema_version: 1,
  preferences: {},
  created_at: "2026-08-10T00:00:00Z",
};

const NODE_CLASS = {
  id: "class-1",
  name: "Sensors",
  description: "Environmental sensors",
  preferences: {},
  created_at: "2026-08-10T00:00:00Z",
};

const FIRMWARE = {
  id: "firmware-1",
  node_class_id: "class-1",
  name: "sensor-fw-1.0.0",
  size: 1024,
  checksum: "abc123",
  binary_path: "firmware/sensor-fw-1.0.0.bin",
  preferences: {},
  created_at: "2026-08-10T00:00:00Z",
};

const ROLE = {
  id: "role-1",
  name: "Operator",
  description: "Fleet operator",
  is_default: false,
  preferences: {},
  created_at: "2026-08-10T00:00:00Z",
};

const USER = {
  id: "user-1",
  role_id: "role-1",
  name: "Ada Lovelace",
  bio: "Operator",
  username: "ada",
  preferences: {},
  created_at: "2026-08-10T00:00:00Z",
};

describe("entity selector actions", () => {
  beforeEach(() => {
    vi.resetAllMocks();
    vi.mocked(requirePermission).mockResolvedValue({
      user: USER,
      permissions: new Set(),
    });
  });

  it.each([
    ["nodes", searchNodesAction, "node:get", listNodes],
    ["node device ids", searchNodeDeviceIdsAction, "node:get", listNodes],
    ["actions", searchActionsAction, "action:get", listActions],
    [
      "node classes",
      searchNodeClassesAction,
      "node_class:get",
      listNodeClasses,
    ],
    ["firmwares", searchFirmwaresAction, "firmware:get", listFirmwares],
    ["roles", searchRolesAction, "role:get", listRoles],
    ["users", searchUsersAction, "user:get", listUsers],
  ] as const)(
    "checks exact permission before %s transport",
    async (_name, action, permission, transport) => {
      vi.mocked(requirePermission).mockRejectedValue(new Error("forbidden"));

      await expect(action(REQUEST)).rejects.toThrow("forbidden");

      expect(requirePermission).toHaveBeenCalledWith(permission);
      expect(transport).not.toHaveBeenCalled();
    },
  );

  it.each([
    [searchNodesAction, listNodes],
    [searchNodeDeviceIdsAction, listNodes],
    [searchActionsAction, listActions],
    [searchNodeClassesAction, listNodeClasses],
    [searchFirmwaresAction, listFirmwares],
    [searchRolesAction, listRoles],
    [searchUsersAction, listUsers],
  ] as const)(
    "rejects malformed input before permission or transport",
    async (action, transport) => {
      await expect(action({ query: "lab", page: 0 })).rejects.toThrow(
        "Invalid search",
      );

      expect(requirePermission).not.toHaveBeenCalled();
      expect(transport).not.toHaveBeenCalled();
    },
  );

  it("maps nodes to presentation-safe options with normalized transport input", async () => {
    vi.mocked(listNodes).mockResolvedValue({ data: [NODE], ...PAGE });

    await expect(searchNodesAction(REQUEST)).resolves.toEqual({
      items: [
        { value: "node-1", label: "Lab sensor", description: "esp32-a1" },
      ],
      page: 2,
      totalPages: 3,
    });
    expect(listNodes).toHaveBeenCalledWith({
      search: "lab",
      page: 2,
      limit: 6,
    });
  });

  it("maps node device ids without returning raw node records", async () => {
    vi.mocked(listNodes).mockResolvedValue({ data: [NODE], ...PAGE });

    await expect(searchNodeDeviceIdsAction(REQUEST)).resolves.toEqual({
      items: [
        { value: "esp32-a1", label: "Lab sensor", description: "esp32-a1" },
      ],
      page: 2,
      totalPages: 3,
    });
  });

  it.each([
    [
      searchActionsAction,
      listActions,
      ACTION,
      { value: "action-1", label: "Restart", description: "Restart the node" },
    ],
    [
      searchNodeClassesAction,
      listNodeClasses,
      NODE_CLASS,
      {
        value: "class-1",
        label: "Sensors",
        description: "Environmental sensors",
      },
    ],
    [
      searchFirmwaresAction,
      listFirmwares,
      FIRMWARE,
      { value: "firmware-1", label: "sensor-fw-1.0.0" },
    ],
    [
      searchRolesAction,
      listRoles,
      ROLE,
      { value: "role-1", label: "Operator", description: "Fleet operator" },
    ],
    [
      searchUsersAction,
      listUsers,
      USER,
      { value: "user-1", label: "Ada Lovelace (@ada)" },
    ],
  ] as const)(
    "maps a resource response to a safe option",
    async (action, transport, resource, option) => {
      vi.mocked(transport).mockResolvedValue({
        data: [resource],
        ...PAGE,
      } as never);

      await expect(action(REQUEST)).resolves.toEqual({
        items: [option],
        page: 2,
        totalPages: 3,
      });
    },
  );

  it("keeps the default-role annotation separate from the plain selection label", async () => {
    vi.mocked(listRoles).mockResolvedValue({
      data: [{ ...ROLE, is_default: true }],
      ...PAGE,
    });

    await expect(searchRolesAction(REQUEST)).resolves.toEqual({
      items: [
        {
          value: "role-1",
          label: "Operator",
          description: "Fleet operator",
          annotation: "(default)",
        },
      ],
      page: 2,
      totalPages: 3,
    });
  });

  it("does not swallow transport errors", async () => {
    vi.mocked(listNodes).mockRejectedValue(new Error("backend unavailable"));

    await expect(searchNodesAction(REQUEST)).rejects.toThrow(
      "backend unavailable",
    );
  });
});
