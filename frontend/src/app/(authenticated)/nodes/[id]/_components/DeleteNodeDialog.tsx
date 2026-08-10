"use client";

import { useActionState, useEffect, useRef, useState } from "react";

import ActionMessage from "@/components/forms/ActionMessage";
import FieldError from "@/components/forms/FieldError";
import Button from "@/components/ui/button";
import Dialog from "@/components/ui/dialog";
import Input from "@/components/ui/input";
import Label from "@/components/ui/label";
import { useActionDialog } from "@/hooks/use-action-dialog";
import { useFirstInvalidField } from "@/hooks/use-first-invalid-field";

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
  // deleteNodeAction redirects server-side on success, so `state` here
  // never actually reaches "success" client-side - useActionDialog's
  // auto-close is inert, but the shared open/reset/remount scaffolding
  // still applies for the cancel/reopen path.
  const [state, setState] = useState(INITIAL_STATE);
  const dialog = useActionDialog({ state });

  return (
    <>
      <button
        type="button"
        onClick={() => dialog.setOpen(true)}
        className="border-critical text-critical hover:bg-critical/10 focus-visible:ring-critical focus-visible:ring-offset-background inline-flex items-center justify-center rounded-xl border px-4 py-2.5 text-sm font-semibold transition-colors focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none"
      >
        Delete node
      </button>
      <DeleteNodeContent
        key={dialog.formKey}
        deviceId={deviceId}
        nodeId={nodeId}
        nodeName={nodeName}
        open={dialog.open}
        onClose={dialog.reset}
        onStateChange={setState}
      />
    </>
  );
}

function DeleteNodeContent({
  deviceId,
  nodeId,
  nodeName,
  open,
  onClose,
  onStateChange,
}: DeleteNodeDialogProps & {
  open: boolean;
  onClose: () => void;
  onStateChange: (state: NodeActionState) => void;
}) {
  const [confirmation, setConfirmation] = useState("");
  const [state, formAction, isPending] = useActionState(
    deleteNodeAction,
    INITIAL_STATE,
  );
  const formRef = useRef<HTMLFormElement>(null);
  useFirstInvalidField(state, formRef);
  useEffect(() => onStateChange(state), [onStateChange, state]);

  const close = () => {
    if (!isPending) {
      onClose();
    }
  };

  return (
    <Dialog
      open={open}
      onClose={close}
      title={`Delete ${nodeName}`}
      variant="sheet"
      dismissible={!isPending}
    >
      <form
        ref={formRef}
        action={formAction}
        onReset={(event) => event.preventDefault()}
        className="space-y-4"
      >
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
          <FieldError id="delete-node-confirmation-error">
            {state.fieldErrors?.confirmation}
          </FieldError>
        </div>

        <ActionMessage state={state} />

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
  );
}
