"use client";

import { useActionState, useEffect, useState } from "react";

import ActionMessage from "@/components/forms/ActionMessage";
import JsonEditor from "@/components/json/JsonEditor";
import Button from "@/components/ui/button";
import Dialog from "@/components/ui/dialog";
import { useActionDialog } from "@/hooks/use-action-dialog";
import type { PreferencesResource } from "@/lib/api/preferences";
import { useRefreshAfterAction } from "@/hooks/use-refresh-after-action";

import { savePreferencesAction } from "./preferences-actions";
import { EMPTY_PREFERENCES_STATE } from "./preferences-state";

interface PreferencesDialogProps {
  resource: PreferencesResource;
  id: string;
  preferences: Record<string, unknown>;
  permissions: readonly string[];
  label?: string;
}

export default function PreferencesDialog({
  resource,
  id,
  preferences,
  permissions,
  label = "Preferences",
}: PreferencesDialogProps) {
  const [state, setState] = useState(EMPTY_PREFERENCES_STATE);
  const dialog = useActionDialog({ state });

  if (!permissions.includes("preferences:set")) return null;

  return (
    <>
      <Button
        type="button"
        variant="secondary"
        onClick={() => dialog.setOpen(true)}
      >
        {label}
      </Button>
      <PreferencesDialogContent
        key={dialog.formKey}
        resource={resource}
        id={id}
        preferences={preferences}
        label={label}
        open={dialog.open}
        onClose={dialog.reset}
        onStateChange={setState}
      />
    </>
  );
}

function PreferencesDialogContent({
  resource,
  id,
  preferences,
  label,
  open,
  onClose,
  onStateChange,
}: Pick<PreferencesDialogProps, "resource" | "id" | "preferences" | "label"> & {
  open: boolean;
  onClose: () => void;
  onStateChange: (state: typeof EMPTY_PREFERENCES_STATE) => void;
}) {
  const [state, action, pending] = useActionState(
    savePreferencesAction,
    EMPTY_PREFERENCES_STATE,
  );
  useEffect(() => onStateChange(state), [onStateChange, state]);
  useRefreshAfterAction(state);

  return (
    <Dialog
      open={open && state.status !== "success"}
      onClose={onClose}
      title={`Edit ${(label ?? "Preferences").toLowerCase()}`}
      variant="sheet"
      dismissible={!pending}
    >
      <form
        action={action}
        onReset={(event) => event.preventDefault()}
        className="space-y-4"
      >
        <input type="hidden" name="resource" value={resource} />
        <input type="hidden" name="id" value={id} />
        <JsonEditor
          name="preferences"
          label="Preferences JSON"
          defaultValue={preferences}
        />
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
            {pending ? "Saving…" : "Save preferences"}
          </Button>
        </div>
      </form>
    </Dialog>
  );
}
