"use client";

import { useRouter } from "next/navigation";
import { useActionState, useState } from "react";

import Button from "@/components/ui/button";
import Dialog from "@/components/ui/dialog";
import type { ApiKeyResponse } from "@/lib/api/api-keys";

import { revokeApiKeyAction } from "../_lib/actions";
import { EMPTY_API_KEY_STATE } from "../_lib/state";

export default function RevokeApiKeyDialog({
  apiKey,
}: {
  apiKey: ApiKeyResponse;
}) {
  const [open, setOpen] = useState(false);
  const [state, action, pending] = useActionState(
    revokeApiKeyAction,
    EMPTY_API_KEY_STATE,
  );
  const router = useRouter();

  function close() {
    setOpen(false);
    if (state.status === "success") {
      router.refresh();
    }
  }

  return (
    <>
      <Button type="button" variant="secondary" onClick={() => setOpen(true)}>
        Revoke
      </Button>
      <Dialog open={open} onClose={close} title="Revoke API key">
        {state.status === "success" ? (
          <div className="space-y-4">
            <p className="text-success text-sm">{state.message}</p>
            <div className="flex justify-end">
              <Button type="button" onClick={close}>
                Done
              </Button>
            </div>
          </div>
        ) : (
          <form action={action} className="space-y-4">
            <input type="hidden" name="id" value={apiKey.id} />
            <p className="text-sm">
              Revoke the API key for{" "}
              <strong>
                {apiKey.user_name} (@{apiKey.user_username})
              </strong>
              ? It will stop authenticating immediately. You can reactivate it
              later by regenerating.
            </p>
            {state.status === "error" ? (
              <p aria-live="assertive" className="text-critical text-sm">
                {state.message}
              </p>
            ) : null}
            <div className="flex justify-end gap-2">
              <Button
                type="button"
                variant="secondary"
                disabled={pending}
                onClick={() => setOpen(false)}
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
    </>
  );
}
