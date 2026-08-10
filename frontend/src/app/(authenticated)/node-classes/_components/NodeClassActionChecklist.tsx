"use client";

import { useActionState } from "react";

import ActionMessage from "@/components/forms/ActionMessage";
import Button from "@/components/ui/button";
import {
  assignmentSubmission,
  type AssignmentSubmission,
  useAssignmentSelection,
} from "@/hooks/use-assignment-selection";
import { useRefreshAfterAction } from "@/hooks/use-refresh-after-action";
import type { ActionResponse } from "@/lib/api/actions";

import {
  EMPTY_NODE_CLASS_ASSIGNMENT_STATE,
  updateNodeClassActionsAction,
  type AssignmentActionState,
} from "../_lib/actions";

type NodeClassAssignmentClientState = AssignmentActionState & {
  submission?: AssignmentSubmission;
};

async function updateNodeClassActionsWithSubmission(
  previous: NodeClassAssignmentClientState,
  data: FormData,
): Promise<NodeClassAssignmentClientState> {
  const result = await updateNodeClassActionsAction(previous, data);
  return {
    ...result,
    submission: assignmentSubmission(data, "action_ids"),
  };
}

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
    updateNodeClassActionsWithSubmission,
    EMPTY_NODE_CLASS_ASSIGNMENT_STATE,
  );
  useRefreshAfterAction(state);
  const assignmentSelection = useAssignmentSelection(
    state,
    selected,
    actions.map(({ id }) => id),
  );

  return (
    <form
      action={formAction}
      onReset={(event) => event.preventDefault()}
      className="space-y-4"
    >
      <input type="hidden" name="node_class_id" value={nodeClassId} />
      <input
        type="hidden"
        name="assignment_snapshot"
        value={assignmentSelection.submissionSnapshot}
      />
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
                  key={`${action.id}:${assignmentSelection.resultGeneration}`}
                  type="checkbox"
                  name="action_ids"
                  value={action.id}
                  checked={assignmentSelection.isSelected(action.id)}
                  onChange={(event) => {
                    assignmentSelection.setSelected(
                      action.id,
                      event.target.checked,
                    );
                  }}
                  className="accent-primary mt-1 size-4"
                />
              ) : (
                <span aria-hidden="true" className="mt-1">
                  {assignmentSelection.isSelected(action.id) ? "✓" : "—"}
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
