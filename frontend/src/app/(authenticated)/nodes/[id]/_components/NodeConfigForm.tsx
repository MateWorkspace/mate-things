"use client";

import { useActionState, useId } from "react";

import Button from "@/components/ui/button";
import Input from "@/components/ui/input";
import Label from "@/components/ui/label";
import { EmptyState } from "@/components/ui/states";
import type {
  FirmwareConfigParameterResponse,
  NodeConfigValueResponse,
} from "@/lib/api";

import { saveNodeConfigAction, type NodeActionState } from "../_lib/actions";

const INITIAL_STATE: NodeActionState = { status: "idle" };
const UINT32_MAX = 4_294_967_295;

interface NodeConfigFormProps {
  nodeId: string;
  parameters: FirmwareConfigParameterResponse[];
  values: NodeConfigValueResponse[];
  canSet: boolean;
}

interface ConfigParameterFormProps {
  nodeId: string;
  parameter: FirmwareConfigParameterResponse;
  currentValue?: NodeConfigValueResponse;
  canSet: boolean;
}

function displayKey(key: string): string {
  return key.replaceAll("_", " ");
}

function ValueControl({
  canSet,
  currentValue,
  inputId,
  parameter,
}: Omit<ConfigParameterFormProps, "nodeId"> & { inputId: string }) {
  const commonProps = {
    id: inputId,
    name: "value",
    disabled: !canSet,
  };

  switch (parameter.value_type) {
    case "string":
      return (
        <Input
          {...commonProps}
          type="text"
          defaultValue={currentValue?.value ?? ""}
        />
      );
    case "uint32":
      return (
        <Input
          {...commonProps}
          type="number"
          required
          inputMode="numeric"
          min={0}
          max={UINT32_MAX}
          step={1}
          defaultValue={currentValue?.value ?? ""}
        />
      );
    case "bool":
      return (
        <select
          {...commonProps}
          required
          defaultValue={currentValue?.value ?? ""}
          className="border-control-border bg-background text-foreground focus-visible:border-focus focus-visible:ring-focus w-full rounded-xl border px-3.5 py-2.5 text-sm transition-colors focus-visible:ring-2 focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-70"
        >
          <option value="" disabled>
            Choose true or false
          </option>
          <option value="true">true</option>
          <option value="false">false</option>
        </select>
      );
    default:
      return (
        <Input
          id={inputId}
          value={currentValue?.value ?? ""}
          disabled
          readOnly
          aria-describedby={`${inputId}-unsupported`}
        />
      );
  }
}

function ConfigParameterForm({
  canSet,
  currentValue,
  nodeId,
  parameter,
}: ConfigParameterFormProps) {
  const [state, formAction, isPending] = useActionState(
    saveNodeConfigAction,
    INITIAL_STATE,
  );
  const inputId = useId();
  const isSupported = ["string", "uint32", "bool"].includes(
    parameter.value_type,
  );
  const editable = canSet && isSupported;

  return (
    <form
      action={formAction}
      className="border-border bg-surface rounded-2xl border p-4"
    >
      <input type="hidden" name="node_id" value={nodeId} />
      <input type="hidden" name="key" value={parameter.key} />
      <input type="hidden" name="value_type" value={parameter.value_type} />

      <div className="flex flex-wrap items-start justify-between gap-2">
        <div>
          <Label htmlFor={inputId} className="capitalize">
            {displayKey(parameter.key)}
          </Label>
          <p className="text-muted-foreground font-mono text-xs">
            {parameter.key} · {parameter.value_type}
          </p>
        </div>
        {!currentValue ? (
          <span className="text-muted-foreground text-xs">Not set</span>
        ) : currentValue.updated_at ? (
          <time
            className="text-muted-foreground text-xs"
            dateTime={currentValue.updated_at}
          >
            Updated{" "}
            {new Intl.DateTimeFormat("en", {
              dateStyle: "medium",
              timeStyle: "short",
              timeZone: "UTC",
            }).format(new Date(currentValue.updated_at))}
          </time>
        ) : (
          <span className="text-muted-foreground text-xs">
            Set — update time unavailable
          </span>
        )}
      </div>

      <div className="mt-3">
        <ValueControl
          canSet={editable}
          currentValue={currentValue}
          inputId={inputId}
          parameter={parameter}
        />
        {!isSupported ? (
          <p
            id={`${inputId}-unsupported`}
            className="text-critical mt-1.5 text-sm"
          >
            This firmware declares an unsupported configuration type.
          </p>
        ) : null}
      </div>

      <div className="mt-3 flex flex-wrap items-center justify-between gap-3">
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
        {editable ? (
          <Button type="submit" disabled={isPending}>
            {isPending ? "Saving…" : "Save value"}
          </Button>
        ) : null}
      </div>
    </form>
  );
}

export default function NodeConfigForm({
  canSet,
  nodeId,
  parameters,
  values,
}: NodeConfigFormProps) {
  const valuesByKey = new Map(values.map((value) => [value.key, value]));

  return (
    <section aria-labelledby="device-configuration-heading">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h2
            id="device-configuration-heading"
            className="font-display text-primary text-2xl tracking-wide"
          >
            Device configuration
          </h2>
          <p className="text-foreground/70 mt-1 text-sm">
            Each value is validated against the current firmware schema and sent
            independently.
          </p>
        </div>
        {!canSet ? (
          <p className="text-muted-foreground text-sm">
            Read only · node_config:set is required to change values
          </p>
        ) : null}
      </div>

      {parameters.length > 0 ? (
        <div className="mt-4 grid gap-4 lg:grid-cols-2">
          {parameters.map((parameter) => {
            const currentValue = valuesByKey.get(parameter.key);

            return (
              <ConfigParameterForm
                key={`${nodeId}:${parameter.key}:${currentValue?.updated_at ?? currentValue?.value ?? ""}`}
                canSet={canSet}
                currentValue={currentValue}
                nodeId={nodeId}
                parameter={parameter}
              />
            );
          })}
        </div>
      ) : (
        <div className="mt-4">
          <EmptyState
            title="No configuration parameters"
            description="The node's current firmware does not declare a configuration schema."
          />
        </div>
      )}
    </section>
  );
}
