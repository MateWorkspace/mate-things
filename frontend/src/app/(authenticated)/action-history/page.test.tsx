import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { listActionLogs } from "@/lib/api/action-logs";
import { getActionById } from "@/lib/api/actions";
import { ApiError } from "@/lib/api/client";
import { getNodeById } from "@/lib/api/nodes";
import { requirePermission } from "@/lib/session";
import { USER } from "@/test/fixtures";

import ActionHistoryPage from "./page";

vi.mock("server-only", () => ({}));
vi.mock("./_lib/actions", () => ({ deleteActionHistoryAction: vi.fn() }));
vi.mock("@/lib/api/action-logs", () => ({ listActionLogs: vi.fn() }));
vi.mock("@/lib/api/actions", () => ({ getActionById: vi.fn() }));
vi.mock("@/lib/api/nodes", () => ({ getNodeById: vi.fn() }));
vi.mock("@/lib/session", () => ({ requirePermission: vi.fn() }));
vi.mock("next/navigation", () => ({
  redirect: vi.fn(),
  usePathname: () => "/action-history",
  useRouter: () => ({ replace: vi.fn() }),
  useSearchParams: () => new URLSearchParams(),
}));

describe("ActionHistoryPage", () => {
  beforeEach(() => {
    vi.resetAllMocks();
    vi.mocked(requirePermission).mockResolvedValue({
      user: USER,
      permissions: new Set(["action_log:get"]),
    });
    vi.mocked(listActionLogs).mockResolvedValue({
      data: [],
      page: { page: 1, limit: 12, total_items: 0 },
    });
  });

  afterEach(cleanup);

  it("does not fetch action or node labels without their read permissions", async () => {
    render(
      await ActionHistoryPage({
        searchParams: Promise.resolve({
          action_id: "action-1",
          node_id: "node-1",
        }),
      }),
    );

    expect(getActionById).not.toHaveBeenCalled();
    expect(getNodeById).not.toHaveBeenCalled();
    expect(screen.getByDisplayValue("action-1")).toHaveAttribute("readonly");
    expect(screen.getByDisplayValue("node-1")).toHaveAttribute("readonly");
  });

  it("falls back to the raw action ID when its optional label no longer exists", async () => {
    vi.mocked(requirePermission).mockResolvedValue({
      user: USER,
      permissions: new Set(["action_log:get", "action:get"]),
    });
    vi.mocked(getActionById).mockRejectedValue(
      new ApiError(404, "Not Found", "Action missing"),
    );

    render(
      await ActionHistoryPage({
        searchParams: Promise.resolve({ action_id: "action-1" }),
      }),
    );

    expect(screen.getByText("action-1")).toBeVisible();
  });
});
