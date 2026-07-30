import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import {
  getFirmwareBinaryUrlById,
  getFirmwareById,
  getFirmwareConfigParameters,
} from "@/lib/api/firmwares";
import { listAllNodeClasses } from "@/lib/api/node-classes";
import { listNodes } from "@/lib/api/nodes";
import { requirePermission } from "@/lib/session";
import { USER } from "@/test/fixtures";

import FirmwareDetailPage from "./page";

vi.mock("next/navigation", () => ({
  notFound: vi.fn(),
  redirect: vi.fn(),
}));

vi.mock("@/lib/api/firmwares", () => ({
  getFirmwareBinaryUrlById: vi.fn(),
  getFirmwareById: vi.fn(),
  getFirmwareConfigParameters: vi.fn(),
}));

vi.mock("@/lib/api/node-classes", () => ({
  listAllNodeClasses: vi.fn(),
}));

vi.mock("@/lib/api/nodes", () => ({
  listNodes: vi.fn(),
}));

vi.mock("@/lib/session", () => ({
  requirePermission: vi.fn(),
}));

vi.mock("../_lib/actions", () => ({
  createFirmwareAction: vi.fn(),
  deleteFirmwareAction: vi.fn(),
  replaceFirmwareBinaryAction: vi.fn(),
  updateFirmwareAction: vi.fn(),
}));

describe("FirmwareDetailPage", () => {
  beforeEach(() => {
    vi.resetAllMocks();
    vi.mocked(requirePermission).mockResolvedValue({
      user: USER,
      permissions: new Set([
        "firmware:get",
        "firmware:set",
        "node_class:get",
        "node:get",
      ]),
    });
    vi.mocked(getFirmwareById).mockResolvedValue({
      id: "firmware-1",
      node_class_id: "class-1",
      name: "freezer-v2",
      size: 2048,
      checksum: "checksum",
      binary_path: "firmware/freezer-v2.bin",
      preferences: { channel: "stable" },
      created_at: "2026-07-30T00:00:00Z",
    });
    vi.mocked(getFirmwareConfigParameters).mockResolvedValue([
      { key: "sample_rate", value_type: "uint32" },
    ]);
    vi.mocked(getFirmwareBinaryUrlById).mockResolvedValue(
      "https://storage.example/freezer-v2",
    );
    vi.mocked(listAllNodeClasses).mockResolvedValue([
      {
        id: "class-1",
        name: "Cold Storage",
        description: "",
        preferences: {},
        created_at: "2026-07-30T00:00:00Z",
      },
      {
        id: "class-2",
        name: "Climate Control",
        description: "",
        preferences: {},
        created_at: "2026-07-30T00:00:00Z",
      },
    ]);
    vi.mocked(listNodes).mockResolvedValue({
      data: [],
      page: { page: 1, limit: 12, total_items: 0 },
    });
  });

  afterEach(cleanup);

  it("shows binary, schema, class, preferences, audit, and download details", async () => {
    render(
      await FirmwareDetailPage({
        params: Promise.resolve({ id: "firmware-1" }),
      }),
    );

    expect(screen.getByRole("link", { name: "Cold Storage" })).toBeVisible();
    expect(screen.getByText("sample_rate")).toBeVisible();
    expect(screen.getAllByText("uint32")).toHaveLength(2);
    expect(screen.getByText(/\"channel\": \"stable\"/)).toBeVisible();
    expect(screen.getByText("2 KB")).toBeVisible();
    expect(screen.getByText("checksum")).toBeVisible();
    expect(
      screen.getByRole("link", { name: "Download binary" }),
    ).toHaveAttribute("href", "https://storage.example/freezer-v2");
    expect(getFirmwareConfigParameters).toHaveBeenCalledWith("firmware-1");
    expect(listAllNodeClasses).toHaveBeenCalledOnce();
    expect(listNodes).toHaveBeenCalledWith({
      firmware_id: "firmware-1",
      limit: 12,
    });
  });

  it("does not expose OTA from global firmware details", async () => {
    vi.mocked(requirePermission).mockResolvedValue({
      user: USER,
      permissions: new Set(["firmware:get", "ota:dispatch"]),
    });

    render(
      await FirmwareDetailPage({
        params: Promise.resolve({ id: "firmware-1" }),
      }),
    );

    expect(
      screen.queryByRole("button", { name: /ota/i }),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("link", { name: /ota/i }),
    ).not.toBeInTheDocument();
  });

  it("resolves and selects a current node class beyond the first 48 options", async () => {
    const user = userEvent.setup();
    vi.mocked(getFirmwareById).mockResolvedValue({
      id: "firmware-1",
      node_class_id: "class-49",
      name: "freezer-v2",
      size: 2048,
      checksum: "checksum",
      binary_path: "firmware/freezer-v2.bin",
      preferences: {},
      created_at: "2026-07-30T00:00:00Z",
    });
    vi.mocked(listAllNodeClasses).mockResolvedValue(
      Array.from({ length: 49 }, (_, index) => ({
        id: `class-${index + 1}`,
        name: `Node class ${index + 1}`,
        description: "",
        preferences: {},
        created_at: "2026-07-30T00:00:00Z",
      })),
    );

    render(
      await FirmwareDetailPage({
        params: Promise.resolve({ id: "firmware-1" }),
      }),
    );

    expect(screen.getByRole("link", { name: "Node class 49" })).toBeVisible();
    await user.click(screen.getByRole("button", { name: "Edit freezer-v2" }));
    expect(screen.getByRole("combobox", { name: "Node class" })).toHaveValue(
      "class-49",
    );
  });
});
