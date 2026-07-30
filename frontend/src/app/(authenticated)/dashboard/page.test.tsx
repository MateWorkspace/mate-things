import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { requireSessionContext } from "@/lib/session";
import { USER } from "@/test/fixtures";

import DashboardPage from "./page";

vi.mock("@/lib/session", () => ({
  requireSessionContext: vi.fn(),
}));

describe("DashboardPage", () => {
  afterEach(cleanup);

  async function renderDashboard(permissions: readonly string[]) {
    vi.mocked(requireSessionContext).mockResolvedValue({
      user: USER,
      permissions: new Set(permissions),
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

  it("shows the urgent structure only with a relevant read permission and makes no health claim", async () => {
    await renderDashboard(["node:get"]);

    expect(
      screen.getByRole("heading", { name: "Urgent attention" }),
    ).toBeVisible();
    expect(
      screen.queryByRole("heading", { name: "Recent warnings" }),
    ).not.toBeInTheDocument();
    expect(screen.getByText("Data not connected")).toBeVisible();
    expect(screen.queryByText(/no urgent attention/i)).not.toBeInTheDocument();
  });

  it("shows recent-warning structure only to node-log readers with neutral copy", async () => {
    await renderDashboard(["node_log:get"]);

    expect(
      screen.getByRole("heading", { name: "Recent warnings" }),
    ).toBeVisible();
    expect(screen.getAllByText("Data not connected").length).toBeGreaterThan(0);
    expect(screen.queryByText(/no recent warnings/i)).not.toBeInTheDocument();
  });
});
