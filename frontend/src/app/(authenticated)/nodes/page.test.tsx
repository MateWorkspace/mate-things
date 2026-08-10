import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { listAllFirmwares, listAllNodeClasses, listNodes } from "@/lib/api";
import { requirePermission } from "@/lib/session";
import { nodeFixture, USER } from "@/test/fixtures";

import NodesPage from "./page";

vi.mock("server-only", () => ({}));
vi.mock("@/lib/api", () => ({
  listAllFirmwares: vi.fn(),
  listAllNodeClasses: vi.fn(),
  listNodes: vi.fn(),
}));
vi.mock("@/lib/session", () => ({ requirePermission: vi.fn() }));
vi.mock("next/navigation", () => ({
  redirect: vi.fn(),
  usePathname: () => "/nodes",
  useRouter: () => ({ replace: vi.fn() }),
  useSearchParams: () => new URLSearchParams(),
}));

describe("NodesPage", () => {
  beforeEach(() => {
    vi.resetAllMocks();
    vi.mocked(requirePermission).mockResolvedValue({
      user: USER,
      permissions: new Set(["node:get"]),
    });
    vi.mocked(listNodes).mockResolvedValue({
      data: [
        nodeFixture({
          id: "connected",
          name: "Connected node",
          is_connected: true,
        }),
        nodeFixture({ id: "disconnected", name: "Disconnected node" }),
      ],
      page: { page: 1, limit: 12, total_items: 2 },
    });
    vi.mocked(listAllNodeClasses).mockResolvedValue([]);
    vi.mocked(listAllFirmwares).mockResolvedValue([]);
  });

  afterEach(cleanup);

  it("ignores unsupported connection URL filters without filtering the fetched page", async () => {
    render(
      await NodesPage({
        searchParams: Promise.resolve({
          is_connected: "true",
          connection: "connected",
        }),
      }),
    );

    expect(listNodes).toHaveBeenCalledWith(
      expect.not.objectContaining({ is_connected: true }),
    );
    expect(screen.getByText("Connected node")).toBeVisible();
    expect(screen.getByText("Disconnected node")).toBeVisible();
    expect(screen.queryByText("Connection")).not.toBeInTheDocument();
  });
});
