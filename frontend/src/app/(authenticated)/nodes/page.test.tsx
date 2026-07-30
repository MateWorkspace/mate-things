import { cleanup, render, screen } from "@testing-library/react";
import { redirect } from "next/navigation";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import {
  listAllFirmwares,
  listAllNodeClasses,
  listFirmwares,
  listNodeClasses,
  listNodes,
} from "@/lib/api";
import { requirePermission } from "@/lib/session";
import { nodeFixture, USER } from "@/test/fixtures";

import NodesPage from "./page";

vi.mock("@/lib/api", () => ({
  listAllFirmwares: vi.fn(),
  listAllNodeClasses: vi.fn(),
  listFirmwares: vi.fn(),
  listNodeClasses: vi.fn(),
  listNodes: vi.fn(),
}));

vi.mock("@/lib/session", () => ({
  requirePermission: vi.fn(),
}));

vi.mock("next/navigation", () => ({
  redirect: vi.fn(),
}));

function permit(...permissions: string[]) {
  vi.mocked(requirePermission).mockResolvedValue({
    user: USER,
    permissions: new Set(["node:get", ...permissions]),
  });
}

const CLASSES = Array.from({ length: 49 }, (_, index) => ({
  id: `class-${index + 1}`,
  name: `Node class ${index + 1}`,
  description: "",
  preferences: {},
  created_at: "2026-07-30T00:00:00Z",
}));

const FIRMWARES = Array.from({ length: 49 }, (_, index) => ({
  id: `firmware-${index + 1}`,
  node_class_id: "class-49",
  name: `Firmware ${index + 1}`,
  size: 1024,
  checksum: `checksum-${index + 1}`,
  binary_path: `firmwares/${index + 1}.bin`,
  preferences: {},
  created_at: "2026-07-30T00:00:00Z",
}));

describe("NodesPage", () => {
  beforeEach(() => {
    vi.resetAllMocks();
    vi.mocked(listNodes).mockResolvedValue({
      data: [
        nodeFixture({
          node_class_id: "class-49",
          firmware_id: "firmware-49",
        }),
      ],
      page: { page: 1, limit: 12, total_items: 1 },
    });
    vi.mocked(listAllNodeClasses).mockResolvedValue(CLASSES);
    vi.mocked(listAllFirmwares).mockResolvedValue(FIRMWARES);
    vi.mocked(listNodeClasses).mockResolvedValue({
      data: CLASSES.slice(0, 48),
      page: { page: 1, limit: 48, total_items: 49 },
    });
    vi.mocked(listFirmwares).mockResolvedValue({
      data: FIRMWARES.slice(0, 48),
      page: { page: 1, limit: 48, total_items: 49 },
    });
  });

  afterEach(cleanup);

  it("keeps class and firmware options after the first 48 selectable", async () => {
    permit("node_class:get", "firmware:get");

    render(
      await NodesPage({
        searchParams: Promise.resolve({
          node_class_id: "class-49",
          firmware_id: "firmware-49",
        }),
      }),
    );

    expect(listAllNodeClasses).toHaveBeenCalledOnce();
    expect(listAllFirmwares).toHaveBeenCalledOnce();
    expect(screen.getByRole("combobox", { name: "Node class" })).toHaveValue(
      "class-49",
    );
    expect(screen.getByRole("combobox", { name: "Firmware" })).toHaveValue(
      "firmware-49",
    );
    expect(screen.getAllByText("Node class 49").length).toBeGreaterThan(0);
    expect(screen.getAllByText("Firmware 49").length).toBeGreaterThan(0);
  });

  it("does not fetch or expose option sets without their read permissions", async () => {
    permit();

    render(await NodesPage({ searchParams: Promise.resolve({}) }));

    expect(listAllNodeClasses).not.toHaveBeenCalled();
    expect(listAllFirmwares).not.toHaveBeenCalled();
    expect(
      screen.queryByRole("combobox", { name: "Node class" }),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("combobox", { name: "Firmware" }),
    ).not.toBeInTheDocument();
  });

  it("redirects an out-of-range page to the filtered collection's last page", async () => {
    permit();
    vi.mocked(listNodes).mockResolvedValue({
      data: [],
      page: { page: 9, limit: 12, total_items: 25 },
    });

    render(
      await NodesPage({
        searchParams: Promise.resolve({
          page: "9",
          limit: "12",
          search: "freezer",
          connection: "disconnected",
        }),
      }),
    );

    expect(redirect).toHaveBeenCalledWith(
      expect.stringMatching(
        /^\/nodes\?(?=.*page=3)(?=.*search=freezer)(?=.*connection=disconnected)/,
      ),
    );
  });
});
