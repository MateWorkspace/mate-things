"use client";

import { useActionState, useCallback, useEffect, useState } from "react";
import { useRouter } from "next/navigation";

import ActionMessage from "@/components/forms/ActionMessage";
import Button from "@/components/ui/button";
import Dialog from "@/components/ui/dialog";
import { useActionDialog } from "@/hooks/use-action-dialog";

import { deleteRecordSessionAction } from "../_lib/actions";
import { EMPTY_RECORD_SESSION_ACTION_STATE } from "../_lib/state";

export default function DeleteRecordSessionDialog({
  sessionId,
}: {
  sessionId: string;
}) {
  const router = useRouter();
  const [state, setState] = useState(EMPTY_RECORD_SESSION_ACTION_STATE);
  // Navigate away instead of refreshing: this route no longer exists.
  const onSuccess = useCallback(() => {
    router.push("/apps/infrared/record");
    router.refresh();
  }, [router]);
  const dialog = useActionDialog({ state, onSuccess });

  return (
    <>
      <button
        type="button"
        aria-label="Delete record session"
        onClick={() => dialog.setOpen(true)}
        className="border-critical text-critical hover:bg-critical/10 focus-visible:ring-critical focus-visible:ring-offset-background inline-flex min-h-9 items-center justify-center rounded-xl border px-3 text-sm font-semibold transition-colors focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none"
      >
        Delete
      </button>
      <DeleteRecordSessionContent
        key={dialog.formKey}
        sessionId={sessionId}
        open={dialog.open}
        onClose={dialog.reset}
        onStateChange={setState}
      />
    </>
  );
}

function DeleteRecordSessionContent({
  sessionId,
  open,
  onClose,
  onStateChange,
}: {
  sessionId: string;
  open: boolean;
  onClose: () => void;
  onStateChange: (state: typeof EMPTY_RECORD_SESSION_ACTION_STATE) => void;
}) {
  const [state, action, pending] = useActionState(
    deleteRecordSessionAction,
    EMPTY_RECORD_SESSION_ACTION_STATE,
  );
  useEffect(() => onStateChange(state), [onStateChange, state]);

  return (
    <Dialog
      open={open}
      onClose={onClose}
      title="Delete record session"
      dismissible={!pending}
    >
      <form
        action={action}
        onReset={(event) => event.preventDefault()}
        className="space-y-4"
      >
        <input type="hidden" name="id" value={sessionId} />
        <p className="text-sm">
          This removes the session with every case, capture, coder and test case
          recorded under it.
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
