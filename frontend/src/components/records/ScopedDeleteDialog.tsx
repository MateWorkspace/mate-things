"use client";

import { useActionState, useEffect, useId, useState } from "react";

import ActionMessage from "@/components/forms/ActionMessage";
import Button from "@/components/ui/button";
import Dialog from "@/components/ui/dialog";
import Input from "@/components/ui/input";
import Label from "@/components/ui/label";
import { useActionDialog } from "@/hooks/use-action-dialog";
import { useRefreshAfterAction } from "@/hooks/use-refresh-after-action";
import type { ActionState } from "@/lib/forms/action-state";

export type ScopedDeleteState = ActionState<string> & {
  count?: number;
};

interface ScopedDeleteDialogProps {
  action: (
    state: ScopedDeleteState,
    formData: FormData,
  ) => Promise<ScopedDeleteState>;
  filters: readonly [string, string][];
  label: string;
}

const INITIAL_STATE: ScopedDeleteState = { status: "idle" };

export default function ScopedDeleteDialog({
  action,
  filters,
  label,
}: ScopedDeleteDialogProps) {
  const [state, setState] = useState<ScopedDeleteState>(INITIAL_STATE);
  const dialog = useActionDialog({ state, closeOnSuccess: false });

  return (
    <>
      <button
        type="button"
        className="border-critical text-critical hover:bg-critical/10 active:bg-critical/15 focus-visible:ring-critical rounded-xl border px-4 py-2.5 text-sm font-semibold transition-colors focus-visible:ring-2 focus-visible:outline-none"
        onClick={() => dialog.setOpen(true)}
      >
        Delete filtered {label}
      </button>
      <ScopedDeleteContent
        key={dialog.formKey}
        action={action}
        filters={filters}
        label={label}
        open={dialog.open}
        onClose={dialog.reset}
        onStateChange={setState}
      />
    </>
  );
}

function ScopedDeleteContent({
  action,
  filters,
  label,
  open,
  onClose,
  onStateChange,
}: ScopedDeleteDialogProps & {
  open: boolean;
  onClose: () => void;
  onStateChange: (state: ScopedDeleteState) => void;
}) {
  const [confirmation, setConfirmation] = useState("");
  const [state, formAction, pending] = useActionState(action, INITIAL_STATE);
  useRefreshAfterAction(state);
  useEffect(() => onStateChange(state), [onStateChange, state]);
  const id = useId();
  const phrase = filters.length ? "DELETE" : "DELETE ALL";

  return (
    <Dialog
      open={open}
      onClose={onClose}
      title={`Delete ${label}`}
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
          action={formAction}
          onReset={(event) => event.preventDefault()}
          className="space-y-4"
        >
          {filters.map(([name, value]) => (
            <input key={name} type="hidden" name={name} value={value} />
          ))}
          <div className="border-critical/30 bg-critical/5 rounded-xl border p-4">
            <p className="font-semibold">Deletion scope</p>
            {filters.length ? (
              <dl className="mt-2 space-y-1 text-sm">
                {filters.map(([name, value]) => (
                  <div key={name} className="flex justify-between gap-4">
                    <dt>{name.replaceAll("_", " ")}</dt>
                    <dd className="max-w-64 text-right font-mono text-xs break-all">
                      {value}
                    </dd>
                  </div>
                ))}
              </dl>
            ) : (
              <p className="text-critical mt-2 text-sm font-semibold">
                No filters are active. This targets every record.
              </p>
            )}
          </div>
          <div>
            <Label htmlFor={`${id}-confirmation`}>
              Type {phrase} to confirm
            </Label>
            <Input
              id={`${id}-confirmation`}
              name="confirmation"
              value={confirmation}
              onChange={(event) => setConfirmation(event.target.value)}
              autoComplete="off"
            />
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
              disabled={pending || confirmation !== phrase}
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
