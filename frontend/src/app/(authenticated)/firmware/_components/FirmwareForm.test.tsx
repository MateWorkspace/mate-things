import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";

import type { FirmwareResponse } from "@/lib/api/firmwares";

import FirmwareForm from "./FirmwareForm";

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
  size: 2048,
  checksum: "checksum",
  binary_path: "firmware/freezer-v2.bin",
  preferences: {},
  created_at: "2026-07-30T00:00:00Z",
};

describe("FirmwareForm", () => {
  afterEach(cleanup);

  it("offers accessible config-schema rows in the upload dialog", async () => {
    const user = userEvent.setup();
    render(
      <FirmwareForm
        nodeClasses={[
          { id: "class-1", name: "Cold Storage" },
          { id: "class-2", name: "Climate Control" },
        ]}
      />,
    );

    await user.click(screen.getByRole("button", { name: "Upload firmware" }));

    expect(screen.getByLabelText("Firmware name")).toBeVisible();
    expect(screen.getByLabelText("Node class")).toBeVisible();
    expect(
      screen.getByRole("option", { name: "Climate Control" }),
    ).toBeInTheDocument();
    expect(screen.getByLabelText("Firmware binary")).toHaveAttribute(
      "type",
      "file",
    );
    expect(screen.getByLabelText("Configuration key 1")).toBeVisible();
    expect(screen.getByLabelText("Value type 1")).toHaveValue("");
    expect(
      screen.getByRole("button", { name: "Add configuration parameter" }),
    ).toBeVisible();
  });

  it("shows only the permitted detail mutation triggers", () => {
    const { rerender } = render(
      <FirmwareForm
        firmware={FIRMWARE}
        configSchema={[]}
        nodeClasses={[{ id: "class-1", name: "Cold Storage" }]}
        canEdit
      />,
    );

    expect(
      screen.getByRole("button", { name: "Edit freezer-v2" }),
    ).toBeVisible();
    expect(
      screen.queryByRole("button", { name: "Replace freezer-v2 binary" }),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Delete freezer-v2" }),
    ).not.toBeInTheDocument();

    rerender(
      <FirmwareForm
        firmware={FIRMWARE}
        configSchema={[]}
        nodeClasses={[{ id: "class-1", name: "Cold Storage" }]}
        canDelete
        canReplace
      />,
    );

    expect(
      screen.queryByRole("button", { name: "Edit freezer-v2" }),
    ).not.toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Replace freezer-v2 binary" }),
    ).toBeVisible();
    expect(
      screen.getByRole("button", { name: "Delete freezer-v2" }),
    ).toBeVisible();
  });

  it("keeps the current node class selected when it is absent from loaded options", async () => {
    const user = userEvent.setup();
    render(
      <FirmwareForm
        firmware={FIRMWARE}
        configSchema={[]}
        nodeClasses={Array.from({ length: 48 }, (_, index) => ({
          id: `other-class-${index + 1}`,
          name: `Other class ${index + 1}`,
        }))}
        canEdit
      />,
    );

    await user.click(screen.getByRole("button", { name: "Edit freezer-v2" }));

    expect(
      screen.getByRole("option", { name: "Current class (class-1)" }),
    ).toBeInTheDocument();
    expect(screen.getByRole("combobox", { name: "Node class" })).toHaveValue(
      "class-1",
    );
  });
});
