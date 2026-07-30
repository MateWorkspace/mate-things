import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { requireSessionContext } from "@/lib/session";
import { USER } from "@/test/fixtures";

import { loadDashboardData, type DashboardData } from "./_lib/dashboard-data";
import DashboardPage from "./page";

vi.mock("@/lib/session", () => ({
  requireSessionContext: vi.fn(),
}));

vi.mock("./_lib/dashboard-data", () => ({
  DASHBOARD_RECENT_WINDOW_HOURS: 6,
  loadDashboardData: vi.fn(),
}));

vi.mock("./_components/DashboardRefresh", () => ({
  default: () => <button type="button">Refresh</button>,
}));

describe("DashboardPage", () => {
  afterEach(cleanup);

  async function renderDashboard(
    permissions: readonly string[],
    data: Partial<DashboardData> = {},
  ) {
    vi.mocked(requireSessionContext).mockResolvedValue({
      user: USER,
      permissions: new Set(permissions),
    });
    vi.mocked(loadDashboardData).mockResolvedValue({
      loadedAt: "2026-07-30T15:00:00.000Z",
      windowStartedAt: "2026-07-30T09:00:00.000Z",
      ...data,
    });

    render(await DashboardPage());
  }

  it("omits operational sections when no relevant read permission is present", async () => {
    await renderDashboard([]);

    expect(
      screen.queryByRole("heading", { name: "Urgent attention" }),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("heading", { name: "Recent warnings" }),
    ).not.toBeInTheDocument();
  });

  it("shows honest node metrics and actionable disconnected records", async () => {
    await renderDashboard(["node:get"], {
      nodes: {
        total: 83,
        sampled: 48,
        connected: 47,
        disconnected: [
          {
            id: "node-1",
            node_class_id: "class-1",
            device_id: "device-1",
            device_info: "",
            name: "Boiler sensor",
            firmware_id: "firmware-1",
            description: "Plant room",
            is_connected: false,
            preferences: {},
            created_at: "2026-07-30T09:00:00.000Z",
          },
        ],
      },
    });

    expect(
      screen.getByRole("heading", { name: "Urgent attention" }),
    ).toBeVisible();
    expect(screen.getByText("83")).toBeVisible();
    expect(screen.getAllByText("48-node page sample")).toHaveLength(2);
    expect(screen.getByRole("link", { name: "Inspect node" })).toHaveAttribute(
      "href",
      "/nodes/node-1",
    );
  });

  it("shows recent warning records only to node-log readers", async () => {
    await renderDashboard(["node_log:get"], {
      recentWarnings: [
        {
          id: 17,
          node_device_id: "device-1",
          level: "WARN",
          tag: "battery",
          message: "Battery is low",
          logged_at: "2026-07-30T14:00:00.000Z",
          created_at: "2026-07-30T14:00:00.100Z",
        },
      ],
    });

    expect(
      screen.getByRole("heading", { name: "Recent warnings" }),
    ).toBeVisible();
    expect(screen.getByText("Battery is low")).toBeVisible();
    expect(
      screen.getByRole("link", {
        name: "Inspect warn logs for device-1",
      }),
    ).toBeVisible();
  });
});
