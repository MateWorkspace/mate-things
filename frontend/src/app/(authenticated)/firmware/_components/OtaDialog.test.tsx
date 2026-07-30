import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";

import type { FirmwareResponse } from "@/lib/api/firmwares";
import { nodeFixture } from "@/test/fixtures";

import { dispatchOtaAction } from "../_lib/actions";
import OtaDialog from "./OtaDialog";

vi.mock("../_lib/actions", () => ({
  dispatchOtaAction: vi.fn(),
}));

const FIRMWARES: FirmwareResponse[] = [
  {
    id: "firmware-2",
    node_class_id: "class-1",
    name: "freezer-v2",
    size: 2048,
    checksum: "bbb",
    binary_path: "firmware/freezer-v2.bin",
    preferences: {},
    created_at: "2026-07-30T00:00:00Z",
  },
];

describe("OtaDialog", () => {
  afterEach(() => {
    cleanup();
    vi.resetAllMocks();
  });

  it("presents node and compatible firmware identity before confirmation", async () => {
    const user = userEvent.setup();
    render(
      <OtaDialog
        availableFirmwares={FIRMWARES}
        node={nodeFixture({ name: "Freezer 7", device_id: "DEVICE-7" })}
      />,
    );

    await user.click(
      screen.getByRole("button", { name: "Dispatch OTA to Freezer 7" }),
    );

    expect(screen.getByRole("dialog")).toBeVisible();
    expect(screen.getAllByText("Freezer 7")).toHaveLength(2);
    expect(screen.getByText("DEVICE-7")).toBeVisible();
    expect(screen.getByRole("option", { name: "freezer-v2" })).toBeVisible();
    expect(
      screen.getByRole("button", { name: "Confirm OTA dispatch" }),
    ).toBeDisabled();
  });

  it("requires explicit confirmation before dispatching", async () => {
    const user = userEvent.setup();
    vi.mocked(dispatchOtaAction).mockResolvedValue({
      status: "success",
      title: "OTA dispatched",
      message: "The update request was sent.",
    });
    render(<OtaDialog availableFirmwares={FIRMWARES} node={nodeFixture()} />);

    await user.click(
      screen.getByRole("button", {
        name: "Dispatch OTA to Cold Storage Sensor 07",
      }),
    );
    await user.selectOptions(
      screen.getByRole("combobox", { name: "Firmware" }),
      "firmware-2",
    );
    await user.click(
      screen.getByRole("checkbox", {
        name: /confirm this firmware update/i,
      }),
    );
    await user.click(
      screen.getByRole("button", { name: "Confirm OTA dispatch" }),
    );

    expect(dispatchOtaAction).toHaveBeenCalledWith({
      nodeId: "node-1",
      firmwareId: "firmware-2",
    });
  });

  it("does not render an OTA trigger when no compatible firmware is available", () => {
    render(<OtaDialog availableFirmwares={[]} node={nodeFixture()} />);

    expect(
      screen.queryByRole("button", { name: /dispatch ota/i }),
    ).not.toBeInTheDocument();
    expect(screen.getByText(/no compatible firmware/i)).toBeVisible();
  });
});
