import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import type { FirmwareResponse } from "@/lib/api/firmwares";

import FirmwareCard from "./FirmwareCard";

vi.mock("../_lib/actions", () => ({
  createFirmwareAction: vi.fn(),
  deleteFirmwareAction: vi.fn(),
  replaceFirmwareBinaryAction: vi.fn(),
  updateFirmwareAction: vi.fn(),
}));

const FIRMWARE: FirmwareResponse = {
  id: "firmware-1",
  node_class_id: "class-1",
  name: "freezer-v2",
  size: 482_112,
  checksum: "3b1e7c9a2f5d8b4e6c1a9f3d7e2b5c8a1d4e6f9b2c5a8d1e4f7b0c3a6d9e2f5",
  binary_path: "firmware/freezer-v2.bin",
  preferences: {},
  created_at: "2026-07-30T00:00:00Z",
};

describe("FirmwareCard", () => {
  afterEach(cleanup);

  it("shows decision-relevant binary and class context", () => {
    render(
      <FirmwareCard
        firmware={FIRMWARE}
        nodeClassName="Cold Storage"
        permissions={["firmware:get"]}
      />,
    );

    expect(screen.getByText("freezer-v2")).toBeVisible();
    expect(screen.getByText("Cold Storage")).toBeVisible();
    expect(screen.getByText("470.8 KB")).toBeVisible();
    expect(screen.getByText("Binary available")).toBeVisible();
    expect(screen.getByRole("link", { name: "View details" })).toHaveAttribute(
      "href",
      "/firmware/firmware-1",
    );
  });

  it("never exposes OTA from a global firmware card", () => {
    render(
      <FirmwareCard
        firmware={FIRMWARE}
        permissions={["firmware:get", "ota:dispatch"]}
      />,
    );

    expect(
      screen.queryByRole("button", { name: /ota/i }),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("link", { name: /ota/i }),
    ).not.toBeInTheDocument();
  });

  it("shows metadata editing only with firmware:set", () => {
    const { rerender } = render(
      <FirmwareCard
        firmware={FIRMWARE}
        nodeClasses={[{ id: "class-1", name: "Cold Storage" }]}
        permissions={["firmware:get"]}
      />,
    );

    expect(
      screen.queryByRole("button", { name: "Edit freezer-v2" }),
    ).not.toBeInTheDocument();

    rerender(
      <FirmwareCard
        firmware={FIRMWARE}
        nodeClasses={[{ id: "class-1", name: "Cold Storage" }]}
        permissions={["firmware:get", "firmware:set"]}
      />,
    );

    expect(
      screen.getByRole("button", { name: "Edit freezer-v2" }),
    ).toBeVisible();
  });
});
