import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { listActions, type ActionResponse } from "@/lib/api/actions";
import {
  listFirmwaresByNodeClassId,
  type FirmwareResponse,
} from "@/lib/api/firmwares";
import {
  getNodeClassById,
  type NodeClassResponse,
} from "@/lib/api/node-classes";
import { listNodes } from "@/lib/api/nodes";
import { requirePermission } from "@/lib/session";
import { nodeFixture, USER } from "@/test/fixtures";

import NodeClassDetailPage from "./page";

vi.mock("next/navigation", () => ({
  notFound: vi.fn(),
}));

vi.mock("../_lib/actions", () => ({
  createNodeClassAction: vi.fn(),
  deleteNodeClassAction: vi.fn(),
  updateNodeClassAction: vi.fn(),
}));

vi.mock("@/lib/api/actions", () => ({
  listActions: vi.fn(),
}));

vi.mock("@/lib/api/firmwares", () => ({
  listFirmwaresByNodeClassId: vi.fn(),
}));

vi.mock("@/lib/api/node-classes", () => ({
  getNodeClassById: vi.fn(),
}));

vi.mock("@/lib/api/nodes", () => ({
  listNodes: vi.fn(),
}));

vi.mock("@/lib/session", () => ({
  requirePermission: vi.fn(),
}));

const NODE_CLASS: NodeClassResponse = {
  id: "class-1",
  name: "Cold Storage",
  description: "Temperature-controlled sensors",
  preferences: {},
  created_at: "2026-07-30T00:00:00Z",
};

const FIRMWARE: FirmwareResponse = {
  id: "firmware-1",
  node_class_id: "class-1",
  name: "freezer-v2.4.1",
  size: 2048,
  checksum: "abc123",
  binary_path: "firmware/freezer-v2.4.1.bin",
  preferences: {},
  created_at: "2026-07-30T00:00:00Z",
};

const ACTION: ActionResponse = {
  id: "action-1",
  node_class_id: "class-1",
  name: "Restart compressor",
  description: "Restarts the compressor controller",
  payload_schema_name: "restart",
  payload_schema_version: 1,
  preferences: {},
  created_at: "2026-07-30T00:00:00Z",
};

function permit(...permissions: string[]) {
  vi.mocked(requirePermission).mockResolvedValue({
    user: USER,
    permissions: new Set(["node_class:get", ...permissions]),
  });
}

describe("NodeClassDetailPage", () => {
  beforeEach(() => {
    vi.resetAllMocks();
    vi.mocked(getNodeClassById).mockResolvedValue(NODE_CLASS);
    vi.mocked(listFirmwaresByNodeClassId).mockResolvedValue({
      data: [FIRMWARE],
      page: { page: 1, limit: 12, total_items: 1 },
    });
    vi.mocked(listNodes).mockResolvedValue({
      data: [nodeFixture()],
      page: { page: 1, limit: 12, total_items: 1 },
    });
    vi.mocked(listActions).mockResolvedValue({
      data: [ACTION],
      page: { page: 1, limit: 12, total_items: 1 },
    });
  });

  afterEach(cleanup);

  it("starts all permitted relationship reads before the class read settles", async () => {
    permit("firmware:get", "node:get", "action:get");
    let resolveNodeClass: (value: NodeClassResponse) => void = () => {};
    vi.mocked(getNodeClassById).mockReturnValue(
      new Promise((resolve) => {
        resolveNodeClass = resolve;
      }),
    );

    const page = NodeClassDetailPage({
      params: Promise.resolve({ id: "class-1" }),
    });

    await vi.waitFor(() => {
      expect(listFirmwaresByNodeClassId).toHaveBeenCalledWith("class-1", {
        limit: 12,
      });
      expect(listNodes).toHaveBeenCalledWith({
        node_class_id: "class-1",
        limit: 12,
      });
      expect(listActions).toHaveBeenCalledWith({
        node_class_id: "class-1",
        limit: 12,
      });
    });

    resolveNodeClass(NODE_CLASS);
    render(await page);

    expect(
      screen.getByRole("link", { name: "View Cold Storage Sensor 07" }),
    ).toHaveAttribute("href", "/nodes/node-1");
    expect(
      screen.getByRole("link", { name: "View freezer-v2.4.1" }),
    ).toHaveAttribute("href", "/firmware/firmware-1");
    expect(
      screen.getByRole("link", { name: "View Restart compressor" }),
    ).toHaveAttribute("href", "/actions/action-1");
  });

  it("does not fetch or render relationships the caller cannot read", async () => {
    permit();

    render(
      await NodeClassDetailPage({
        params: Promise.resolve({ id: "class-1" }),
      }),
    );

    expect(listFirmwaresByNodeClassId).not.toHaveBeenCalled();
    expect(listNodes).not.toHaveBeenCalled();
    expect(listActions).not.toHaveBeenCalled();
    expect(
      screen.queryByRole("heading", { name: "Firmware" }),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("heading", { name: "Nodes" }),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("heading", { name: "Actions" }),
    ).not.toBeInTheDocument();
  });
});
