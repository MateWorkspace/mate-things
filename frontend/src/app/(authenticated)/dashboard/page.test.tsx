import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { requireSessionContext } from "@/lib/session";
import { nodeFixture, USER } from "@/test/fixtures";

import { loadDashboardData } from "./_lib/dashboard-data";
import DashboardPage from "./page";

vi.mock("server-only", () => ({}));
vi.mock("next/navigation", () => ({ useRouter: () => ({ refresh: vi.fn() }) }));
vi.mock("@/lib/session", () => ({ requireSessionContext: vi.fn() }));
vi.mock("./_lib/dashboard-data", async (importOriginal) => {
  const actual = await importOriginal<typeof import("./_lib/dashboard-data")>();
  return { ...actual, loadDashboardData: vi.fn() };
});

describe("DashboardPage", () => {
  beforeEach(() => {
    vi.resetAllMocks();
    vi.mocked(requireSessionContext).mockResolvedValue({
      user: USER,
      permissions: new Set(["node:get"]),
    });
    vi.mocked(loadDashboardData).mockResolvedValue({
      nodes: {
        total: 120,
        sampled: 48,
        connected: 47,
        disconnected: [nodeFixture({ name: "Sampled offline node" })],
      },
      windowStartedAt: "2026-08-10T06:00:00.000Z",
      loadedAt: "2026-08-10T12:00:00.000Z",
    });
  });

  afterEach(cleanup);

  it("labels connection data and disconnected alerts as bounded samples without filter links", async () => {
    render(await DashboardPage());

    expect(
      screen.getByText(/Connection figures are a bounded first-page sample/),
    ).toBeVisible();
    expect(screen.getByText(/Sample-only node alerts/)).toBeVisible();
    expect(screen.getByText("Sampled offline node")).toBeVisible();

    for (const link of screen.getAllByRole("link")) {
      expect(link).not.toHaveAttribute(
        "href",
        expect.stringContaining("connection="),
      );
      expect(link).not.toHaveAttribute(
        "href",
        expect.stringContaining("is_connected="),
      );
    }
  });
});
