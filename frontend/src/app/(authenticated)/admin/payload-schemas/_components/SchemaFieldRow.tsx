"use client";

import { useId, useState } from "react";

import Input from "@/components/ui/input";
import Label from "@/components/ui/label";

import SchemaFieldList from "./SchemaFieldList";
import {
  FIELD_TYPES,
  FIELD_TYPE_LABELS,
  baseType,
  isArrayType,
  isEnumType,
  isObjectType,
  isValidFieldName,
  type FieldRow,
  type FieldType,
} from "./schema-definition-types";

interface SchemaFieldRowProps {
  row: FieldRow;
  path: string[];
  errors: Record<string, string>;
  siblingNames: string[];
  onChange: (row: FieldRow) => void;
  onRemove: () => void;
}

function NumberField({
  id,
  label,
  value,
  onChange,
}: {
  id: string;
  label: string;
  value: string;
  onChange: (value: string) => void;
}) {
  return (
    <div>
      <Label htmlFor={id}>{label}</Label>
      <Input
        id={id}
        type="number"
        value={value}
        onChange={(event) => onChange(event.target.value)}
      />
    </div>
  );
}

function OptionsField({
  id,
  options,
  onChange,
}: {
  id: string;
  options: string[];
  onChange: (options: string[]) => void;
}) {
  const [draft, setDraft] = useState("");
  return (
    <div className="sm:col-span-2">
      <Label htmlFor={id}>Options</Label>
      {options.length ? (
        <div className="mb-2 flex flex-wrap gap-2">
          {options.map((option) => (
            <span
              key={option}
              className="bg-muted flex items-center gap-1.5 rounded-full px-3 py-1 text-sm"
            >
              {option}
              <button
                type="button"
                aria-label={`Remove option ${option}`}
                onClick={() => onChange(options.filter((o) => o !== option))}
                className="text-muted-foreground hover:text-critical"
              >
                ✕
              </button>
            </span>
          ))}
        </div>
      ) : null}
      <Input
        id={id}
        value={draft}
        placeholder="Type an option and press Enter"
        onChange={(event) => setDraft(event.target.value)}
        onKeyDown={(event) => {
          if (event.key !== "Enter") return;
          event.preventDefault();
          const value = draft.trim();
          if (!value || options.includes(value)) return;
          onChange([...options, value]);
          setDraft("");
        }}
      />
    </div>
  );
}

export default function SchemaFieldRow({
  row,
  path,
  errors,
  siblingNames,
  onChange,
  onRemove,
}: SchemaFieldRowProps) {
  const id = useId();
  const [renaming, setRenaming] = useState(false);
  const [nameDraft, setNameDraft] = useState(row.name);
  const [nameError, setNameError] = useState<string>();
  const rowError = errors[path.join(".")];
  const type = baseType(row.type);

  function commitRename() {
    const name = nameDraft.trim();
    if (!name) {
      setNameError("Name is required.");
      return;
    }
    if (!isValidFieldName(name)) {
      setNameError(
        "Use lowercase letters, numbers, and underscores only (e.g. sample_rate).",
      );
      return;
    }
    if (name !== row.name && siblingNames.includes(name)) {
      setNameError("A field with this name already exists here.");
      return;
    }
    onChange({ ...row, name });
    setRenaming(false);
    setNameError(undefined);
  }

  return (
    <div className="border-control-border rounded-xl border p-3">
      <div className="flex flex-wrap items-center gap-3">
        {renaming ? (
          <Input
            autoFocus
            value={nameDraft}
            onChange={(event) =>
              setNameDraft(event.target.value.replace(/ +/g, "_"))
            }
            onBlur={commitRename}
            onKeyDown={(event) => {
              if (event.key === "Enter") {
                event.preventDefault();
                commitRename();
              }
            }}
            className="max-w-48"
          />
        ) : (
          <button
            type="button"
            onClick={() => {
              setNameDraft(row.name);
              setRenaming(true);
            }}
            className="bg-muted rounded-lg px-2.5 py-1 font-mono text-sm font-semibold"
          >
            {row.name}
          </button>
        )}

        <select
          aria-label={`${row.name} type`}
          value={row.type}
          onChange={(event) => {
            const nextType = event.target.value as FieldType;
            if (nextType === row.type) return;
            onChange({
              ...row,
              type: nextType,
              minimum: "",
              maximum: "",
              minimumLength: "",
              maximumLength: "",
              minimumItem: "",
              maximumItem: "",
              unit: "",
              options: [],
            });
          }}
          className="border-control-border bg-background rounded-lg border px-2.5 py-1.5 text-sm"
        >
          {FIELD_TYPES.map((option) => (
            <option key={option} value={option}>
              {FIELD_TYPE_LABELS[option]}
            </option>
          ))}
        </select>

        <label className="flex items-center gap-1.5 text-sm">
          <input
            type="checkbox"
            checked={row.required}
            onChange={(event) =>
              onChange({ ...row, required: event.target.checked })
            }
            className="accent-primary size-4"
          />
          Required
        </label>

        <button
          type="button"
          aria-label={`Remove ${row.name}`}
          onClick={onRemove}
          className="text-critical hover:bg-critical/10 ml-auto rounded-lg px-2 py-1 text-sm transition-colors"
        >
          ✕
        </button>
      </div>

      {nameError ? (
        <p className="text-critical mt-1 text-sm">{nameError}</p>
      ) : null}
      {rowError ? <p className="text-critical mt-1 text-sm">{rowError}</p> : null}

      <div className="mt-3 grid gap-3 sm:grid-cols-2">
        {isArrayType(row.type) ? (
          <>
            <NumberField
              id={`${id}-min-items`}
              label="Min items"
              value={row.minimumItem}
              onChange={(value) => onChange({ ...row, minimumItem: value })}
            />
            <NumberField
              id={`${id}-max-items`}
              label="Max items"
              value={row.maximumItem}
              onChange={(value) => onChange({ ...row, maximumItem: value })}
            />
          </>
        ) : null}

        {type === "string" ? (
          <>
            <NumberField
              id={`${id}-min-length`}
              label="Min length"
              value={row.minimumLength}
              onChange={(value) => onChange({ ...row, minimumLength: value })}
            />
            <NumberField
              id={`${id}-max-length`}
              label="Max length"
              value={row.maximumLength}
              onChange={(value) => onChange({ ...row, maximumLength: value })}
            />
          </>
        ) : null}

        {type === "float" || type === "integer" ? (
          <>
            <NumberField
              id={`${id}-min`}
              label="Min"
              value={row.minimum}
              onChange={(value) => onChange({ ...row, minimum: value })}
            />
            <NumberField
              id={`${id}-max`}
              label="Max"
              value={row.maximum}
              onChange={(value) => onChange({ ...row, maximum: value })}
            />
            <div>
              <Label htmlFor={`${id}-unit`}>Unit</Label>
              <Input
                id={`${id}-unit`}
                value={row.unit}
                placeholder="e.g. °C"
                onChange={(event) =>
                  onChange({ ...row, unit: event.target.value })
                }
              />
            </div>
          </>
        ) : null}

        {isEnumType(row.type) ? (
          <OptionsField
            id={`${id}-options`}
            options={row.options}
            onChange={(options) => onChange({ ...row, options })}
          />
        ) : null}
      </div>

      {isObjectType(row.type) ? (
        <div className="border-control-border mt-3 space-y-3 border-l-2 pl-4">
          <SchemaFieldList
            rows={row.children}
            path={path}
            errors={errors}
            onChange={(children) => onChange({ ...row, children })}
          />
        </div>
      ) : null}
    </div>
  );
}
