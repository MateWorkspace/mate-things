"use client";

import { useMemo, useState } from "react";

import Button from "@/components/ui/button";
import Card from "@/components/ui/card";

import type { ConfigSchemaEntry, SettingsSnapshot } from "../../../_lib/ble/protocol";
import SettingsField, { type SettingsFieldValue } from "./SettingsField";

function initialValues(
  schema: readonly ConfigSchemaEntry[],
  snapshot: SettingsSnapshot,
): Record<string, SettingsFieldValue> {
  const values: Record<string, SettingsFieldValue> = {};
  for (const entry of schema) {
    if (entry.type === "bool") {
      values[entry.key] = snapshot[entry.key] === true;
    } else {
      const raw = snapshot[entry.key];
      values[entry.key] =
        typeof raw === "string" || typeof raw === "number" ? String(raw) : "";
    }
  }
  return values;
}

export default function SettingsBlock({
  schema,
  snapshot,
  restartRequired,
  onSave,
  onRestart,
}: {
  schema: readonly ConfigSchemaEntry[];
  snapshot: SettingsSnapshot;
  restartRequired: boolean;
  onSave: (partial: Record<string, unknown>) => Promise<void>;
  onRestart: () => Promise<void>;
}) {
  const [values, setValues] = useState<Record<string, SettingsFieldValue>>(
    () => initialValues(schema, snapshot),
  );
  const [touched, setTouched] = useState<Record<string, boolean>>({});
  const [saveError, setSaveError] = useState<string>();
  const [savePending, setSavePending] = useState(false);
  const [restartError, setRestartError] = useState<string>();
  const [restartPending, setRestartPending] = useState(false);

  // Reset the form whenever the device gives us a fresh snapshot (after a
  // successful save or a reconnect) - adjusting state during render per
  // https://react.dev/learn/you-might-not-need-an-effect#adjusting-some-state-when-a-prop-changes,
  // not in an effect, so this doesn't cause an extra render pass.
  const [syncedSnapshot, setSyncedSnapshot] = useState(snapshot);
  if (snapshot !== syncedSnapshot) {
    setSyncedSnapshot(snapshot);
    setValues(initialValues(schema, snapshot));
    setTouched({});
  }

  const fieldErrors = useMemo(() => {
    const errors: Record<string, string> = {};
    for (const entry of schema) {
      if (entry.type !== "uint32" || !touched[entry.key]) continue;
      const raw = values[entry.key];
      if (typeof raw !== "string" || raw.trim() === "") continue;
      if (!/^\d+$/.test(raw.trim())) {
        errors[entry.key] = "Enter a whole number.";
      }
    }
    return errors;
  }, [schema, touched, values]);

  const touchedKeys = Object.keys(touched).filter((key) => touched[key]);
  const hasErrors = Object.keys(fieldErrors).length > 0;

  function handleChange(entry: ConfigSchemaEntry, value: SettingsFieldValue) {
    setValues((current) => ({ ...current, [entry.key]: value }));
    setTouched((current) => ({ ...current, [entry.key]: true }));
  }

  async function handleSave() {
    setSaveError(undefined);
    setSavePending(true);
    try {
      const partial: Record<string, unknown> = {};
      for (const entry of schema) {
        if (!touched[entry.key]) continue;
        const raw = values[entry.key];
        if (entry.type === "bool") {
          partial[entry.key] = raw === true;
        } else if (entry.type === "uint32") {
          if (typeof raw === "string" && raw.trim() !== "") {
            partial[entry.key] = Number.parseInt(raw, 10);
          }
        } else if (entry.type === "string") {
          if (typeof raw === "string" && raw !== "") {
            partial[entry.key] = raw;
          }
        }
      }
      await onSave(partial);
    } catch (error) {
      setSaveError(
        error instanceof Error ? error.message : "Failed to save settings.",
      );
    } finally {
      setSavePending(false);
    }
  }

  async function handleRestart() {
    setRestartError(undefined);
    setRestartPending(true);
    try {
      await onRestart();
    } catch (error) {
      setRestartError(
        error instanceof Error ? error.message : "Failed to restart device.",
      );
    } finally {
      setRestartPending(false);
    }
  }

  return (
    <Card>
      <h2 className="font-display text-primary text-xl tracking-wide">
        Settings
      </h2>

      {restartRequired ? (
        <div className="border-warning bg-warning/10 mt-4 flex flex-wrap items-center justify-between gap-3 rounded-xl border p-3">
          <p className="text-sm font-semibold">
            Some changes require a restart to take effect.
          </p>
          <Button
            variant="secondary"
            disabled={restartPending}
            onClick={() => void handleRestart()}
          >
            {restartPending ? "Restarting…" : "Restart now"}
          </Button>
        </div>
      ) : null}
      {restartError ? (
        <p className="text-critical mt-2 text-sm">{restartError}</p>
      ) : null}

      <div className="divide-border mt-4 divide-y">
        {schema.map((entry) => {
          const secretSet =
            entry.type === "string" &&
            !(entry.key in snapshot) &&
            typeof snapshot[`${entry.key}_set`] === "boolean"
              ? (snapshot[`${entry.key}_set`] as boolean)
              : undefined;
          return (
            <SettingsField
              key={entry.key}
              entry={entry}
              value={values[entry.key] ?? (entry.type === "bool" ? false : "")}
              secretSet={secretSet}
              error={fieldErrors[entry.key]}
              onChange={(value) => handleChange(entry, value)}
            />
          );
        })}
      </div>

      {saveError ? (
        <p className="text-critical mt-2 text-sm">{saveError}</p>
      ) : null}
      <Button
        className="mt-4"
        disabled={touchedKeys.length === 0 || hasErrors || savePending}
        onClick={() => void handleSave()}
      >
        {savePending ? "Saving…" : "Save settings"}
      </Button>
    </Card>
  );
}
