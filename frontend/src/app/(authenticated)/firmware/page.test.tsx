import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { listFirmwares } from "@/lib/api/firmwares";
import { listNodeClasses } from "@/lib/api/node-classes";
import { requirePermission } from "@/lib/session";
import { USER } from "@/test/fixtures";

import FirmwarePage from "./page";

vi.mock("@/lib/api/firmwares", () => ({
  listFirmwares: vi.fn(),
}));

vi.mock("@/lib/api/node-classes", () => ({
  listNodeClasses: vi.fn(),
}));

vi.mock("@/lib/session", () => ({
  requirePermission: vi.fn(),
}));

vi.mock("./_lib/actions", () => ({
  createFirmwareAction: vi.fn(),
  deleteFirmwareAction: vi.fn(),
  replaceFirmwareBinaryAction: vi.fn(),
  updateFirmwareAction: vi.fn(),
}));

function permit(...permissions: string[]) {
  vi.mocked(requirePermission).mockResolvedValue({
    user: USER,
    permissions: new Set(["firmware:get", ...permissions]),
  });
}

describe("FirmwarePage", () => {
  beforeEach(() => {
    vi.resetAllMocks();
    permit("node_class:get");
    vi.mocked(listFirmwares).mockResolvedValue({
      data: [
        {
          id: "firmware-1",
          node_class_id: "class-1",
          name: "freezer-v2",
          size: 2048,
          checksum: "checksum",
          binary_path: "firmware/freezer-v2.bin",
          preferences: {},
          created_at: "2026-07-30T00:00:00Z",
        },
      ],
      page: { page: 2, limit: 12, total_items: 13 },
    });
    vi.mocked(listNodeClasses).mockResolvedValue({
      data: [
        {
          id: "class-1",
          name: "Cold Storage",
          description: "",
          preferences: {},
          created_at: "2026-07-30T00:00:00Z",
        },
      ],
      page: { page: 1, limit: 48, total_items: 1 },
    });
  });

  afterEach(cleanup);

  it("loads URL-backed filters and preserves them in pagination", async () => {
    render(
      await FirmwarePage({
        searchParams: Promise.resolve({
          page: "2",
          limit: "12",
          search: " freezer ",
          node_class_id: "class-1",
        }),
      }),
    );

    expect(listFirmwares).toHaveBeenCalledWith({
      page: 2,
      limit: 12,
      search: "freezer",
      node_class_id: "class-1",
    });
    expect(screen.getAllByText("Cold Storage")).toHaveLength(2);
    expect(screen.getByRole("link", { name: "Previous page" })).toHaveAttribute(
      "href",
      expect.stringContaining("node_class_id=class-1"),
    );
  });

  it("shows upload only with firmware:add", async () => {
    permit("node_class:get");
    render(await FirmwarePage({ searchParams: Promise.resolve({}) }));
    expect(
      screen.queryByRole("button", { name: "Upload firmware" }),
    ).not.toBeInTheDocument();

    cleanup();
    permit("node_class:get", "firmware:add");
    render(await FirmwarePage({ searchParams: Promise.resolve({}) }));
    expect(
      screen.getByRole("button", { name: "Upload firmware" }),
    ).toBeVisible();
  });
});
