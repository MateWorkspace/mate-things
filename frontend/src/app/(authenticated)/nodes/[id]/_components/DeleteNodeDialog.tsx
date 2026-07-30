"use client";

import { useActionState, useState } from "react";

import Button from "@/components/ui/button";
import Dialog from "@/components/ui/dialog";
import Input from "@/components/ui/input";
import Label from "@/components/ui/label";

import { deleteNodeAction, type NodeActionState } from "../_lib/actions";

const INITIAL_STATE: NodeActionState = { status: "idle" };

interface DeleteNodeDialogProps {
  nodeId: string;
  nodeName: string;
  deviceId: string;
}

export default function DeleteNodeDialog({
  deviceId,
  nodeId,
  nodeName,
}: DeleteNodeDialogProps) {
  const [open, setOpen] = useState(false);
  const [confirmation, setConfirmation] = useState("");
  const [state, formAction, isPending] = useActionState(
    deleteNodeAction,
    INITIAL_STATE,
  );

  const close = () => {
    if (!isPending) {
      setOpen(false);
      setConfirmation("");
    }
  };

  return (
    <>
      <button
        type="button"
        onClick={() => setOpen(true)}
        className="border-critical text-critical hover:bg-critical/10 focus-visible:ring-critical focus-visible:ring-offset-background inline-flex items-center justify-center rounded-xl border px-4 py-2.5 text-sm font-semibold transition-colors focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none"
      >
        Delete node
      </button>

      <Dialog
        open={open}
        onClose={close}
        title={`Delete ${nodeName}`}
        variant="sheet"
      >
        <form action={formAction} className="space-y-4">
          <input type="hidden" name="node_id" value={nodeId} />

          <div className="border-critical/40 bg-critical/5 rounded-xl border p-4">
            <p className="text-sm">
              This permanently removes the node record. Backend dependency
              constraints may reject the operation.
            </p>
            <p className="mt-2 text-sm">
              Enter the exact device ID{" "}
              <strong className="font-mono">{deviceId}</strong> to continue.
            </p>
          </div>

          <div>
            <Label htmlFor="delete-node-confirmation">Confirm device ID</Label>
            <Input
              id="delete-node-confirmation"
              name="confirmation"
              value={confirmation}
              onChange={(event) => setConfirmation(event.target.value)}
              autoComplete="off"
              spellCheck={false}
              required
              className="font-mono"
              aria-invalid={Boolean(state.fieldErrors?.confirmation)}
              aria-describedby={
                state.fieldErrors?.confirmation
                  ? "delete-node-confirmation-error"
                  : undefined
              }
            />
            {state.fieldErrors?.confirmation ? (
              <p
                id="delete-node-confirmation-error"
                className="text-critical mt-1.5 text-sm"
              >
                {state.fieldErrors.confirmation}
              </p>
            ) : null}
          </div>

          <p aria-live="assertive" className="text-critical text-sm">
            {state.message}
          </p>

          <div className="flex flex-wrap justify-end gap-2">
            <Button
              type="button"
              variant="secondary"
              onClick={close}
              disabled={isPending}
            >
              Cancel
            </Button>
            <button
              type="submit"
              disabled={isPending || confirmation !== deviceId}
              className="bg-critical text-background focus-visible:ring-critical focus-visible:ring-offset-background inline-flex items-center justify-center rounded-xl px-4 py-2.5 text-sm font-semibold transition-opacity hover:opacity-90 focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-50"
            >
              {isPending ? "Deleting…" : "Permanently delete node"}
            </button>
          </div>
        </form>
      </Dialog>
    </>
  );
}
