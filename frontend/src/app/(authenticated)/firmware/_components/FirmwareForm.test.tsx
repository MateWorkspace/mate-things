import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
  within,
} from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import type { FirmwareResponse } from "@/lib/api/firmwares";

import FirmwareForm from "./FirmwareForm";

const mocks = vi.hoisted(() => ({
  createFirmwareAction: vi.fn(),
  deleteFirmwareAction: vi.fn(),
  refresh: vi.fn(),
  replaceFirmwareBinaryAction: vi.fn(),
  updateFirmwareAction: vi.fn(),
}));

vi.mock("next/navigation", () => ({
  useRouter: () => ({ refresh: mocks.refresh }),
}));

vi.mock("../_lib/actions", () => ({
  createFirmwareAction: mocks.createFirmwareAction,
  deleteFirmwareAction: mocks.deleteFirmwareAction,
  replaceFirmwareBinaryAction: mocks.replaceFirmwareBinaryAction,
  updateFirmwareAction: mocks.updateFirmwareAction,
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

function renderReplacement(): void {
  render(
    <FirmwareForm
      firmware={FIRMWARE}
      configSchema={[{ key: "sample_rate", value_type: "uint32" }]}
      nodeClasses={[{ id: "class-1", name: "Cold Storage" }]}
      canReplace
    />,
  );
}

describe("FirmwareForm replacement schema intent", () => {
  beforeEach(() => {
    vi.resetAllMocks();
  });

  afterEach(cleanup);

  it("offers explicit Keep, Replace, and Clear choices with conditional guidance", async () => {
    const user = userEvent.setup();
    renderReplacement();

    await user.click(
      screen.getByRole("button", { name: "Replace freezer-v2 binary" }),
    );

    const dialog = screen.getByRole("dialog");
    expect(within(dialog).getByRole("radio", { name: "Keep" })).toBeChecked();
    expect(
      within(dialog).getByText(/existing configuration schema.*unchanged/i),
    ).toBeVisible();
    expect(
      within(dialog).queryByLabelText("Configuration key 1"),
    ).not.toBeInTheDocument();

    await user.click(within(dialog).getByRole("radio", { name: "Replace" }));

    expect(within(dialog).getByLabelText("Configuration key 1")).toHaveValue(
      "sample_rate",
    );
    expect(within(dialog).getByLabelText("Value type 1")).toHaveValue("uint32");

    await user.click(within(dialog).getByRole("radio", { name: "Clear" }));

    expect(
      within(dialog).getByText(/remove the existing configuration schema/i),
    ).toBeVisible();
    expect(
      within(dialog).queryByLabelText("Configuration key 1"),
    ).not.toBeInTheDocument();
  });

  it("preserves Replace selection and editor values after validation errors", async () => {
    mocks.replaceFirmwareBinaryAction.mockResolvedValue({
      status: "error",
      title: "Check the replacement",
      message: "Correct the highlighted fields and try again.",
      fieldErrors: {
        config_schema: "Configuration keys must be unique.",
      },
    });
    const user = userEvent.setup();
    renderReplacement();

    await user.click(
      screen.getByRole("button", { name: "Replace freezer-v2 binary" }),
    );
    const dialog = screen.getByRole("dialog");
    await user.click(within(dialog).getByRole("radio", { name: "Replace" }));
    const keyInput = within(dialog).getByLabelText("Configuration key 1");
    await user.clear(keyInput);
    await user.type(keyInput, "draft_rate");
    const binaryInput = within(dialog).getByLabelText(
      "Replacement firmware binary",
    );
    await user.upload(
      binaryInput,
      new File(["firmware"], "replacement.bin", {
        type: "application/octet-stream",
      }),
    );
    fireEvent.submit(dialog.querySelector("form") as HTMLFormElement);

    await waitFor(() => {
      expect(mocks.replaceFirmwareBinaryAction).toHaveBeenCalledOnce();
    });

    expect(
      await within(dialog).findByText("Configuration keys must be unique."),
    ).toBeVisible();
    expect(
      within(dialog).getByRole("radio", { name: "Replace" }),
    ).toBeChecked();
    expect(within(dialog).getByLabelText("Configuration key 1")).toHaveValue(
      "draft_rate",
    );
    expect(
      within(dialog).getByRole("group", { name: "Configuration schema" }),
    ).toHaveFocus();

    const submittedData = mocks.replaceFirmwareBinaryAction.mock.calls[0]?.[1];
    expect(submittedData).toBeInstanceOf(FormData);
    expect((submittedData as FormData).get("schema_intent")).toBe("replace");
    expect((submittedData as FormData).get("config_schema")).toBe(
      '[{"key":"draft_rate","value_type":"uint32"}]',
    );
  });
});
