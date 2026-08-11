"use client";

import { useActionState, useEffect, useState } from "react";

import ActionMessage from "@/components/forms/ActionMessage";
import Button from "@/components/ui/button";
import Dialog from "@/components/ui/dialog";
import { useActionDialog } from "@/hooks/use-action-dialog";
import { useRefreshAfterAction } from "@/hooks/use-refresh-after-action";
import type { InfraredStateResponse } from "@/lib/api/infrared";

import { deleteStateAction } from "../_lib/actions";
import { EMPTY_INFRARED_STATE_STATE } from "../_lib/state";

export default function DeleteStateDialog({
  state: infraredState,
}: {
  state: InfraredStateResponse;
}) {
  const [actionState, setActionState] = useState(EMPTY_INFRARED_STATE_STATE);
  const dialog = useActionDialog({ state: actionState });

  return (
    <>
      <button
        type="button"
        aria-label={`Delete ${infraredState.name}`}
        onClick={() => dialog.setOpen(true)}
        className="border-critical text-critical hover:bg-critical/10 focus-visible:ring-critical focus-visible:ring-offset-background inline-flex min-h-9 items-center justify-center rounded-xl border px-3 text-sm font-semibold transition-colors focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none"
      >
        Delete
      </button>
      <DeleteStateContent
        key={dialog.formKey}
        infraredState={infraredState}
        open={dialog.open}
        onClose={dialog.reset}
        onStateChange={setActionState}
      />
    </>
  );
}

function DeleteStateContent({
  infraredState,
  open,
  onClose,
  onStateChange,
}: {
  infraredState: InfraredStateResponse;
  open: boolean;
  onClose: () => void;
  onStateChange: (state: typeof EMPTY_INFRARED_STATE_STATE) => void;
}) {
  const [state, action, pending] = useActionState(
    deleteStateAction,
    EMPTY_INFRARED_STATE_STATE,
  );
  useRefreshAfterAction(state);
  useEffect(() => onStateChange(state), [onStateChange, state]);

  return (
    <Dialog
      open={open}
      onClose={onClose}
      title={`Delete ${infraredState.name}`}
      dismissible={!pending}
    >
      <form
        action={action}
        onReset={(event) => event.preventDefault()}
        className="space-y-4"
      >
        <input type="hidden" name="id" value={infraredState.id} />
        <p className="text-sm">
          This removes the <strong>{infraredState.name}</strong> state. Devices
          with definitions for this state may prevent this deletion.
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
