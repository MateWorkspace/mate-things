"use client";

import { useRouter } from "next/navigation";
import { useActionState, useState } from "react";

import Button from "@/components/ui/button";
import CopyButton from "@/components/ui/copy-button";
import Dialog from "@/components/ui/dialog";
import Label from "@/components/ui/label";
import type { ApiKeyResponse } from "@/lib/api/api-keys";

import { regenerateApiKeyAction } from "../_lib/actions";
import { EMPTY_API_KEY_STATE } from "../_lib/state";

function localDate(value?: string): string {
  if (!value) return "";
  const date = new Date(value);
  const shifted = new Date(date.valueOf() - date.getTimezoneOffset() * 60000);
  return shifted.toISOString().slice(0, 16);
}

export default function RegenerateApiKeyDialog({
  apiKey,
}: {
  apiKey: ApiKeyResponse;
}) {
  const [open, setOpen] = useState(false);
  const [state, action, pending] = useActionState(
    regenerateApiKeyAction,
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
        Regenerate
      </Button>
      <Dialog
        open={open}
        onClose={close}
        title={`Regenerate ${apiKey.user_username}'s API key`}
        variant="sheet"
      >
        {state.status === "success" && state.key ? (
          <div className="space-y-4">
            <p className="text-success text-sm">{state.message}</p>
            <div className="border-border bg-muted flex items-center justify-between gap-3 rounded-xl border p-3">
              <code className="overflow-x-auto text-xs break-all">
                {state.key}
              </code>
              <CopyButton value={state.key} />
            </div>
            <div className="flex justify-end">
              <Button type="button" onClick={close}>
                Done
              </Button>
            </div>
          </div>
        ) : (
          <form action={action} className="space-y-4">
            <input type="hidden" name="id" value={apiKey.id} />
            <p className="text-muted-foreground text-sm">
              This issues a brand new value for{" "}
              <strong className="text-foreground">
                {apiKey.user_name} (@{apiKey.user_username})
              </strong>
              . The current key stops working immediately.
            </p>
            <div>
              <Label htmlFor={`regenerate-${apiKey.id}-expires-at`}>
                Expires at (optional)
              </Label>
              <input
                id={`regenerate-${apiKey.id}-expires-at`}
                name="expires_at"
                type="datetime-local"
                defaultValue={localDate(apiKey.expires_at)}
                className="border-control-border bg-background text-foreground focus-visible:border-focus focus-visible:ring-focus min-h-11 w-full rounded-xl border px-3.5 py-2.5 text-sm transition-colors focus-visible:ring-2 focus-visible:outline-none"
              />
              <p className="text-muted-foreground mt-1.5 text-xs">
                Leave blank for a key that never expires.
              </p>
              {state.fieldErrors?.expires_at ? (
                <p className="text-critical mt-1 text-sm">
                  {state.fieldErrors.expires_at}
                </p>
              ) : null}
            </div>
            {state.status === "error" ? (
              <p aria-live="assertive" className="text-critical text-sm">
                {state.message}
              </p>
            ) : null}
            <div className="flex justify-end gap-2">
              <Button
                type="button"
                variant="secondary"
                onClick={() => setOpen(false)}
              >
                Cancel
              </Button>
              <Button type="submit" disabled={pending}>
                {pending ? "Regenerating…" : "Regenerate"}
              </Button>
            </div>
          </form>
        )}
      </Dialog>
    </>
  );
}
