import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { ApiError } from "@/lib/api/client";
import { listNodeLogs } from "@/lib/api/node-logs";
import { getNodeByDeviceId } from "@/lib/api/nodes";
import { requirePermission } from "@/lib/session";
import { USER } from "@/test/fixtures";

import NodeLogsPage from "./page";

vi.mock("server-only", () => ({}));
vi.mock("./_lib/actions", () => ({ deleteNodeLogsAction: vi.fn() }));
vi.mock("@/lib/api/node-logs", () => ({ listNodeLogs: vi.fn() }));
vi.mock("@/lib/api/nodes", () => ({ getNodeByDeviceId: vi.fn() }));
vi.mock("@/lib/session", () => ({ requirePermission: vi.fn() }));
vi.mock("next/navigation", () => ({
  usePathname: () => "/node-logs",
  useRouter: () => ({ replace: vi.fn() }),
  useSearchParams: () => new URLSearchParams(),
}));

describe("NodeLogsPage", () => {
  beforeEach(() => {
    vi.resetAllMocks();
    vi.mocked(requirePermission).mockResolvedValue({
      user: USER,
      permissions: new Set(["node_log:get"]),
    });
    vi.mocked(listNodeLogs).mockResolvedValue({ data: [], total_items: 0 });
  });

  afterEach(cleanup);

  it("does not fetch a node label without node:get", async () => {
    render(
      await NodeLogsPage({
        searchParams: Promise.resolve({ node_device_id: "device-1" }),
      }),
    );

    expect(getNodeByDeviceId).not.toHaveBeenCalled();
    expect(screen.getByDisplayValue("device-1")).toHaveAttribute("readonly");
    expect(
      screen.queryByRole("button", { name: /node/i }),
    ).not.toBeInTheDocument();
  });

  it("falls back to the raw node device ID when its optional label returns 404", async () => {
    vi.mocked(requirePermission).mockResolvedValue({
      user: USER,
      permissions: new Set(["node_log:get", "node:get"]),
    });
    vi.mocked(getNodeByDeviceId).mockRejectedValue(
      new ApiError(404, "Not Found", "Node missing"),
    );

    render(
      await NodeLogsPage({
        searchParams: Promise.resolve({ node_device_id: "device-1" }),
      }),
    );

    expect(screen.getByText("device-1")).toBeVisible();
  });

  it.each([
    ["401", new ApiError(401, "Unauthorized", "Sign in")],
    ["403", new ApiError(403, "Forbidden", "Not allowed")],
    ["500", new ApiError(500, "Server Error", "Try again")],
    ["network failure", new TypeError("fetch failed")],
  ])("rethrows a %s optional-label failure", async (_name, failure) => {
    vi.mocked(requirePermission).mockResolvedValue({
      user: USER,
      permissions: new Set(["node_log:get", "node:get"]),
    });
    vi.mocked(getNodeByDeviceId).mockRejectedValue(failure);

    await expect(
      NodeLogsPage({
        searchParams: Promise.resolve({ node_device_id: "device-1" }),
      }),
    ).rejects.toBe(failure);
  });
});
