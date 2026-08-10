import { useActionState, useEffect, useState } from "react";
import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, expect, it, vi } from "vitest";

import Dialog from "@/components/ui/dialog";
import {
  INITIAL_ACTION_STATE,
  type ActionState,
} from "@/lib/forms/action-state";

import { useActionDialog } from "./use-action-dialog";

type SubmitAction = (
  previous: ActionState,
  formData: FormData,
) => Promise<ActionState>;

const submitAction = vi.fn<SubmitAction>(async () => ({
  status: "success",
}));

function StatefulForm({
  onStateChange,
}: {
  onStateChange: (state: ActionState) => void;
}) {
  const [state, action] = useActionState(submitAction, INITIAL_ACTION_STATE);
  useEffect(() => onStateChange(state), [onStateChange, state]);

  return (
    <form action={action}>
      <input aria-label="Draft" name="draft" defaultValue="fresh" />
      <button type="submit">Submit</button>
    </form>
  );
}

function Harness({
  pending = false,
  onSuccess,
  closeOnSuccess,
}: {
  pending?: boolean;
  onSuccess?: () => void;
  closeOnSuccess?: boolean;
}) {
  const [state, setState] = useState<ActionState>(INITIAL_ACTION_STATE);
  const dialog = useActionDialog({ state, onSuccess, closeOnSuccess });

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
        {state.status === "success" ? (
          <button type="button" onClick={dialog.reset}>
            Done
          </button>
        ) : (
          <StatefulForm key={dialog.formKey} onStateChange={setState} />
        )}
      </Dialog>
    </>
  );
}

afterEach(cleanup);

it("submits successfully twice through a keyed useActionState owner", async () => {
  const user = userEvent.setup();
  const onSuccess = vi.fn();
  submitAction.mockClear();
  render(<Harness onSuccess={onSuccess} />);

  await user.click(screen.getByRole("button", { name: "Open" }));
  await user.clear(screen.getByLabelText("Draft"));
  await user.type(screen.getByLabelText("Draft"), "changed");
  await user.click(screen.getByRole("button", { name: "Submit" }));
  await waitFor(() =>
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument(),
  );
  expect(onSuccess).toHaveBeenCalledOnce();

  await user.click(screen.getByRole("button", { name: "Open" }));
  expect(screen.getByRole("dialog")).toBeVisible();
  expect(screen.getByLabelText("Draft")).toHaveValue("fresh");
  await user.click(screen.getByRole("button", { name: "Submit" }));

  await waitFor(() => expect(onSuccess).toHaveBeenCalledTimes(2));
  expect(submitAction).toHaveBeenCalledTimes(2);
  expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
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

it("stays open on success and only closes via a manual reset when closeOnSuccess is false", async () => {
  const user = userEvent.setup();
  const onSuccess = vi.fn();
  submitAction.mockClear();
  render(<Harness onSuccess={onSuccess} closeOnSuccess={false} />);

  await user.click(screen.getByRole("button", { name: "Open" }));
  await user.click(screen.getByRole("button", { name: "Submit" }));

  await waitFor(() => expect(onSuccess).toHaveBeenCalledOnce());
  expect(screen.getByRole("dialog")).toBeVisible();
  const doneButton = await screen.findByRole("button", { name: "Done" });

  await user.click(doneButton);
  expect(screen.queryByRole("dialog")).not.toBeInTheDocument();

  await user.click(screen.getByRole("button", { name: "Open" }));
  expect(screen.getByLabelText("Draft")).toHaveValue("fresh");
});
