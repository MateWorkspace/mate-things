"use client";

import { useActionState, useContext, useEffect, useRef, useState } from "react";

import Button from "@/components/ui/button";
import Dialog from "@/components/ui/dialog";
import Input from "@/components/ui/input";
import Label from "@/components/ui/label";
import { ToastContext } from "@/components/ui/toast-provider";
import type { NodeResponse } from "@/lib/api";

import { saveNodeAction, type NodeActionState } from "../_lib/actions";

const INITIAL_STATE: NodeActionState = { status: "idle" };

interface NodeEditFormProps {
  node: NodeResponse;
}

export default function NodeEditForm({ node }: NodeEditFormProps) {
  const [open, setOpen] = useState(false);
  const [dialogInstance, setDialogInstance] = useState(0);

  return (
    <>
      <Button
        type="button"
        variant="secondary"
        onClick={() => {
          setDialogInstance((current) => current + 1);
          setOpen(true);
        }}
      >
        Edit node
      </Button>

      {open ? (
        <NodeEditDialog
          key={dialogInstance}
          node={node}
          onClose={() => setOpen(false)}
        />
      ) : null}
    </>
  );
}

interface NodeEditDialogProps {
  node: NodeResponse;
  onClose: () => void;
}

function NodeEditDialog({ node, onClose }: NodeEditDialogProps) {
  const [state, formAction, isPending] = useActionState(
    saveNodeAction,
    INITIAL_STATE,
  );
  const toast = useContext(ToastContext);
  const lastShown = useRef<NodeActionState | null>(null);

  useEffect(() => {
    if (
      state.status === "idle" ||
      !state.title ||
      state === lastShown.current
    ) {
      return;
    }

    lastShown.current = state;
    const message = state.message ?? "";

    if (state.status === "success") {
      toast?.success(state.title, message);
    } else {
      toast?.error(state.title, message);
    }
  }, [state, toast]);

  return (
    <Dialog
      open={state.status !== "success"}
      onClose={onClose}
      title={`Edit ${node.name}`}
      variant="sheet"
    >
      <form action={formAction} className="space-y-4">
        <input type="hidden" name="node_id" value={node.id} />

        <div>
          <Label htmlFor="node-edit-name">Name</Label>
          <Input
            id="node-edit-name"
            name="name"
            defaultValue={node.name}
            required
            aria-invalid={Boolean(state.fieldErrors?.name)}
            aria-describedby={
              state.fieldErrors?.name ? "node-edit-name-error" : undefined
            }
          />
          {state.fieldErrors?.name ? (
            <p
              id="node-edit-name-error"
              className="text-critical mt-1.5 text-sm"
            >
              {state.fieldErrors.name}
            </p>
          ) : null}
        </div>

        <div>
          <p className="text-foreground/80 mb-1.5 text-sm font-medium">
            Device ID
          </p>
          <p className="border-control-border bg-muted text-muted-foreground rounded-xl border px-3.5 py-2.5 font-mono text-sm break-all">
            {node.device_id}
          </p>
          <p className="text-muted-foreground mt-1.5 text-xs">
            Device identity is assigned during registration and cannot be
            changed here.
          </p>
        </div>

        <div>
          <Label htmlFor="node-edit-description">Description</Label>
          <textarea
            id="node-edit-description"
            name="description"
            defaultValue={node.description}
            rows={4}
            className="border-control-border bg-background text-foreground placeholder:text-foreground/40 focus-visible:border-focus focus-visible:ring-focus w-full resize-y rounded-xl border px-3.5 py-2.5 text-sm transition-colors focus-visible:ring-2 focus-visible:outline-none"
          />
        </div>

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

        <div className="flex flex-wrap justify-end gap-2">
          <Button
            type="button"
            variant="secondary"
            onClick={onClose}
            disabled={isPending}
          >
            Cancel
          </Button>
          <Button type="submit" disabled={isPending}>
            {isPending ? "Saving…" : "Save node"}
          </Button>
        </div>
      </form>
    </Dialog>
  );
}
