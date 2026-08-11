"use client";

import { useActionState, useEffect, useId, useRef, useState } from "react";

import ActionMessage from "@/components/forms/ActionMessage";
import FieldError from "@/components/forms/FieldError";
import Button from "@/components/ui/button";
import Dialog from "@/components/ui/dialog";
import Input from "@/components/ui/input";
import Label from "@/components/ui/label";
import Select from "@/components/ui/select";
import { useActionDialog } from "@/hooks/use-action-dialog";
import { useFirstInvalidField } from "@/hooks/use-first-invalid-field";
import { useRefreshAfterAction } from "@/hooks/use-refresh-after-action";

import { createStateAction } from "../_lib/actions";
import { EMPTY_INFRARED_STATE_STATE } from "../_lib/state";

export default function StateForm({ deviceTypeId }: { deviceTypeId: string }) {
  const [state, setState] = useState(EMPTY_INFRARED_STATE_STATE);
  const dialog = useActionDialog({ state });

  return (
    <>
      <Button type="button" onClick={() => dialog.setOpen(true)}>
        New state
      </Button>
      <StateFormContent
        key={dialog.formKey}
        deviceTypeId={deviceTypeId}
        open={dialog.open}
        onClose={dialog.reset}
        onStateChange={setState}
      />
    </>
  );
}

function StateFormContent({
  deviceTypeId,
  open,
  onClose,
  onStateChange,
}: {
  deviceTypeId: string;
  open: boolean;
  onClose: () => void;
  onStateChange: (state: typeof EMPTY_INFRARED_STATE_STATE) => void;
}) {
  const [state, action, pending] = useActionState(
    createStateAction,
    EMPTY_INFRARED_STATE_STATE,
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
      title="Create state"
      variant="sheet"
      dismissible={!pending}
    >
      <form
        ref={formRef}
        action={action}
        onReset={(event) => event.preventDefault()}
        className="space-y-4"
      >
        <input type="hidden" name="device_type_id" value={deviceTypeId} />

        <div>
          <Label htmlFor={`${fieldId}-name`}>Name</Label>
          <Input
            id={`${fieldId}-name`}
            name="name"
            placeholder="POWER"
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

        <div>
          <Label htmlFor={`${fieldId}-type`}>Type</Label>
          <Select id={`${fieldId}-type`} name="type" defaultValue="ENUM">
            <option value="ENUM">Enum — a fixed set of options</option>
            <option value="RANGE">Range — a numeric min/max/step</option>
          </Select>
          <FieldError>{state.fieldErrors?.type}</FieldError>
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
            {pending ? "Creating…" : "Create state"}
          </Button>
        </div>
      </form>
    </Dialog>
  );
}
