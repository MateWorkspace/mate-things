"use client";

import { useActionState, useEffect, useId, useRef, useState } from "react";

import ActionMessage from "@/components/forms/ActionMessage";
import FieldError from "@/components/forms/FieldError";
import Button from "@/components/ui/button";
import Dialog from "@/components/ui/dialog";
import Input from "@/components/ui/input";
import Label from "@/components/ui/label";
import { useActionDialog } from "@/hooks/use-action-dialog";
import type { ApiKeyResponse } from "@/lib/api/api-keys";
import { useFirstInvalidField } from "@/hooks/use-first-invalid-field";
import { useRefreshAfterAction } from "@/hooks/use-refresh-after-action";

import { deleteApiKeyAction } from "../_lib/actions";
import { EMPTY_API_KEY_STATE } from "../_lib/state";

export default function DeleteApiKeyDialog({
  apiKey,
}: {
  apiKey: ApiKeyResponse;
}) {
  const [state, setState] = useState(EMPTY_API_KEY_STATE);
  const dialog = useActionDialog({ state, closeOnSuccess: false });

  return (
    <>
      <button
        type="button"
        className="border-critical text-critical hover:bg-critical/10 active:bg-critical/15 focus-visible:ring-critical rounded-xl border px-4 py-2.5 text-sm font-semibold transition-colors focus-visible:ring-2 focus-visible:outline-none"
        onClick={() => dialog.setOpen(true)}
      >
        Delete
      </button>
      <DeleteApiKeyContent
        key={dialog.formKey}
        apiKey={apiKey}
        open={dialog.open}
        onClose={dialog.reset}
        onStateChange={setState}
      />
    </>
  );
}

function DeleteApiKeyContent({
  apiKey,
  open,
  onClose,
  onStateChange,
}: {
  apiKey: ApiKeyResponse;
  open: boolean;
  onClose: () => void;
  onStateChange: (state: typeof EMPTY_API_KEY_STATE) => void;
}) {
  const [confirmation, setConfirmation] = useState("");
  const [state, action, pending] = useActionState(
    deleteApiKeyAction,
    EMPTY_API_KEY_STATE,
  );
  const formRef = useRef<HTMLFormElement>(null);
  useRefreshAfterAction(state);
  useFirstInvalidField(state, formRef);
  useEffect(() => onStateChange(state), [onStateChange, state]);
  const id = useId();

  return (
    <Dialog
      open={open}
      onClose={onClose}
      title="Delete API key"
      variant="sheet"
      dismissible={!pending}
    >
      {state.status === "success" ? (
        <div className="space-y-4">
          <p className="text-success">{state.message}</p>
          <Button type="button" onClick={onClose}>
            Done
          </Button>
        </div>
      ) : (
        <form
          ref={formRef}
          action={action}
          onReset={(event) => event.preventDefault()}
          className="space-y-4"
        >
          <input type="hidden" name="id" value={apiKey.id} />
          <input type="hidden" name="username" value={apiKey.user_username} />
          <p className="text-sm">
            This permanently removes the API key for{" "}
            <strong>
              {apiKey.user_name} (@{apiKey.user_username})
            </strong>
            . This cannot be undone.
          </p>
          <div>
            <Label htmlFor={`${id}-confirmation`}>
              Type {apiKey.user_username} to confirm
            </Label>
            <Input
              id={`${id}-confirmation`}
              name="confirmation"
              value={confirmation}
              onChange={(event) => setConfirmation(event.target.value)}
              autoComplete="off"
            />
            <FieldError>{state.fieldErrors?.confirmation}</FieldError>
          </div>
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
            <button
              type="submit"
              disabled={pending || confirmation !== apiKey.user_username}
              className="bg-critical text-background rounded-xl px-4 py-2.5 text-sm font-semibold transition-opacity hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-50"
            >
              {pending ? "Deleting…" : "Permanently delete"}
            </button>
          </div>
        </form>
      )}
    </Dialog>
  );
}
