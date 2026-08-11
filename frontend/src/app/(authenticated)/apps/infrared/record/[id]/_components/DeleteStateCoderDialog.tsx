"use client";

import { useActionState, useEffect, useState } from "react";

import ActionMessage from "@/components/forms/ActionMessage";
import Button from "@/components/ui/button";
import Dialog from "@/components/ui/dialog";
import { useActionDialog } from "@/hooks/use-action-dialog";
import { useRefreshAfterAction } from "@/hooks/use-refresh-after-action";

import { deleteStateCoderAction } from "../_lib/actions";
import { EMPTY_RECORD_SESSION_ACTION_STATE } from "../_lib/state";

export default function DeleteStateCoderDialog({
  coderId,
}: {
  coderId: string;
}) {
  const [state, setState] = useState(EMPTY_RECORD_SESSION_ACTION_STATE);
  const dialog = useActionDialog({ state });

  return (
    <>
      <button
        type="button"
        aria-label="Delete coder"
        onClick={() => dialog.setOpen(true)}
        className="border-critical text-critical hover:bg-critical/10 focus-visible:ring-critical focus-visible:ring-offset-background inline-flex min-h-9 items-center justify-center rounded-xl border px-3 text-sm font-semibold transition-colors focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none"
      >
        Delete
      </button>
      <DeleteStateCoderContent
        key={dialog.formKey}
        coderId={coderId}
        open={dialog.open}
        onClose={dialog.reset}
        onStateChange={setState}
      />
    </>
  );
}

function DeleteStateCoderContent({
  coderId,
  open,
  onClose,
  onStateChange,
}: {
  coderId: string;
  open: boolean;
  onClose: () => void;
  onStateChange: (state: typeof EMPTY_RECORD_SESSION_ACTION_STATE) => void;
}) {
  const [state, action, pending] = useActionState(
    deleteStateCoderAction,
    EMPTY_RECORD_SESSION_ACTION_STATE,
  );
  useRefreshAfterAction(state);
  useEffect(() => onStateChange(state), [onStateChange, state]);

  return (
    <Dialog
      open={open}
      onClose={onClose}
      title="Delete coder"
      dismissible={!pending}
    >
      <form
        action={action}
        onReset={(event) => event.preventDefault()}
        className="space-y-4"
      >
        <input type="hidden" name="id" value={coderId} />
        <p className="text-sm">
          This removes the generated encoder and decoder for this session. Test
          cases that rely on it may stop working.
        </p>
        <ActionMessage state={state} />
        <div className="flex justify-end gap-2">
          <Button
            type="button"
            variant="secondary"
            disabled={pending}
            onClick={onClose}
          >
            Cancel
          </Button>
          <Button type="submit" variant="critical" disabled={pending}>
            {pending ? "Deleting…" : "Delete"}
          </Button>
        </div>
      </form>
    </Dialog>
  );
}
