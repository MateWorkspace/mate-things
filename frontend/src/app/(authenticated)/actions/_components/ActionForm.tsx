"use client";

import { useRouter } from "next/navigation";
import { useActionState, useEffect, useId, useState } from "react";

import Button from "@/components/ui/button";
import Dialog from "@/components/ui/dialog";
import Input from "@/components/ui/input";
import Label from "@/components/ui/label";
import type { ActionResponse } from "@/lib/api/actions";
import type { PayloadSchemaResponse } from "@/lib/api/payload-schemas";

import {
  createActionFormAction,
  deleteActionFormAction,
  updateActionFormAction,
  type ActionFormState,
} from "../_lib/actions";

const EMPTY_STATE: ActionFormState = { status: "idle" };

interface ActionFormProps {
  action?: ActionResponse;
  schemas: readonly PayloadSchemaResponse[];
  canEdit?: boolean;
  canDelete?: boolean;
}

export default function ActionForm({
  action,
  schemas,
  canEdit = true,
  canDelete = false,
}: ActionFormProps) {
  const [editorOpen, setEditorOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);
  return (
    <div className="flex flex-wrap gap-2">
      {canEdit ? (
        <Button
          type="button"
          variant={action ? "secondary" : "primary"}
          onClick={() => setEditorOpen(true)}
        >
          {action ? "Edit action" : "Create action"}
        </Button>
      ) : null}
      {action && canDelete ? (
        <button
          type="button"
          className="border-critical text-critical hover:bg-critical/10 active:bg-critical/15 focus-visible:ring-critical rounded-xl border px-4 py-2.5 text-sm font-semibold transition-colors focus-visible:ring-2 focus-visible:outline-none"
          onClick={() => setDeleteOpen(true)}
        >
          Delete action
        </button>
      ) : null}
      {canEdit ? (
        <EditorDialog
          action={action}
          schemas={schemas}
          open={editorOpen}
          onClose={() => setEditorOpen(false)}
        />
      ) : null}
      {action && canDelete ? (
        <DeleteDialog
          action={action}
          open={deleteOpen}
          onClose={() => setDeleteOpen(false)}
        />
      ) : null}
    </div>
  );
}

function EditorDialog({
  action,
  schemas,
  open,
  onClose,
}: Omit<ActionFormProps, "canEdit" | "canDelete"> & {
  open: boolean;
  onClose: () => void;
}) {
  const [state, formAction, pending] = useActionState(
    action ? updateActionFormAction : createActionFormAction,
    EMPTY_STATE,
  );
  const router = useRouter();
  useEffect(() => {
    if (state.status === "success") {
      router.refresh();
    }
  }, [state, router]);
  const id = useId();
  return (
    <Dialog
      open={open && state.status !== "success"}
      onClose={onClose}
      title={action ? `Edit ${action.name}` : "Create action"}
      variant="sheet"
    >
      <form action={formAction} className="space-y-4">
        {action ? (
          <input type="hidden" name="action_id" value={action.id} />
        ) : null}
        <div>
          <Label htmlFor={`${id}-name`}>Name</Label>
          <Input
            id={`${id}-name`}
            name="name"
            required
            defaultValue={action?.name}
          />
        </div>
        <div>
          <Label htmlFor={`${id}-description`}>Description</Label>
          <textarea
            id={`${id}-description`}
            name="description"
            defaultValue={action?.description}
            rows={3}
            className="border-control-border bg-background focus-visible:ring-focus w-full rounded-xl border px-3.5 py-2.5 text-sm focus-visible:ring-2 focus-visible:outline-none"
          />
        </div>
        <div>
          <Label htmlFor={`${id}-schema`}>Payload schema/version</Label>
          <select
            id={`${id}-schema`}
            name="schema"
            defaultValue={
              action
                ? `${action.payload_schema_name}::${action.payload_schema_version}`
                : ""
            }
            required
            onChange={(event) => {
              const [name = "", version = ""] = event.target.value.split("::");
              const form = event.currentTarget.form;
              if (form) {
                (
                  form.elements.namedItem(
                    "payload_schema_name",
                  ) as HTMLInputElement
                ).value = name;
                (
                  form.elements.namedItem(
                    "payload_schema_version",
                  ) as HTMLInputElement
                ).value = version;
              }
            }}
            className="border-control-border bg-background focus-visible:ring-focus min-h-11 w-full rounded-xl border px-3.5 text-sm focus-visible:ring-2 focus-visible:outline-none"
          >
            <option value="">Select schema</option>
            {schemas.map((schema) => (
              <option
                key={schema.id}
                value={`${schema.name}::${schema.version}`}
              >
                {schema.name} · v{schema.version}
              </option>
            ))}
          </select>
          <input
            type="hidden"
            name="payload_schema_name"
            defaultValue={action?.payload_schema_name}
          />
          <input
            type="hidden"
            name="payload_schema_version"
            defaultValue={action?.payload_schema_version}
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
            {pending ? "Saving…" : "Save action"}
          </Button>
        </div>
      </form>
    </Dialog>
  );
}

function DeleteDialog({
  action,
  open,
  onClose,
}: {
  action: ActionResponse;
  open: boolean;
  onClose: () => void;
}) {
  const [confirmation, setConfirmation] = useState("");
  const [state, formAction, pending] = useActionState(
    deleteActionFormAction,
    EMPTY_STATE,
  );
  const id = useId();
  return (
    <Dialog
      open={open}
      onClose={onClose}
      title={`Delete ${action.name}`}
      variant="sheet"
    >
      <form action={formAction} className="space-y-4">
        <input type="hidden" name="action_id" value={action.id} />
        <input type="hidden" name="action_name" value={action.name} />
        <p className="text-sm">
          Enter <strong>{action.name}</strong> to permanently delete this
          definition.
        </p>
        <div>
          <Label htmlFor={`${id}-confirmation`}>Confirm action name</Label>
          <Input
            id={`${id}-confirmation`}
            name="confirmation"
            value={confirmation}
            onChange={(event) => setConfirmation(event.target.value)}
          />
        </div>
        <p aria-live="assertive" className="text-critical text-sm">
          {state.message}
        </p>
        <div className="flex justify-end gap-2">
          <Button type="button" variant="secondary" onClick={onClose}>
            Cancel
          </Button>
          <button
            type="submit"
            disabled={pending || confirmation !== action.name}
            className="bg-critical text-background hover:opacity-90 rounded-xl px-4 py-2.5 text-sm font-semibold transition-opacity disabled:cursor-not-allowed disabled:opacity-50"
          >
            {pending ? "Deleting…" : "Delete action"}
          </button>
        </div>
      </form>
    </Dialog>
  );
}
