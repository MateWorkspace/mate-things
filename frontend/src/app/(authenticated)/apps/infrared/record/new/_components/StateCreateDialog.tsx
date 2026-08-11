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
import type { InfraredStateType } from "@/lib/api/infrared";

import { createStateForWizardAction } from "../_lib/actions";
import {
  EMPTY_WIZARD_CREATE_STATE,
  type WizardCreatedItem,
} from "../_lib/state";

export default function StateCreateDialog({
  deviceTypeId,
  onCreated,
}: {
  deviceTypeId: string;
  onCreated: (item: WizardCreatedItem, type: InfraredStateType) => void;
}) {
  const [state, setState] = useState(EMPTY_WIZARD_CREATE_STATE);
  const [type, setType] = useState<InfraredStateType>("ENUM");
  const dialog = useActionDialog({
    state,
    onSuccess: () => {
      if (state.created) {
        onCreated(state.created, type);
      }
    },
  });

  return (
    <>
      <Button
        type="button"
        variant="secondary"
        disabled={!deviceTypeId}
        onClick={() => dialog.setOpen(true)}
      >
        New state
      </Button>
      <StateCreateContent
        key={dialog.formKey}
        deviceTypeId={deviceTypeId}
        open={dialog.open}
        onClose={dialog.reset}
        onStateChange={setState}
        onTypeChange={setType}
      />
    </>
  );
}

function StateCreateContent({
  deviceTypeId,
  open,
  onClose,
  onStateChange,
  onTypeChange,
}: {
  deviceTypeId: string;
  open: boolean;
  onClose: () => void;
  onStateChange: (state: typeof EMPTY_WIZARD_CREATE_STATE) => void;
  onTypeChange: (type: InfraredStateType) => void;
}) {
  const [state, action, pending] = useActionState(
    createStateForWizardAction,
    EMPTY_WIZARD_CREATE_STATE,
  );
  const fieldId = useId();
  const formRef = useRef<HTMLFormElement>(null);
  useFirstInvalidField(state, formRef);
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
            placeholder="Mode"
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
          <Select
            id={`${fieldId}-type`}
            name="type"
            defaultValue="ENUM"
            onChange={(event) =>
              onTypeChange(event.target.value as InfraredStateType)
            }
          >
            <option value="ENUM">Enum (fixed options)</option>
            <option value="RANGE">Range (min/max/step)</option>
          </Select>
          <FieldError id={`${fieldId}-type-error`}>
            {state.fieldErrors?.type}
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
            {pending ? "Creating…" : "Create state"}
          </Button>
        </div>
      </form>
    </Dialog>
  );
}
