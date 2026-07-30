import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";

import Dialog from "./dialog";

describe("Dialog", () => {
  afterEach(cleanup);

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

  it("restores focus to the trigger after a controlled close", () => {
    const onClose = vi.fn();
    const { rerender } = render(
      <>
        <button type="button">Open profile</button>
        <Dialog open={false} onClose={onClose} title="Edit profile">
          Content
        </Dialog>
      </>,
    );
    const trigger = screen.getByRole("button", { name: "Open profile" });
    trigger.focus();

    rerender(
      <>
        <button type="button">Open profile</button>
        <Dialog open onClose={onClose} title="Edit profile">
          Content
        </Dialog>
      </>,
    );
    rerender(
      <>
        <button type="button">Open profile</button>
        <Dialog open={false} onClose={onClose} title="Edit profile">
          Content
        </Dialog>
      </>,
    );

    expect(trigger).toHaveFocus();
    expect(onClose).not.toHaveBeenCalled();
  });
});
