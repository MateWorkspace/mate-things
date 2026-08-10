"use client";

import { useActionState, useState } from "react";

import ActionMessage from "@/components/forms/ActionMessage";
import Button from "@/components/ui/button";
import { useRefreshAfterAction } from "@/hooks/use-refresh-after-action";
import type { ActionResponse } from "@/lib/api/actions";

import {
  EMPTY_NODE_CLASS_ASSIGNMENT_STATE,
  updateNodeClassActionsAction,
} from "../_lib/actions";

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
    EMPTY_NODE_CLASS_ASSIGNMENT_STATE,
  );
  useRefreshAfterAction(state);
  const [assignmentSelection, setAssignmentSelection] = useState(() => ({
    handled: state,
    overrides: new Map<string, boolean>(),
    settled: new Set<string>(),
  }));
  if (assignmentSelection.handled !== state) {
    const settled = new Set(assignmentSelection.settled);
    state.appliedIds.forEach((id) => settled.add(id));
    state.failed.forEach(({ id }) => settled.delete(id));
    setAssignmentSelection({
      ...assignmentSelection,
      handled: state,
      settled,
    });
  }
  const isSelected = (id: string) =>
    assignmentSelection.overrides.has(id) &&
    !assignmentSelection.settled.has(id)
      ? (assignmentSelection.overrides.get(id) ?? false)
      : selected.has(id);

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
                  checked={isSelected(action.id)}
                  onChange={(event) => {
                    const checked = event.target.checked;
                    setAssignmentSelection((current) => {
                      const overrides = new Map(current.overrides);
                      const settled = new Set(current.settled);
                      overrides.set(action.id, checked);
                      settled.delete(action.id);
                      return { ...current, overrides, settled };
                    });
                  }}
                  className="accent-primary mt-1 size-4"
                />
              ) : (
                <span aria-hidden="true" className="mt-1">
                  {isSelected(action.id) ? "✓" : "—"}
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
