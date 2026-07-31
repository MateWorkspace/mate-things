"use client";

import { useRouter } from "next/navigation";
import { useActionState, useEffect, useState } from "react";

import JsonEditor from "@/components/json/JsonEditor";
import Button from "@/components/ui/button";
import Dialog from "@/components/ui/dialog";
import type { PreferencesResource } from "@/lib/api/preferences";

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
  const [open, setOpen] = useState(false);
  const [state, action, pending] = useActionState(
    savePreferencesAction,
    EMPTY_PREFERENCES_STATE,
  );
  const router = useRouter();
  useEffect(() => {
    if (state.status === "success") {
      router.refresh();
    }
  }, [state, router]);

  if (!permissions.includes("preferences:set")) return null;

  return (
    <>
      <Button type="button" variant="secondary" onClick={() => setOpen(true)}>
        {label}
      </Button>
      <Dialog
        open={open && state.status !== "success"}
        onClose={() => !pending && setOpen(false)}
        title={`Edit ${label.toLowerCase()}`}
        variant="sheet"
      >
        <form action={action} className="space-y-4">
          <input type="hidden" name="resource" value={resource} />
          <input type="hidden" name="id" value={id} />
          <JsonEditor
            name="preferences"
            label="Preferences JSON"
            defaultValue={preferences}
          />
          <p
            aria-live="polite"
            className={
              state.status === "error"
                ? "text-critical text-sm"
                : "text-success text-sm"
            }
          >
            {state.message}
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
            <Button type="submit" disabled={pending}>
              {pending ? "Saving…" : "Save preferences"}
            </Button>
          </div>
        </form>
      </Dialog>
    </>
  );
}
