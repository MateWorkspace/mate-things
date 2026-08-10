import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { getNodeByDeviceId } from "@/lib/api/nodes";
import { listTelemetryRecords } from "@/lib/api/telemetry";
import { requirePermission } from "@/lib/session";
import { USER } from "@/test/fixtures";

import TelemetryPage from "./page";

vi.mock("server-only", () => ({}));
vi.mock("./_lib/actions", () => ({ deleteTelemetryAction: vi.fn() }));
vi.mock("@/lib/api/nodes", () => ({ getNodeByDeviceId: vi.fn() }));
vi.mock("@/lib/api/telemetry", () => ({ listTelemetryRecords: vi.fn() }));
vi.mock("@/lib/session", () => ({ requirePermission: vi.fn() }));
vi.mock("next/navigation", () => ({
  usePathname: () => "/telemetry",
  useRouter: () => ({ replace: vi.fn() }),
  useSearchParams: () => new URLSearchParams(),
}));

describe("TelemetryPage", () => {
  beforeEach(() => {
    vi.resetAllMocks();
    vi.mocked(requirePermission).mockResolvedValue({
      user: USER,
      permissions: new Set(["telemetry_record:get"]),
    });
    vi.mocked(listTelemetryRecords).mockResolvedValue({
      data: [],
      total_items: 0,
    });
  });

  afterEach(cleanup);

  it("does not fetch a node label without node:get", async () => {
    render(
      await TelemetryPage({
        searchParams: Promise.resolve({ node_device_id: "device-1" }),
      }),
    );

    expect(getNodeByDeviceId).not.toHaveBeenCalled();
    expect(screen.getByDisplayValue("device-1")).toHaveAttribute("readonly");
    expect(
      screen.queryByRole("button", { name: /node/i }),
    ).not.toBeInTheDocument();
  });
});
