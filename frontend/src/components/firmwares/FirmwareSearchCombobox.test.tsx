import { fireEvent, render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { searchFirmwaresAction } from "@/lib/actions/entity-search-actions";

import FirmwareSearchCombobox from "./FirmwareSearchCombobox";

vi.mock("@/lib/actions/entity-search-actions", () => ({
  searchFirmwaresAction: vi.fn(),
}));

describe("FirmwareSearchCombobox", () => {
  beforeEach(() => {
    vi.resetAllMocks();
    vi.mocked(searchFirmwaresAction).mockResolvedValue({
      items: [{ value: "firmware-1", label: "sensor-fw-1.0.0" }],
      page: 1,
      totalPages: 1,
    });
  });

  it("selects a firmware option and populates the hidden field", async () => {
    const { container } = render(<FirmwareSearchCombobox name="firmware_id" />);

    fireEvent.click(screen.getByRole("button", { name: "Any firmware" }));

    const option = await screen.findByRole("option", {
      name: "sensor-fw-1.0.0",
    });
    fireEvent.click(option);

    expect(
      screen.getByRole("button", { name: "sensor-fw-1.0.0" }),
    ).toBeVisible();
    expect(
      container.querySelector<HTMLInputElement>('input[name="firmware_id"]'),
    ).toHaveValue("firmware-1");
  });
});
