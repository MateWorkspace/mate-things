"use client";

import { useRouter } from "next/navigation";
import { useActionState, useId, useState } from "react";

import Button from "@/components/ui/button";
import Dialog from "@/components/ui/dialog";
import Input from "@/components/ui/input";
import Label from "@/components/ui/label";
import type { ApiKeyResponse } from "@/lib/api/api-keys";

import { deleteApiKeyAction } from "../_lib/actions";
import { EMPTY_API_KEY_STATE } from "../_lib/state";

export default function DeleteApiKeyDialog({
  apiKey,
}: {
  apiKey: ApiKeyResponse;
}) {
  const [open, setOpen] = useState(false);
  const [confirmation, setConfirmation] = useState("");
  const [state, action, pending] = useActionState(
    deleteApiKeyAction,
    EMPTY_API_KEY_STATE,
  );
  const router = useRouter();
  const id = useId();

  function close() {
    setOpen(false);
    if (state.status === "success") {
      router.refresh();
    }
  }

  return (
    <>
      <button
        type="button"
        className="border-critical text-critical hover:bg-critical/10 active:bg-critical/15 focus-visible:ring-critical rounded-xl border px-4 py-2.5 text-sm font-semibold transition-colors focus-visible:ring-2 focus-visible:outline-none"
        onClick={() => setOpen(true)}
      >
        Delete
      </button>
      <Dialog open={open} onClose={close} title="Delete API key" variant="sheet">
        {state.status === "success" ? (
          <div className="space-y-4">
            <p className="text-success">{state.message}</p>
            <Button type="button" onClick={close}>
              Done
            </Button>
          </div>
        ) : (
          <form action={action} className="space-y-4">
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
              {state.fieldErrors?.confirmation ? (
                <p className="text-critical mt-1 text-sm">
                  {state.fieldErrors.confirmation}
                </p>
              ) : null}
            </div>
            <p aria-live="assertive" className="text-critical text-sm">
              {state.status === "error" ? state.message : null}
            </p>
            <div className="flex justify-end gap-2">
              <Button
                type="button"
                variant="secondary"
                disabled={pending}
                onClick={() => setOpen(false)}
              >
                Cancel
              </Button>
              <button
                type="submit"
                disabled={pending || confirmation !== apiKey.user_username}
                className="bg-critical text-background hover:opacity-90 rounded-xl px-4 py-2.5 text-sm font-semibold transition-opacity disabled:cursor-not-allowed disabled:opacity-50"
              >
                {pending ? "Deleting…" : "Permanently delete"}
              </button>
            </div>
          </form>
        )}
      </Dialog>
    </>
  );
}
