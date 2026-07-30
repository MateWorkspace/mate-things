import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";

import FirmwareDetailError from "./[id]/error";
import FirmwareDetailLoading from "./[id]/loading";
import FirmwareNotFound from "./[id]/not-found";
import FirmwareError from "./error";
import FirmwareLoading from "./loading";

describe("firmware route boundaries", () => {
  afterEach(cleanup);

  it("renders collection and detail loading skeletons", () => {
    const { rerender } = render(<FirmwareLoading />);

    expect(
      screen.getByRole("status", { name: "Loading firmware filters" }),
    ).toBeVisible();
    expect(
      screen.getAllByRole("status", { name: "Loading firmware record" }),
    ).toHaveLength(3);

    rerender(<FirmwareDetailLoading />);

    expect(
      screen.getByRole("status", { name: "Loading firmware details" }),
    ).toBeVisible();
    expect(
      screen.getByRole("status", { name: "Loading configuration schema" }),
    ).toBeVisible();
  });

  it.each([
    {
      Component: FirmwareError,
      title: "Firmware unavailable",
    },
    {
      Component: FirmwareDetailError,
      title: "Firmware details unavailable",
    },
  ])("offers Retry recovery for $title", async ({ Component, title }) => {
    const user = userEvent.setup();
    const unstableRetry = vi.fn();
    render(<Component unstable_retry={unstableRetry} />);

    expect(screen.getByText(title)).toBeVisible();
    await user.click(screen.getByRole("button", { name: "Retry" }));

    expect(unstableRetry).toHaveBeenCalledOnce();
  });

  it("links a missing firmware detail back to the collection", () => {
    render(<FirmwareNotFound />);

    expect(screen.getByText("Firmware not found")).toBeVisible();
    expect(
      screen.getByRole("link", { name: "Return to firmware" }),
    ).toHaveAttribute("href", "/firmware");
  });
});
