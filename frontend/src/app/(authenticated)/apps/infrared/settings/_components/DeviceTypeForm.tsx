"use client";

import { useActionState, useEffect, useId, useRef, useState } from "react";

import ActionMessage from "@/components/forms/ActionMessage";
import FieldError from "@/components/forms/FieldError";
import Button from "@/components/ui/button";
import Dialog from "@/components/ui/dialog";
import Input from "@/components/ui/input";
import Label from "@/components/ui/label";
import { useActionDialog } from "@/hooks/use-action-dialog";
import { useFirstInvalidField } from "@/hooks/use-first-invalid-field";
import { useRefreshAfterAction } from "@/hooks/use-refresh-after-action";

import { createDeviceTypeAction } from "../_lib/actions";
import { EMPTY_INFRARED_SETTINGS_STATE } from "../_lib/state";

export default function DeviceTypeForm() {
  const [state, setState] = useState(EMPTY_INFRARED_SETTINGS_STATE);
  const dialog = useActionDialog({ state });

  return (
    <>
      <Button type="button" onClick={() => dialog.setOpen(true)}>
        New device type
      </Button>
      <DeviceTypeFormContent
        key={dialog.formKey}
        open={dialog.open}
        onClose={dialog.reset}
        onStateChange={setState}
      />
    </>
  );
}

function DeviceTypeFormContent({
  open,
  onClose,
  onStateChange,
}: {
  open: boolean;
  onClose: () => void;
  onStateChange: (state: typeof EMPTY_INFRARED_SETTINGS_STATE) => void;
}) {
  const [state, action, pending] = useActionState(
    createDeviceTypeAction,
    EMPTY_INFRARED_SETTINGS_STATE,
  );
  const fieldId = useId();
  const formRef = useRef<HTMLFormElement>(null);
  useFirstInvalidField(state, formRef);
  useRefreshAfterAction(state);
  useEffect(() => onStateChange(state), [onStateChange, state]);

  return (
    <Dialog
      open={open}
      onClose={onClose}
      title="Create device type"
      variant="sheet"
      dismissible={!pending}
    >
      <form
        ref={formRef}
        action={action}
        onReset={(event) => event.preventDefault()}
        className="space-y-4"
      >
        <div>
          <Label htmlFor={`${fieldId}-name`}>Name</Label>
          <Input
            id={`${fieldId}-name`}
            name="name"
            placeholder="Air Conditioner"
            required
            aria-invalid={Boolean(state.fieldErrors?.name)}
            aria-describedby={
              state.fieldErrors?.name ? `${fieldId}-name-error` : undefined
            }
          />
          <FieldError id={`${fieldId}-name-error`}>
            {state.fieldErrors?.name}
          </FieldError>
        </div>

        <ActionMessage state={state} />

        <div className="flex flex-wrap justify-end gap-2">
          <Button
            type="button"
            variant="secondary"
            disabled={pending}
            onClick={onClose}
          >
            Cancel
          </Button>
          <Button type="submit" disabled={pending}>
            {pending ? "Creating…" : "Create device type"}
          </Button>
        </div>
      </form>
    </Dialog>
  );
}
