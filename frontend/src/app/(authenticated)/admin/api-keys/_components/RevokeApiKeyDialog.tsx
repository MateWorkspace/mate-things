"use client";

import { useActionState, useState } from "react";

import ActionMessage from "@/components/forms/ActionMessage";
import Button from "@/components/ui/button";
import Dialog from "@/components/ui/dialog";
import type { ApiKeyResponse } from "@/lib/api/api-keys";
import { useRefreshAfterAction } from "@/hooks/use-refresh-after-action";

import { revokeApiKeyAction } from "../_lib/actions";
import { EMPTY_API_KEY_STATE } from "../_lib/state";

export default function RevokeApiKeyDialog({
  apiKey,
}: {
  apiKey: ApiKeyResponse;
}) {
  const [open, setOpen] = useState(false);
  const [generation, setGeneration] = useState(0);
  return (
    <>
      <Button
        type="button"
        variant="secondary"
        onClick={() => {
          setGeneration((current) => current + 1);
          setOpen(true);
        }}
      >
        Revoke
      </Button>
      <RevokeApiKeyContent
        key={generation}
        apiKey={apiKey}
        open={open}
        onClose={() => setOpen(false)}
      />
    </>
  );
}

function RevokeApiKeyContent({
  apiKey,
  open,
  onClose,
}: {
  apiKey: ApiKeyResponse;
  open: boolean;
  onClose: () => void;
}) {
  const [state, action, pending] = useActionState(
    revokeApiKeyAction,
    EMPTY_API_KEY_STATE,
  );
  useRefreshAfterAction(state);

  return (
    <Dialog
      open={open}
      onClose={onClose}
      title="Revoke API key"
      dismissible={!pending}
    >
      {state.status === "success" ? (
        <div className="space-y-4">
          <p className="text-success text-sm">{state.message}</p>
          <div className="flex justify-end">
            <Button type="button" onClick={onClose}>
              Done
            </Button>
          </div>
        </div>
      ) : (
        <form
          action={action}
          onReset={(event) => event.preventDefault()}
          className="space-y-4"
        >
          <input type="hidden" name="id" value={apiKey.id} />
          <p className="text-sm">
            Revoke the API key for{" "}
            <strong>
              {apiKey.user_name} (@{apiKey.user_username})
            </strong>
            ? It will stop authenticating immediately. You can reactivate it
            later by regenerating.
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
              {pending ? "Revoking…" : "Revoke"}
            </Button>
          </div>
        </form>
      )}
    </Dialog>
  );
}
