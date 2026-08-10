import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, expect, it, vi } from "vitest";

import Dialog from "@/components/ui/dialog";
import type { ActionState } from "@/lib/forms/action-state";

import { useActionDialog } from "./use-action-dialog";

function StatefulForm() {
  return <input aria-label="Draft" defaultValue="fresh" />;
}

function Harness({
  pending = false,
  state = { status: "idle" },
  onSuccess,
}: {
  pending?: boolean;
  state?: ActionState;
  onSuccess?: () => void;
}) {
  const dialog = useActionDialog({ state, onSuccess });

  return (
    <>
      <button type="button" onClick={() => dialog.setOpen(true)}>
        Open
      </button>
      <Dialog
        open={dialog.open}
        onClose={dialog.reset}
        dismissible={!pending}
        title="Action"
      >
        <StatefulForm key={dialog.formKey} />
      </Dialog>
    </>
  );
}

afterEach(cleanup);

it("can reopen with a fresh keyed child after success", async () => {
  const user = userEvent.setup();
  const onSuccess = vi.fn();
  const { rerender } = render(<Harness onSuccess={onSuccess} />);

  await user.click(screen.getByRole("button", { name: "Open" }));
  await user.clear(screen.getByLabelText("Draft"));
  await user.type(screen.getByLabelText("Draft"), "changed");

  rerender(<Harness state={{ status: "success" }} onSuccess={onSuccess} />);
  expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  expect(onSuccess).toHaveBeenCalledOnce();

  await user.click(screen.getByRole("button", { name: "Open" }));
  expect(screen.getByRole("dialog")).toBeVisible();
  expect(screen.getByLabelText("Draft")).toHaveValue("fresh");
});

it("blocks close while pending", async () => {
  const user = userEvent.setup();
  render(<Harness pending />);

  await user.click(screen.getByRole("button", { name: "Open" }));
  await user.keyboard("{Escape}");

  expect(screen.getByRole("dialog")).toBeVisible();
});

it("restores the trigger focus and resets the keyed owner after cancel", async () => {
  const user = userEvent.setup();
  render(<Harness />);
  const trigger = screen.getByRole("button", { name: "Open" });

  await user.click(trigger);
  await user.clear(screen.getByLabelText("Draft"));
  await user.type(screen.getByLabelText("Draft"), "changed");
  await user.click(screen.getByRole("button", { name: "Close Action" }));

  expect(trigger).toHaveFocus();
  await user.click(trigger);
  expect(screen.getByLabelText("Draft")).toHaveValue("fresh");
});
