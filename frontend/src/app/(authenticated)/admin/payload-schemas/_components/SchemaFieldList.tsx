"use client";

import { useId, useState } from "react";

import Input from "@/components/ui/input";

import SchemaFieldRow from "./SchemaFieldRow";
import { createEmptyRow, isValidFieldName, type FieldRow } from "./schema-definition-types";

interface SchemaFieldListProps {
  rows: FieldRow[];
  onChange: (rows: FieldRow[]) => void;
  path: string[];
  errors: Record<string, string>;
}

export default function SchemaFieldList({
  rows,
  onChange,
  path,
  errors,
}: SchemaFieldListProps) {
  const [draftName, setDraftName] = useState("");
  const [draftError, setDraftError] = useState<string>();
  const inputId = useId();

  function commitDraft() {
    const name = draftName.trim();
    if (!name) return;
    if (!isValidFieldName(name)) {
      setDraftError(
        "Use lowercase letters, numbers, and underscores only (e.g. sample_rate).",
      );
      return;
    }
    if (rows.some((row) => row.name === name)) {
      setDraftError("A field with this name already exists here.");
      return;
    }
    onChange([...rows, createEmptyRow(name)]);
    setDraftName("");
    setDraftError(undefined);
  }

  return (
    <div className="space-y-3">
      {rows.map((row) => (
        <SchemaFieldRow
          key={row.key}
          row={row}
          path={[...path, row.name]}
          errors={errors}
          siblingNames={rows
            .filter((sibling) => sibling.key !== row.key)
            .map((sibling) => sibling.name)}
          onChange={(next) =>
            onChange(rows.map((sibling) => (sibling.key === row.key ? next : sibling)))
          }
          onRemove={() => onChange(rows.filter((sibling) => sibling.key !== row.key))}
        />
      ))}
      <div>
        <Input
          id={inputId}
          value={draftName}
          placeholder="variable_name"
          aria-label="New field name"
          onChange={(event) => {
            setDraftName(event.target.value.replace(/ +/g, "_"));
            setDraftError(undefined);
          }}
          onKeyDown={(event) => {
            if (event.key === "Enter") {
              event.preventDefault();
              commitDraft();
            }
          }}
        />
        {draftError ? (
          <p className="text-critical mt-1 text-sm">{draftError}</p>
        ) : null}
      </div>
    </div>
  );
}
