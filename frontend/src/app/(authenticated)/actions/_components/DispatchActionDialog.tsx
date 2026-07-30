"use client";

import Link from "next/link";
import { useActionState, useId, useState } from "react";

import Button from "@/components/ui/button";
import Dialog from "@/components/ui/dialog";
import Input from "@/components/ui/input";
import Label from "@/components/ui/label";
import type { ActionResponse } from "@/lib/api/actions";
import type { NodeResponse } from "@/lib/api/nodes";

import {
  dispatchActionFormAction,
  type ActionFormState,
} from "../_lib/actions";

const EMPTY_STATE: ActionFormState = { status: "idle" };

interface DispatchActionDialogProps {
  action: ActionResponse;
  nodes: readonly NodeResponse[];
}

export default function DispatchActionDialog({
  action,
  nodes,
}: DispatchActionDialogProps) {
  const [open, setOpen] = useState(false);
  const [state, formAction, pending] = useActionState(
    dispatchActionFormAction,
    EMPTY_STATE,
  );
  const id = useId();
  return (
    <>
      <Button type="button" onClick={() => setOpen(true)}>
        Dispatch action
      </Button>
      <Dialog
        open={open}
        onClose={() => setOpen(false)}
        title={`Dispatch ${action.name}`}
        variant="sheet"
      >
        {state.status === "success" && state.executionId ? (
          <div className="space-y-4">
            <p>{state.message}</p>
            <Link
              className="bg-primary text-surface inline-flex rounded-xl px-4 py-2.5 text-sm font-semibold"
              href={`/action-history?execution_id=${encodeURIComponent(state.executionId)}`}
            >
              View execution
            </Link>
          </div>
        ) : (
          <form action={formAction} className="space-y-4">
            <input type="hidden" name="action_id" value={action.id} />
            <div>
              <Label htmlFor={`${id}-node`}>Compatible node</Label>
              <select
                id={`${id}-node`}
                name="node_id"
                required
                className="border-control-border bg-background focus-visible:ring-focus min-h-11 w-full rounded-xl border px-3.5 text-sm focus-visible:ring-2 focus-visible:outline-none"
              >
                <option value="">Select node</option>
                {nodes
                  .filter((node) => node.node_class_id === action.node_class_id)
                  .map((node) => (
                    <option key={node.id} value={node.id}>
                      {node.name} · {node.device_id}
                    </option>
                  ))}
              </select>
              {nodes.every(
                (node) => node.node_class_id !== action.node_class_id,
              ) ? (
                <p className="text-warning mt-2 text-sm">
                  No compatible nodes are available.
                </p>
              ) : null}
            </div>
            <div>
              <Label htmlFor={`${id}-payload`}>Payload JSON</Label>
              <textarea
                id={`${id}-payload`}
                name="payload"
                rows={8}
                defaultValue={"{}"}
                className="border-control-border bg-background focus-visible:ring-focus w-full rounded-xl border p-3 font-mono text-sm focus-visible:ring-2 focus-visible:outline-none"
              />
              {state.fieldErrors?.payload ? (
                <p className="text-critical mt-1 text-sm">
                  {state.fieldErrors.payload}
                </p>
              ) : null}
            </div>
            <div>
              <Label htmlFor={`${id}-executed`}>Execute at (optional)</Label>
              <Input
                id={`${id}-executed`}
                name="executed_at"
                type="datetime-local"
              />
            </div>
            <p aria-live="polite" className="text-critical text-sm">
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
              <Button type="submit" disabled={pending || nodes.length === 0}>
                {pending ? "Dispatching…" : "Dispatch"}
              </Button>
            </div>
          </form>
        )}
      </Dialog>
    </>
  );
}
