"use client";

import { useActionState, useState } from "react";

import JsonEditor from "@/components/json/JsonEditor";
import Button from "@/components/ui/button";
import Dialog from "@/components/ui/dialog";
import Input from "@/components/ui/input";
import Label from "@/components/ui/label";
import type { PayloadSchemaResponse } from "@/lib/api/payload-schemas";

import {
  createPayloadSchemaAction,
  deletePayloadSchemaAction,
  EMPTY_SCHEMA_STATE,
  updatePayloadSchemaAction,
} from "../_lib/actions";

function localDate(value?: string): string {
  if (!value) return "";
  const date = new Date(value);
  const shifted = new Date(date.valueOf() - date.getTimezoneOffset() * 60000);
  return shifted.toISOString().slice(0, 16);
}

export default function PayloadSchemaForm({
  schema,
  canEdit = true,
  canDelete = false,
}: {
  schema?: PayloadSchemaResponse;
  canEdit?: boolean;
  canDelete?: boolean;
}) {
  const [editorOpen, setEditorOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);
  return (
    <div className="flex flex-wrap gap-2">
      {!schema || canEdit ? (
        <Button
          type="button"
          variant={schema ? "secondary" : "primary"}
          onClick={() => setEditorOpen(true)}
        >
          {schema ? "Edit schema" : "Create schema"}
        </Button>
      ) : null}
      {schema && canDelete ? (
        <button
          type="button"
          onClick={() => setDeleteOpen(true)}
          className="border-critical text-critical rounded-xl border px-4 py-2.5 text-sm font-semibold"
        >
          Delete schema
        </button>
      ) : null}
      <SchemaEditor
        key={`${editorOpen}-${schema?.id ?? "new"}`}
        schema={schema}
        open={editorOpen}
        onClose={() => setEditorOpen(false)}
      />
      {schema ? (
        <DeleteSchema
          key={`${deleteOpen}-${schema.id}`}
          schema={schema}
          open={deleteOpen}
          onClose={() => setDeleteOpen(false)}
        />
      ) : null}
    </div>
  );
}

function SchemaEditor({
  schema,
  open,
  onClose,
}: {
  schema?: PayloadSchemaResponse;
  open: boolean;
  onClose: () => void;
}) {
  const [state, action, pending] = useActionState(
    schema ? updatePayloadSchemaAction : createPayloadSchemaAction,
    EMPTY_SCHEMA_STATE,
  );
  return (
    <Dialog
      open={open && state.status !== "success"}
      onClose={onClose}
      title={
        schema
          ? `Edit ${schema.name} v${schema.version}`
          : "Create payload schema"
      }
      variant="sheet"
    >
      <form action={action} className="space-y-4">
        {schema ? (
          <input type="hidden" name="payload_schema_id" value={schema.id} />
        ) : null}
        <div>
          <Label htmlFor="schema-name">Name</Label>
          <Input
            id="schema-name"
            name="name"
            defaultValue={schema?.name}
            required
          />
          {state.fieldErrors?.name ? (
            <p className="text-critical text-sm">{state.fieldErrors.name}</p>
          ) : null}
        </div>
        <div>
          <Label htmlFor="schema-version">Version</Label>
          <Input
            id="schema-version"
            name="version"
            type="number"
            min={1}
            step={1}
            defaultValue={schema?.version ?? 1}
            required
          />
          {state.fieldErrors?.version ? (
            <p className="text-critical text-sm">{state.fieldErrors.version}</p>
          ) : null}
        </div>
        <JsonEditor
          name="definition"
          label="Definition"
          defaultValue={
            schema?.definition ?? { type: "object", properties: {} }
          }
        />
        <div className="grid gap-4 sm:grid-cols-2">
          <div>
            <Label htmlFor="schema-valid-from">Valid from</Label>
            <Input
              id="schema-valid-from"
              name="valid_from"
              type="datetime-local"
              defaultValue={localDate(schema?.valid_from)}
            />
          </div>
          <div>
            <Label htmlFor="schema-valid-to">Valid to</Label>
            <Input
              id="schema-valid-to"
              name="valid_to"
              type="datetime-local"
              defaultValue={localDate(schema?.valid_to)}
            />
            {state.fieldErrors?.valid_to ? (
              <p className="text-critical text-sm">
                {state.fieldErrors.valid_to}
              </p>
            ) : null}
          </div>
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
          <Button type="button" variant="secondary" onClick={onClose}>
            Cancel
          </Button>
          <Button type="submit" disabled={pending}>
            {pending ? "Saving…" : "Save schema"}
          </Button>
        </div>
      </form>
    </Dialog>
  );
}

function DeleteSchema({
  schema,
  open,
  onClose,
}: {
  schema: PayloadSchemaResponse;
  open: boolean;
  onClose: () => void;
}) {
  const [state, action, pending] = useActionState(
    deletePayloadSchemaAction,
    EMPTY_SCHEMA_STATE,
  );
  const identity = `${schema.name} v${schema.version}`;
  return (
    <Dialog
      open={open}
      onClose={onClose}
      title={`Delete ${identity}`}
      variant="sheet"
    >
      <form action={action} className="space-y-4">
        <input type="hidden" name="payload_schema_id" value={schema.id} />
        <input type="hidden" name="identity" value={identity} />
        <p className="bg-muted rounded-xl p-4 text-sm">
          Actions may reference this exact version, which can cause the backend
          to reject deletion. Enter <strong>{identity}</strong> to continue.
        </p>
        <Input
          name="confirmation"
          aria-label="Confirm schema identity"
          required
        />
        <p
          className={
            state.status === "error"
              ? "text-critical text-sm"
              : "text-success text-sm"
          }
        >
          {state.message}
        </p>
        <div className="flex justify-end gap-2">
          <Button type="button" variant="secondary" onClick={onClose}>
            Cancel
          </Button>
          <Button type="submit" disabled={pending}>
            {pending ? "Deleting…" : "Delete schema"}
          </Button>
        </div>
      </form>
    </Dialog>
  );
}
