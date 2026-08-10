"use client";

import { useActionState, useRef, useState } from "react";

import ActionMessage from "@/components/forms/ActionMessage";
import FieldError from "@/components/forms/FieldError";
import Button from "@/components/ui/button";
import CopyButton from "@/components/ui/copy-button";
import Dialog from "@/components/ui/dialog";
import Label from "@/components/ui/label";
import UserSearchCombobox from "@/components/users/UserSearchCombobox";
import { useRefreshAfterAction } from "@/hooks/use-refresh-after-action";
import { useFirstInvalidField } from "@/hooks/use-first-invalid-field";

import { generateApiKeyAction } from "../_lib/actions";
import { EMPTY_API_KEY_STATE } from "../_lib/state";

export default function GenerateApiKeyDialog() {
  const [open, setOpen] = useState(false);
  const [generation, setGeneration] = useState(0);
  return (
    <>
      <Button
        type="button"
        onClick={() => {
          setGeneration((current) => current + 1);
          setOpen(true);
        }}
      >
        Generate API key
      </Button>
      <GenerateApiKeyContent
        key={generation}
        open={open}
        onClose={() => setOpen(false)}
      />
    </>
  );
}

function GenerateApiKeyContent({
  open,
  onClose,
}: {
  open: boolean;
  onClose: () => void;
}) {
  const [state, action, pending] = useActionState(
    generateApiKeyAction,
    EMPTY_API_KEY_STATE,
  );
  const formRef = useRef<HTMLFormElement>(null);
  useRefreshAfterAction(state);
  useFirstInvalidField(state, formRef);

  return (
    <Dialog
      open={open}
      onClose={onClose}
      title="Generate API key"
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
          <UserSearchCombobox name="user_id" />
          <FieldError>{state.fieldErrors?.user_id}</FieldError>
          <div>
            <Label htmlFor="generate-api-key-expires-at">
              Expires at (optional)
            </Label>
            <input
              id="generate-api-key-expires-at"
              name="expires_at"
              type="datetime-local"
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
              {pending ? "Generating…" : "Generate"}
            </Button>
          </div>
        </form>
      )}
    </Dialog>
  );
}
