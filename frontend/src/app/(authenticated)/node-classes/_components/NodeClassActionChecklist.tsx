"use client";

import { useActionState } from "react";

import ActionMessage from "@/components/forms/ActionMessage";
import Button from "@/components/ui/button";
import { useRefreshAfterAction } from "@/hooks/use-refresh-after-action";
import type { ActionResponse } from "@/lib/api/actions";

import { updateNodeClassActionsAction } from "../_lib/actions";
import type { FormActionState } from "../_lib/actions";

const EMPTY_STATE: FormActionState = { status: "idle" };

export default function NodeClassActionChecklist({
  nodeClassId,
  actions,
  selected,
  editable,
}: {
  nodeClassId: string;
  actions: readonly ActionResponse[];
  selected: ReadonlySet<string>;
  editable: boolean;
}) {
  const [state, formAction, pending] = useActionState(
    updateNodeClassActionsAction,
    EMPTY_STATE,
  );
  useRefreshAfterAction(state);

  return (
    <form
      action={formAction}
      onReset={(event) => event.preventDefault()}
      className="space-y-4"
    >
      <input type="hidden" name="node_class_id" value={nodeClassId} />
      <fieldset className="border-border rounded-xl border p-4">
        <legend className="sr-only">Compatible actions</legend>
        <div className="space-y-2">
          {actions.map((action) => (
            <label
              key={action.id}
              className="bg-muted flex gap-3 rounded-lg p-3"
            >
              {editable ? (
                <input
                  type="checkbox"
                  name="action_ids"
                  value={action.id}
                  defaultChecked={selected.has(action.id)}
                  className="accent-primary mt-1 size-4"
                />
              ) : (
                <span aria-hidden="true" className="mt-1">
                  {selected.has(action.id) ? "✓" : "—"}
                </span>
              )}
              <span>
                <span className="block text-sm font-semibold">
                  {action.name}
                </span>
                <span className="text-muted-foreground text-xs">
                  {action.description || "No description."}
                </span>
              </span>
            </label>
          ))}
        </div>
      </fieldset>
      <ActionMessage state={state} />
      {editable ? (
        <Button type="submit" disabled={pending}>
          {pending ? "Saving…" : "Save assignments"}
        </Button>
      ) : null}
    </form>
  );
}
