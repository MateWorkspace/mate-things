"use client";

import {
  useActionState,
  useContext,
  useEffect,
  useId,
  useRef,
  useState,
} from "react";

import Button from "@/components/ui/button";
import Dialog from "@/components/ui/dialog";
import Input from "@/components/ui/input";
import Label from "@/components/ui/label";
import { ToastContext } from "@/components/ui/toast-provider";
import type { NodeClassResponse } from "@/lib/api/node-classes";

import {
  createNodeClassAction,
  deleteNodeClassAction,
  updateNodeClassAction,
  type FormActionState,
} from "../_lib/actions";

const INITIAL_STATE: FormActionState = { status: "idle" };

interface NodeClassFormProps {
  canDelete?: boolean;
  canEdit?: boolean;
  nodeClass?: NodeClassResponse;
}

function useActionToast(state: FormActionState): void {
  const toast = useContext(ToastContext);
  const lastShown = useRef<FormActionState | null>(null);

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
}

export default function NodeClassForm({
  canDelete = false,
  canEdit = true,
  nodeClass,
}: NodeClassFormProps) {
  const [editorOpen, setEditorOpen] = useState(false);
  const [editorInstance, setEditorInstance] = useState(0);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [deleteInstance, setDeleteInstance] = useState(0);
  const showEditor = !nodeClass || canEdit;

  return (
    <div className="flex flex-wrap justify-end gap-2">
      {showEditor ? (
        <Button
          aria-label={
            nodeClass ? `Edit ${nodeClass.name}` : "Create node class"
          }
          type="button"
          variant={nodeClass ? "secondary" : "primary"}
          onClick={() => {
            setEditorInstance((current) => current + 1);
            setEditorOpen(true);
          }}
        >
          {nodeClass ? "Edit class" : "Create node class"}
        </Button>
      ) : null}

      {nodeClass && canDelete ? (
        <button
          aria-label={`Delete ${nodeClass.name}`}
          type="button"
          className="border-critical text-critical hover:bg-critical/10 focus-visible:ring-critical focus-visible:ring-offset-background inline-flex items-center justify-center rounded-xl border px-4 py-2.5 text-sm font-semibold transition-colors focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none"
          onClick={() => {
            setDeleteInstance((current) => current + 1);
            setDeleteOpen(true);
          }}
        >
          Delete class
        </button>
      ) : null}

      {showEditor ? (
        <NodeClassEditorDialog
          key={`editor-${editorInstance}`}
          nodeClass={nodeClass}
          open={editorOpen}
          onClose={() => setEditorOpen(false)}
        />
      ) : null}

      {nodeClass && canDelete ? (
        <DeleteNodeClassDialog
          key={`delete-${deleteInstance}`}
          nodeClass={nodeClass}
          open={deleteOpen}
          onClose={() => setDeleteOpen(false)}
        />
      ) : null}
    </div>
  );
}

function NodeClassEditorDialog({
  nodeClass,
  open,
  onClose,
}: {
  nodeClass?: NodeClassResponse;
  open: boolean;
  onClose: () => void;
}) {
  const action = nodeClass ? updateNodeClassAction : createNodeClassAction;
  const [state, formAction, isPending] = useActionState(action, INITIAL_STATE);
  const fieldId = useId();
  useActionToast(state);
  const close = () => {
    if (!isPending) {
      onClose();
    }
  };

  return (
    <Dialog
      open={open && state.status !== "success"}
      onClose={close}
      title={nodeClass ? `Edit ${nodeClass.name}` : "Create node class"}
      variant="sheet"
    >
      <form action={formAction} className="space-y-4">
        {nodeClass ? (
          <input type="hidden" name="node_class_id" value={nodeClass.id} />
        ) : null}

        <div>
          <Label htmlFor={`${fieldId}-name`}>Name</Label>
          <Input
            id={`${fieldId}-name`}
            name="name"
            defaultValue={nodeClass?.name}
            required
            aria-invalid={Boolean(state.fieldErrors?.name)}
            aria-describedby={
              state.fieldErrors?.name ? `${fieldId}-name-error` : undefined
            }
          />
          {state.fieldErrors?.name ? (
            <p
              id={`${fieldId}-name-error`}
              className="text-critical mt-1.5 text-sm"
            >
              {state.fieldErrors.name}
            </p>
          ) : null}
        </div>

        <div>
          <Label htmlFor={`${fieldId}-description`}>Description</Label>
          <textarea
            id={`${fieldId}-description`}
            name="description"
            defaultValue={nodeClass?.description}
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
            disabled={isPending}
            onClick={close}
          >
            Cancel
          </Button>
          <Button type="submit" disabled={isPending}>
            {isPending ? "Saving…" : nodeClass ? "Save class" : "Create class"}
          </Button>
        </div>
      </form>
    </Dialog>
  );
}

function DeleteNodeClassDialog({
  nodeClass,
  open,
  onClose,
}: {
  nodeClass: NodeClassResponse;
  open: boolean;
  onClose: () => void;
}) {
  const [confirmation, setConfirmation] = useState("");
  const [state, formAction, isPending] = useActionState(
    deleteNodeClassAction,
    INITIAL_STATE,
  );
  const fieldId = useId();
  const close = () => {
    if (!isPending) {
      onClose();
    }
  };

  return (
    <Dialog
      open={open}
      onClose={close}
      title={`Delete ${nodeClass.name}`}
      variant="sheet"
    >
      <form action={formAction} className="space-y-4">
        <input type="hidden" name="node_class_id" value={nodeClass.id} />
        <input type="hidden" name="node_class_name" value={nodeClass.name} />

        <div className="border-critical/40 bg-critical/5 rounded-xl border p-4">
          <p className="text-sm">
            Dependent nodes, firmware, or actions can prevent this class from
            being removed. The backend may reject this deletion to preserve
            those resources.
          </p>
          <p className="mt-2 text-sm">
            Enter the exact class name{" "}
            <strong className="font-semibold">{nodeClass.name}</strong> to
            continue.
          </p>
        </div>

        <div>
          <Label htmlFor={`${fieldId}-confirmation`}>Confirm class name</Label>
          <Input
            id={`${fieldId}-confirmation`}
            name="confirmation"
            value={confirmation}
            placeholder={nodeClass.name}
            autoComplete="off"
            required
            aria-invalid={Boolean(state.fieldErrors?.confirmation)}
            aria-describedby={
              state.fieldErrors?.confirmation
                ? `${fieldId}-confirmation-error`
                : undefined
            }
            onChange={(event) => setConfirmation(event.target.value)}
          />
          {state.fieldErrors?.confirmation ? (
            <p
              id={`${fieldId}-confirmation-error`}
              className="text-critical mt-1.5 text-sm"
            >
              {state.fieldErrors.confirmation}
            </p>
          ) : null}
        </div>

        <p aria-live="assertive" className="text-critical text-sm">
          {state.message}
        </p>

        <div className="flex flex-wrap justify-end gap-2">
          <Button
            type="button"
            variant="secondary"
            disabled={isPending}
            onClick={close}
          >
            Cancel
          </Button>
          <button
            type="submit"
            disabled={isPending || confirmation !== nodeClass.name}
            className="bg-critical text-background focus-visible:ring-critical focus-visible:ring-offset-background inline-flex items-center justify-center rounded-xl px-4 py-2.5 text-sm font-semibold transition-opacity hover:opacity-90 focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-50"
          >
            {isPending ? "Deleting…" : "Permanently delete class"}
          </button>
        </div>
      </form>
    </Dialog>
  );
}
