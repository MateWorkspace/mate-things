"use client";

import Link from "next/link";
import { useActionState, useId, useRef, useState } from "react";

import ActionMessage from "@/components/forms/ActionMessage";
import FieldError from "@/components/forms/FieldError";
import Button from "@/components/ui/button";
import Dialog from "@/components/ui/dialog";
import Input from "@/components/ui/input";
import Label from "@/components/ui/label";
import type { ActionResponse } from "@/lib/api/actions";
import type { NodeResponse } from "@/lib/api/nodes";
import { useFirstInvalidField } from "@/hooks/use-first-invalid-field";

import {
  dispatchActionFormAction,
  type ActionFormState,
} from "../_lib/actions";

const EMPTY_STATE: ActionFormState = { status: "idle" };

interface DispatchActionDialogProps {
  action: ActionResponse;
  nodes: readonly NodeResponse[];
  compatibleNodeClassIds: ReadonlySet<string>;
}

export default function DispatchActionDialog({
  action,
  nodes,
  compatibleNodeClassIds,
}: DispatchActionDialogProps) {
  const [open, setOpen] = useState(false);
  const [generation, setGeneration] = useState(0);
  return (
    <>
      <Button
        type="button"
        onClick={() => {
          setGeneration((current) => current + 1);
          setOpen(true);
        }}
      >
        Dispatch action
      </Button>
      <DispatchDialogContent
        key={generation}
        action={action}
        nodes={nodes}
        compatibleNodeClassIds={compatibleNodeClassIds}
        open={open}
        onClose={() => setOpen(false)}
      />
    </>
  );
}

function DispatchDialogContent({
  action,
  nodes,
  compatibleNodeClassIds,
  open,
  onClose,
}: DispatchActionDialogProps & { open: boolean; onClose: () => void }) {
  const [state, formAction, pending] = useActionState(
    dispatchActionFormAction,
    EMPTY_STATE,
  );
  const formRef = useRef<HTMLFormElement>(null);
  useFirstInvalidField(state, formRef);
  const id = useId();
  const compatibleNodes = nodes.filter((node) =>
    compatibleNodeClassIds.has(node.node_class_id),
  );
  return (
    <Dialog
      open={open}
      onClose={onClose}
      title={`Dispatch ${action.name}`}
      variant="sheet"
      dismissible={!pending}
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
        <form
          ref={formRef}
          action={formAction}
          onReset={(event) => event.preventDefault()}
          className="space-y-4"
        >
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
              {compatibleNodes.map((node) => (
                <option key={node.id} value={node.id}>
                  {node.name} · {node.device_id}
                </option>
              ))}
            </select>
            <FieldError>{state.fieldErrors?.node_id}</FieldError>
            {compatibleNodes.length === 0 ? (
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
            <FieldError>{state.fieldErrors?.payload}</FieldError>
          </div>
          <div>
            <Label htmlFor={`${id}-executed`}>Execute at (optional)</Label>
            <Input
              id={`${id}-executed`}
              name="executed_at"
              type="datetime-local"
            />
            <FieldError>{state.fieldErrors?.executed_at}</FieldError>
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
            <Button
              type="submit"
              disabled={pending || compatibleNodes.length === 0}
            >
              {pending ? "Dispatching…" : "Dispatch"}
            </Button>
          </div>
        </form>
      )}
    </Dialog>
  );
}
