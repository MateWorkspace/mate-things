"use client";

import { useActionState, useEffect, useMemo, useRef, useState } from "react";

import ActionMessage from "@/components/forms/ActionMessage";
import FieldError from "@/components/forms/FieldError";
import Button from "@/components/ui/button";
import Dialog from "@/components/ui/dialog";
import Input from "@/components/ui/input";
import Label from "@/components/ui/label";
import type { PayloadSchemaResponse } from "@/lib/api/payload-schemas";
import { useFirstInvalidField } from "@/hooks/use-first-invalid-field";
import { useRefreshAfterAction } from "@/hooks/use-refresh-after-action";

import {
  createPayloadSchemaAction,
  deletePayloadSchemaAction,
  updatePayloadSchemaAction,
} from "../_lib/actions";
import { EMPTY_SCHEMA_STATE } from "../_lib/state";
import { parseDefinitionFieldError } from "./schema-definition-errors";
import SchemaDefinitionBuilder from "./SchemaDefinitionBuilder";

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
  const [editorGeneration, setEditorGeneration] = useState(0);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [deleteGeneration, setDeleteGeneration] = useState(0);
  return (
    <div className="flex flex-wrap gap-2">
      {!schema || canEdit ? (
        <Button
          type="button"
          variant={schema ? "secondary" : "primary"}
          onClick={() => {
            setEditorGeneration((generation) => generation + 1);
            setEditorOpen(true);
          }}
        >
          {schema ? "Edit schema" : "Create schema"}
        </Button>
      ) : null}
      {schema && canDelete ? (
        <button
          type="button"
          onClick={() => {
            setDeleteGeneration((generation) => generation + 1);
            setDeleteOpen(true);
          }}
          className="border-critical text-critical hover:bg-critical/10 active:bg-critical/15 rounded-xl border px-4 py-2.5 text-sm font-semibold transition-colors"
        >
          Delete schema
        </button>
      ) : null}
      {/* Keyed on a generation counter bumped only when opening, not on
          editorOpen/deleteOpen themselves - remounting on close would
          unmount a still-open native <dialog>, dropping it from the
          browser's top layer without a clean close() call. */}
      <SchemaEditor
        key={`editor-${editorGeneration}-${schema?.id ?? "new"}`}
        schema={schema}
        open={editorOpen}
        onClose={() => setEditorOpen(false)}
      />
      {schema ? (
        <DeleteSchema
          key={`delete-${deleteGeneration}-${schema.id}`}
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
  const submitAction = schema
    ? updatePayloadSchemaAction.bind(null, schema.id)
    : createPayloadSchemaAction;
  const [state, action, pending] = useActionState(
    submitAction,
    EMPTY_SCHEMA_STATE,
  );
  const formRef = useRef<HTMLFormElement>(null);
  useRefreshAfterAction(state);
  useFirstInvalidField(state, formRef);
  useEffect(() => {
    if (state.status === "success") {
      onClose();
    }
  }, [state, onClose]);
  const definitionErrors = useMemo(() => {
    if (state.status !== "error" || !state.message) return {};
    const parsed = parseDefinitionFieldError(state.message);
    return parsed ? { [parsed.path.join(".")]: parsed.message } : {};
  }, [state]);
  const [definitionRepresentable, setDefinitionRepresentable] = useState(true);
  const [definitionValid, setDefinitionValid] = useState(true);
  return (
    <Dialog
      open={open}
      onClose={onClose}
      title={
        schema
          ? `Edit ${schema.name} v${schema.version}`
          : "Create payload schema"
      }
      variant="sheet"
      dismissible={!pending}
    >
      <form
        ref={formRef}
        action={action}
        onReset={(event) => event.preventDefault()}
        className="space-y-4"
      >
        {schema ? (
          <input type="hidden" name="payload_schema_id" value={schema.id} />
        ) : null}
        <fieldset disabled={!definitionRepresentable} className="space-y-4">
          <div>
            <Label htmlFor="schema-name">Name</Label>
            <Input
              id="schema-name"
              name="name"
              defaultValue={schema?.name}
              required
            />
            <FieldError>{state.fieldErrors?.name}</FieldError>
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
            <FieldError>{state.fieldErrors?.version}</FieldError>
          </div>
          <div data-field-name="definition" tabIndex={-1}>
            <h3 className="text-foreground/80 mb-1.5 block text-sm font-medium">
              Definition
            </h3>
            <SchemaDefinitionBuilder
              name="definition"
              defaultValue={schema?.definition}
              errors={definitionErrors}
              onRepresentableChange={setDefinitionRepresentable}
              onValidityChange={setDefinitionValid}
            />
            <FieldError>{state.fieldErrors?.definition}</FieldError>
          </div>
          <div className="grid gap-4 sm:grid-cols-2">
            <div>
              <Label htmlFor="schema-valid-from">Valid from</Label>
              <Input
                id="schema-valid-from"
                name="valid_from"
                type="datetime-local"
                defaultValue={localDate(schema?.valid_from)}
              />
              <FieldError>{state.fieldErrors?.valid_from}</FieldError>
            </div>
            <div>
              <Label htmlFor="schema-valid-to">Valid to</Label>
              <Input
                id="schema-valid-to"
                name="valid_to"
                type="datetime-local"
                defaultValue={localDate(schema?.valid_to)}
              />
              <FieldError>{state.fieldErrors?.valid_to}</FieldError>
            </div>
          </div>
        </fieldset>
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
            disabled={pending || !definitionRepresentable || !definitionValid}
          >
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
  const formRef = useRef<HTMLFormElement>(null);
  useFirstInvalidField(state, formRef);
  const identity = `${schema.name} v${schema.version}`;
  return (
    <Dialog
      open={open}
      onClose={onClose}
      title={`Delete ${identity}`}
      variant="sheet"
      dismissible={!pending}
    >
      <form
        ref={formRef}
        action={action}
        onReset={(event) => event.preventDefault()}
        className="space-y-4"
      >
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
        <FieldError>{state.fieldErrors?.confirmation}</FieldError>
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
          <Button type="submit" disabled={pending}>
            {pending ? "Deleting…" : "Delete schema"}
          </Button>
        </div>
      </form>
    </Dialog>
  );
}
