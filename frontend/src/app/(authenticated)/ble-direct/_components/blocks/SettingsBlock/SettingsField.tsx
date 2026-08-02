import Input from "@/components/ui/input";
import Label from "@/components/ui/label";

import type { ConfigSchemaEntry } from "../../../_lib/ble/protocol";

export type SettingsFieldValue = string | boolean;

export default function SettingsField({
  entry,
  value,
  secretSet,
  error,
  onChange,
}: {
  entry: ConfigSchemaEntry;
  value: SettingsFieldValue;
  secretSet: boolean | undefined;
  error: string | undefined;
  onChange: (value: SettingsFieldValue) => void;
}) {
  const fieldId = `ble-settings-${entry.key}`;

  if (entry.type === "bool") {
    return (
      <div className="flex items-center justify-between gap-4 py-2">
        <Label htmlFor={fieldId}>{entry.key}</Label>
        <input
          id={fieldId}
          type="checkbox"
          checked={value === true}
          onChange={(event) => onChange(event.target.checked)}
          className="size-4"
        />
      </div>
    );
  }

  if (entry.type === "uint32") {
    return (
      <div className="py-2">
        <Label htmlFor={fieldId}>{entry.key}</Label>
        <Input
          id={fieldId}
          type="number"
          min={0}
          step={1}
          value={typeof value === "string" ? value : ""}
          onChange={(event) => onChange(event.target.value)}
        />
        {error ? <p className="text-critical mt-1 text-sm">{error}</p> : null}
      </div>
    );
  }

  if (entry.type === "string") {
    const isSecret = secretSet !== undefined;
    return (
      <div className="py-2">
        <Label htmlFor={fieldId}>{entry.key}</Label>
        <Input
          id={fieldId}
          type={isSecret ? "password" : "text"}
          value={typeof value === "string" ? value : ""}
          placeholder={
            isSecret
              ? secretSet
                ? "Already set — leave blank to keep"
                : "Not set"
              : undefined
          }
          onChange={(event) => onChange(event.target.value)}
        />
      </div>
    );
  }

  return (
    <div className="py-2">
      <p className="text-sm font-medium">{entry.key}</p>
      <p className="text-muted-foreground text-sm">
        Unsupported type &quot;{entry.type}&quot; — not editable here.
      </p>
    </div>
  );
}
