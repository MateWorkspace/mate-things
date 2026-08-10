"use client";

import {
  useActionState,
  useId,
  useRef,
  useState,
  type ChangeEvent,
} from "react";
import { Plus, Trash2 } from "lucide-react";

import ActionMessage from "@/components/forms/ActionMessage";
import FieldError from "@/components/forms/FieldError";
import Button from "@/components/ui/button";
import Dialog from "@/components/ui/dialog";
import Input from "@/components/ui/input";
import Label from "@/components/ui/label";
import type {
  FirmwareConfigSchemaItem,
  FirmwareResponse,
} from "@/lib/api/firmwares";
import { useFirstInvalidField } from "@/hooks/use-first-invalid-field";
import { useRefreshAfterAction } from "@/hooks/use-refresh-after-action";

import {
  createFirmwareAction,
  deleteFirmwareAction,
  replaceFirmwareBinaryAction,
  updateFirmwareAction,
  type FormActionState,
} from "../_lib/actions";

export interface FirmwareNodeClassOption {
  id: string;
  name: string;
}

interface FirmwareFormProps {
  canDelete?: boolean;
  canEdit?: boolean;
  canReplace?: boolean;
  configSchema?: readonly FirmwareConfigSchemaItem[];
  firmware?: FirmwareResponse;
  nodeClasses: readonly FirmwareNodeClassOption[];
}

interface SchemaRow extends FirmwareConfigSchemaItem {
  rowId: number;
}

const INITIAL_STATE: FormActionState = { status: "idle" };

function NodeClassField({
  defaultValue,
  error,
  fieldId,
  nodeClasses,
}: {
  defaultValue?: string;
  error?: string;
  fieldId: string;
  nodeClasses: readonly FirmwareNodeClassOption[];
}) {
  if (nodeClasses.length === 0 && !defaultValue) {
    return (
      <div>
        <Label htmlFor={`${fieldId}-node-class`}>Node class ID</Label>
        <Input
          id={`${fieldId}-node-class`}
          name="node_class_id"
          defaultValue={defaultValue}
          required
        />
        <FieldError>{error}</FieldError>
      </div>
    );
  }

  const currentClassMissing = Boolean(
    defaultValue &&
    !nodeClasses.some((nodeClass) => nodeClass.id === defaultValue),
  );

  return (
    <div>
      <Label htmlFor={`${fieldId}-node-class`}>Node class</Label>
      <select
        id={`${fieldId}-node-class`}
        name="node_class_id"
        defaultValue={defaultValue ?? ""}
        required
        className="border-control-border bg-background text-foreground focus-visible:border-focus focus-visible:ring-focus min-h-11 w-full rounded-xl border px-3.5 py-2.5 text-sm focus-visible:ring-2 focus-visible:outline-none"
      >
        <option value="">Select a node class</option>
        {currentClassMissing ? (
          <option value={defaultValue}>Current class ({defaultValue})</option>
        ) : null}
        {nodeClasses.map((nodeClass) => (
          <option key={nodeClass.id} value={nodeClass.id}>
            {nodeClass.name}
          </option>
        ))}
      </select>
      <FieldError>{error}</FieldError>
    </div>
  );
}

function ConfigSchemaFields({
  error,
  initialSchema,
  serializedName,
}: {
  error?: string;
  initialSchema: readonly FirmwareConfigSchemaItem[];
  serializedName?: string;
}) {
  const id = useId();
  const [nextId, setNextId] = useState(initialSchema.length + 1);
  const [rows, setRows] = useState<SchemaRow[]>(() =>
    (initialSchema.length > 0
      ? initialSchema
      : [{ key: "", value_type: "" }]
    ).map((item, index) => ({ ...item, rowId: index + 1 })),
  );

  const updateRow = (
    rowId: number,
    field: keyof FirmwareConfigSchemaItem,
    event: ChangeEvent<HTMLInputElement | HTMLSelectElement>,
  ) => {
    setRows((current) =>
      current.map((row) =>
        row.rowId === rowId ? { ...row, [field]: event.target.value } : row,
      ),
    );
  };

  return (
    <fieldset
      data-field-name="config_schema"
      tabIndex={-1}
      className="border-border space-y-3 rounded-xl border p-4"
    >
      <legend className="px-1 text-sm font-semibold">
        Configuration schema
      </legend>
      <p className="text-muted-foreground text-sm">
        Define keys the node can configure. Supported types are string, uint32,
        and bool.
      </p>
      {serializedName ? (
        <input
          type="hidden"
          name={serializedName}
          value={JSON.stringify(
            rows.map(({ key, value_type }) => ({ key, value_type })),
          )}
        />
      ) : null}

      {rows.map((row, index) => (
        <div
          key={row.rowId}
          className="border-border grid gap-3 border-t pt-3 first:border-t-0 first:pt-0 sm:grid-cols-[1fr_10rem_auto]"
        >
          <div>
            <Label htmlFor={`${id}-key-${row.rowId}`}>
              Configuration key {index + 1}
            </Label>
            <Input
              id={`${id}-key-${row.rowId}`}
              name="schema_key"
              value={row.key}
              onChange={(event) => updateRow(row.rowId, "key", event)}
            />
          </div>
          <div>
            <Label htmlFor={`${id}-type-${row.rowId}`}>
              Value type {index + 1}
            </Label>
            <select
              id={`${id}-type-${row.rowId}`}
              name="schema_value_type"
              value={row.value_type}
              className="border-control-border bg-background text-foreground focus-visible:border-focus focus-visible:ring-focus min-h-11 w-full rounded-xl border px-3.5 py-2.5 text-sm focus-visible:ring-2 focus-visible:outline-none"
              onChange={(event) => updateRow(row.rowId, "value_type", event)}
            >
              <option value="">Select a type</option>
              <option value="string">string</option>
              <option value="uint32">uint32</option>
              <option value="bool">bool</option>
            </select>
          </div>
          <button
            type="button"
            aria-label={`Remove configuration parameter ${index + 1}`}
            disabled={rows.length === 1}
            className="border-border text-primary focus-visible:ring-focus mt-6 inline-flex min-h-11 min-w-11 items-center justify-center rounded-xl border disabled:cursor-not-allowed disabled:opacity-40"
            onClick={() =>
              setRows((current) =>
                current.filter((item) => item.rowId !== row.rowId),
              )
            }
          >
            <Trash2 aria-hidden="true" className="size-4" />
          </button>
        </div>
      ))}

      <Button
        type="button"
        variant="secondary"
        className="gap-2"
        onClick={() => {
          setRows((current) => [
            ...current,
            { key: "", value_type: "", rowId: nextId },
          ]);
          setNextId((current) => current + 1);
        }}
      >
        <Plus aria-hidden="true" className="size-4" />
        Add configuration parameter
      </Button>
      <FieldError>{error}</FieldError>
    </fieldset>
  );
}

type SchemaIntentMode = "keep" | "replace" | "clear";

function SchemaIntentFields({
  error,
  initialSchema,
  intentError,
}: {
  error?: string;
  initialSchema: readonly FirmwareConfigSchemaItem[];
  intentError?: string;
}) {
  const id = useId();
  const [mode, setMode] = useState<SchemaIntentMode>("keep");

  return (
    <fieldset
      data-field-name="schema_intent"
      tabIndex={-1}
      className="border-border space-y-3 rounded-xl border p-4"
    >
      <legend className="px-1 text-sm font-semibold">
        Configuration schema handling
      </legend>
      <div className="grid gap-2 sm:grid-cols-3">
        {(["keep", "replace", "clear"] as const).map((value) => (
          <label
            key={value}
            htmlFor={`${id}-${value}`}
            className="border-border focus-within:ring-focus flex min-h-11 cursor-pointer items-center gap-2 rounded-xl border px-3 py-2 text-sm font-medium focus-within:ring-2"
          >
            <input
              id={`${id}-${value}`}
              type="radio"
              name="schema_intent"
              value={value}
              // Keep the chosen mode as the reset baseline for React actions.
              defaultChecked={mode === value}
              onChange={() => setMode(value)}
            />
            {value[0].toUpperCase() + value.slice(1)}
          </label>
        ))}
      </div>
      <FieldError>{intentError}</FieldError>
      {mode === "keep" ? (
        <p className="text-muted-foreground text-sm">
          The existing configuration schema will remain unchanged.
        </p>
      ) : null}
      {mode === "replace" ? (
        <ConfigSchemaFields
          error={error}
          initialSchema={initialSchema}
          serializedName="config_schema"
        />
      ) : null}
      {mode === "clear" ? (
        <p className="text-muted-foreground text-sm">
          Clear will remove the existing configuration schema after the binary
          is replaced.
        </p>
      ) : null}
      {mode !== "replace" ? <FieldError>{error}</FieldError> : null}
    </fieldset>
  );
}

function UploadDialog({
  nodeClasses,
  open,
  onClose,
}: {
  nodeClasses: readonly FirmwareNodeClassOption[];
  open: boolean;
  onClose: () => void;
}) {
  const [state, formAction, isPending] = useActionState(
    createFirmwareAction,
    INITIAL_STATE,
  );
  const formRef = useRef<HTMLFormElement>(null);
  useRefreshAfterAction(state);
  useFirstInvalidField(state, formRef);
  const id = useId();
  const close = () => {
    if (!isPending) onClose();
  };

  return (
    <Dialog
      open={open && state.status !== "success"}
      onClose={close}
      title="Upload firmware"
      variant="sheet"
      dismissible={!isPending}
    >
      <form
        ref={formRef}
        action={formAction}
        onReset={(event) => event.preventDefault()}
        className="space-y-4"
      >
        <div>
          <Label htmlFor={`${id}-name`}>Firmware name</Label>
          <Input id={`${id}-name`} name="name" required />
          <FieldError>{state.fieldErrors?.name}</FieldError>
        </div>
        <NodeClassField
          error={state.fieldErrors?.node_class_id}
          fieldId={id}
          nodeClasses={nodeClasses}
        />
        <div>
          <Label htmlFor={`${id}-file`}>Firmware binary</Label>
          <Input
            id={`${id}-file`}
            name="file"
            type="file"
            accept=".bin,application/octet-stream"
            required
          />
          <FieldError>{state.fieldErrors?.file}</FieldError>
        </div>
        <ConfigSchemaFields
          error={state.fieldErrors?.config_schema}
          initialSchema={[]}
        />
        <ActionMessage state={state} />
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
            {isPending ? "Uploading…" : "Upload firmware"}
          </Button>
        </div>
      </form>
    </Dialog>
  );
}

function EditDialog({
  firmware,
  nodeClasses,
  open,
  onClose,
}: {
  firmware: FirmwareResponse;
  nodeClasses: readonly FirmwareNodeClassOption[];
  open: boolean;
  onClose: () => void;
}) {
  const [state, formAction, isPending] = useActionState(
    updateFirmwareAction,
    INITIAL_STATE,
  );
  const formRef = useRef<HTMLFormElement>(null);
  useRefreshAfterAction(state);
  useFirstInvalidField(state, formRef);
  const id = useId();
  const close = () => {
    if (!isPending) onClose();
  };

  return (
    <Dialog
      open={open && state.status !== "success"}
      onClose={close}
      title={`Edit ${firmware.name}`}
      variant="sheet"
      dismissible={!isPending}
    >
      <form
        ref={formRef}
        action={formAction}
        onReset={(event) => event.preventDefault()}
        className="space-y-4"
      >
        <input type="hidden" name="firmware_id" value={firmware.id} />
        <div>
          <Label htmlFor={`${id}-name`}>Firmware name</Label>
          <Input
            id={`${id}-name`}
            name="name"
            defaultValue={firmware.name}
            required
          />
          <FieldError>{state.fieldErrors?.name}</FieldError>
        </div>
        <NodeClassField
          defaultValue={firmware.node_class_id}
          error={state.fieldErrors?.node_class_id}
          fieldId={id}
          nodeClasses={nodeClasses}
        />
        <ActionMessage state={state} />
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
            {isPending ? "Saving…" : "Save firmware"}
          </Button>
        </div>
      </form>
    </Dialog>
  );
}

function ReplaceDialog({
  configSchema,
  firmware,
  open,
  onClose,
}: {
  configSchema: readonly FirmwareConfigSchemaItem[];
  firmware: FirmwareResponse;
  open: boolean;
  onClose: () => void;
}) {
  const [state, formAction, isPending] = useActionState(
    replaceFirmwareBinaryAction,
    INITIAL_STATE,
  );
  const formRef = useRef<HTMLFormElement>(null);
  useRefreshAfterAction(state);
  useFirstInvalidField(state, formRef);
  const id = useId();
  const close = () => {
    if (!isPending) onClose();
  };

  return (
    <Dialog
      open={open && state.status !== "success"}
      onClose={close}
      title={`Replace ${firmware.name} binary`}
      variant="sheet"
      dismissible={!isPending}
    >
      <form
        ref={formRef}
        action={formAction}
        onReset={(event) => event.preventDefault()}
        className="space-y-4"
      >
        <input type="hidden" name="firmware_id" value={firmware.id} />
        <p className="border-warning/40 bg-warning/10 rounded-xl border p-4 text-sm">
          Choose separately whether to keep, replace, or clear the configuration
          schema. Existing node values remain governed by backend compatibility
          rules.
        </p>
        <div>
          <Label htmlFor={`${id}-file`}>Replacement firmware binary</Label>
          <Input
            id={`${id}-file`}
            name="file"
            type="file"
            accept=".bin,application/octet-stream"
            required
          />
          <FieldError>{state.fieldErrors?.file}</FieldError>
        </div>
        <SchemaIntentFields
          error={state.fieldErrors?.config_schema}
          initialSchema={configSchema}
          intentError={state.fieldErrors?.schema_intent}
        />
        <ActionMessage state={state} />
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
            {isPending ? "Replacing…" : "Replace binary"}
          </Button>
        </div>
      </form>
    </Dialog>
  );
}

function DeleteDialog({
  firmware,
  open,
  onClose,
}: {
  firmware: FirmwareResponse;
  open: boolean;
  onClose: () => void;
}) {
  const [confirmation, setConfirmation] = useState("");
  const [state, formAction, isPending] = useActionState(
    deleteFirmwareAction,
    INITIAL_STATE,
  );
  const formRef = useRef<HTMLFormElement>(null);
  useFirstInvalidField(state, formRef);
  const id = useId();
  const close = () => {
    if (!isPending) onClose();
  };

  return (
    <Dialog
      open={open}
      onClose={close}
      title={`Delete ${firmware.name}`}
      variant="sheet"
      dismissible={!isPending}
    >
      <form
        ref={formRef}
        action={formAction}
        onReset={(event) => event.preventDefault()}
        className="space-y-4"
      >
        <input type="hidden" name="firmware_id" value={firmware.id} />
        <p className="border-critical/40 bg-critical/5 rounded-xl border p-4 text-sm">
          Nodes or other dependent resources can prevent deletion. Enter the
          exact firmware name <strong>{firmware.name}</strong> to continue.
        </p>
        <div>
          <Label htmlFor={`${id}-confirmation`}>Confirm firmware name</Label>
          <Input
            id={`${id}-confirmation`}
            name="confirmation"
            value={confirmation}
            required
            autoComplete="off"
            onChange={(event) => setConfirmation(event.target.value)}
          />
          <FieldError>{state.fieldErrors?.confirmation}</FieldError>
        </div>
        <ActionMessage state={state} />
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
            disabled={isPending || confirmation !== firmware.name}
            className="bg-critical text-background focus-visible:ring-critical focus-visible:ring-offset-background inline-flex min-h-11 items-center justify-center rounded-xl px-4 py-2.5 text-sm font-semibold transition-opacity hover:opacity-90 focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-50"
          >
            {isPending ? "Deleting…" : "Permanently delete firmware"}
          </button>
        </div>
      </form>
    </Dialog>
  );
}

export default function FirmwareForm({
  canDelete = false,
  canEdit = false,
  canReplace = false,
  configSchema = [],
  firmware,
  nodeClasses,
}: FirmwareFormProps) {
  const [dialog, setDialog] = useState<
    "upload" | "edit" | "replace" | "delete" | null
  >(null);
  const [instance, setInstance] = useState(0);
  const open = (next: Exclude<typeof dialog, null>) => {
    setInstance((current) => current + 1);
    setDialog(next);
  };
  const close = () => setDialog(null);

  if (!firmware) {
    return (
      <>
        <Button type="button" onClick={() => open("upload")}>
          Upload firmware
        </Button>
        <UploadDialog
          key={instance}
          nodeClasses={nodeClasses}
          open={dialog === "upload"}
          onClose={close}
        />
      </>
    );
  }

  return (
    <div className="flex flex-wrap justify-end gap-2">
      {canEdit ? (
        <Button
          type="button"
          variant="secondary"
          aria-label={`Edit ${firmware.name}`}
          onClick={() => open("edit")}
        >
          Edit firmware
        </Button>
      ) : null}
      {canReplace ? (
        <Button
          type="button"
          variant="secondary"
          aria-label={`Replace ${firmware.name} binary`}
          onClick={() => open("replace")}
        >
          Replace binary
        </Button>
      ) : null}
      {canDelete ? (
        <button
          type="button"
          aria-label={`Delete ${firmware.name}`}
          className="border-critical text-critical hover:bg-critical/10 focus-visible:ring-critical focus-visible:ring-offset-background inline-flex min-h-11 items-center justify-center rounded-xl border px-4 py-2.5 text-sm font-semibold focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none"
          onClick={() => open("delete")}
        >
          Delete firmware
        </button>
      ) : null}

      {canEdit ? (
        <EditDialog
          key={`edit-${instance}`}
          firmware={firmware}
          nodeClasses={nodeClasses}
          open={dialog === "edit"}
          onClose={close}
        />
      ) : null}
      {canReplace ? (
        <ReplaceDialog
          key={`replace-${instance}`}
          configSchema={configSchema}
          firmware={firmware}
          open={dialog === "replace"}
          onClose={close}
        />
      ) : null}
      {canDelete ? (
        <DeleteDialog
          key={`delete-${instance}`}
          firmware={firmware}
          open={dialog === "delete"}
          onClose={close}
        />
      ) : null}
    </div>
  );
}
