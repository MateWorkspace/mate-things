import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import Dialog from "./dialog";

describe("Dialog", () => {
  it("labels the modal and closes on Escape", async () => {
    const onClose = vi.fn();
    const user = userEvent.setup();

    render(
      <Dialog open onClose={onClose} title="Edit profile">
        Content
      </Dialog>,
    );

    expect(
      screen.getByRole("dialog", { name: "Edit profile" }),
    ).toHaveAttribute("aria-labelledby");

    await user.keyboard("{Escape}");

    expect(onClose).toHaveBeenCalledOnce();
  });
});
