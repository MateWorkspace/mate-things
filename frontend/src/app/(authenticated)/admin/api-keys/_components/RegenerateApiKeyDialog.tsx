"use client";

import { useActionState, useRef, useState } from "react";

import ActionMessage from "@/components/forms/ActionMessage";
import FieldError from "@/components/forms/FieldError";
import Button from "@/components/ui/button";
import CopyButton from "@/components/ui/copy-button";
import Dialog from "@/components/ui/dialog";
import Label from "@/components/ui/label";
import type { ApiKeyResponse } from "@/lib/api/api-keys";
import { useFirstInvalidField } from "@/hooks/use-first-invalid-field";
import { useRefreshAfterAction } from "@/hooks/use-refresh-after-action";

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
        Regenerate
      </Button>
      <RegenerateApiKeyContent
        key={generation}
        apiKey={apiKey}
        open={open}
        onClose={() => setOpen(false)}
      />
    </>
  );
}

function RegenerateApiKeyContent({
  apiKey,
  open,
  onClose,
}: {
  apiKey: ApiKeyResponse;
  open: boolean;
  onClose: () => void;
}) {
  const [state, action, pending] = useActionState(
    regenerateApiKeyAction,
    EMPTY_API_KEY_STATE,
  );
  const formRef = useRef<HTMLFormElement>(null);
  useRefreshAfterAction(state);
  useFirstInvalidField(state, formRef);

  return (
    <Dialog
      open={open}
      onClose={onClose}
      title={`Regenerate ${apiKey.user_username}'s API key`}
      variant="sheet"
      dismissible={!pending}
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
            <Button type="button" onClick={onClose}>
              Done
            </Button>
          </div>
        </div>
      ) : (
        <form
          ref={formRef}
          action={action}
          onReset={(event) => event.preventDefault()}
          className="space-y-4"
        >
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
            <FieldError>{state.fieldErrors?.expires_at}</FieldError>
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
            <Button type="submit" disabled={pending}>
              {pending ? "Regenerating…" : "Regenerate"}
            </Button>
          </div>
        </form>
      )}
    </Dialog>
  );
}
