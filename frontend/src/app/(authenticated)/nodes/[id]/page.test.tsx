import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { getFirmwareConfigParameters } from "@/lib/api/firmwares";
import { getNodeConfig } from "@/lib/api/node-config";
import { getNodeById } from "@/lib/api/nodes";
import { requirePermission } from "@/lib/session";
import { nodeFixture, USER } from "@/test/fixtures";

import NodeDetailPage from "./page";

vi.mock("next/navigation", () => ({
  notFound: vi.fn(),
  redirect: vi.fn(),
  usePathname: () => "/nodes/node-1",
  useRouter: () => ({ push: vi.fn() }),
  useSearchParams: () => new URLSearchParams(),
}));

vi.mock("@/lib/api/firmwares", () => ({
  getFirmwareConfigParameters: vi.fn(),
}));

vi.mock("@/lib/api/node-config", () => ({
  getNodeConfig: vi.fn(),
}));

vi.mock("@/lib/api/nodes", () => ({
  getNodeById: vi.fn(),
}));

vi.mock("@/lib/session", () => ({
  requirePermission: vi.fn(),
}));

function permit(...permissions: string[]) {
  vi.mocked(requirePermission).mockResolvedValue({
    user: USER,
    permissions: new Set(["node:get", ...permissions]),
  });
}

async function renderPage(
  tab: string | string[] | undefined,
  permissions: string[] = [],
) {
  permit(...permissions);

  render(
    await NodeDetailPage({
      params: Promise.resolve({ id: "node-1" }),
      searchParams: Promise.resolve({ tab }),
    }),
  );
}

describe("NodeDetailPage", () => {
  beforeEach(() => {
    vi.resetAllMocks();
    vi.mocked(getNodeById).mockResolvedValue(nodeFixture());
    vi.mocked(getNodeConfig).mockResolvedValue([]);
    vi.mocked(getFirmwareConfigParameters).mockResolvedValue([]);
  });

  afterEach(cleanup);

  it("normalizes an invalid tab to overview without configuration reads", async () => {
    await renderPage("unsupported");

    expect(
      screen.getByRole("heading", { name: "Node identity" }),
    ).toBeVisible();
    expect(getNodeConfig).not.toHaveBeenCalled();
    expect(getFirmwareConfigParameters).not.toHaveBeenCalled();
  });

  it("loads configuration values and schema only for an authorized active configuration tab", async () => {
    await renderPage("configuration", ["node_config:get", "firmware:get"]);

    expect(getNodeConfig).toHaveBeenCalledWith("node-1");
    expect(getFirmwareConfigParameters).toHaveBeenCalledWith("firmware-1");
    expect(
      screen.getByRole("heading", { name: "Device configuration" }),
    ).toBeVisible();
  });

  it("does not read firmware schema when the configuration reader lacks firmware:get", async () => {
    await renderPage("configuration", ["node_config:get"]);

    expect(getNodeConfig).toHaveBeenCalledWith("node-1");
    expect(getFirmwareConfigParameters).not.toHaveBeenCalled();
    expect(screen.getByText(/firmware:get permission/i)).toBeVisible();
  });

  it("renders a truthful actions placeholder without configuration reads", async () => {
    await renderPage("actions", ["action:get"]);

    expect(screen.getByText(/action data is not connected yet/i)).toBeVisible();
    expect(getNodeConfig).not.toHaveBeenCalled();
    expect(getFirmwareConfigParameters).not.toHaveBeenCalled();
  });
});
