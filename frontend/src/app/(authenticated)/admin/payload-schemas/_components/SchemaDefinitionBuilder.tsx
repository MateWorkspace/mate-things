"use client";

import { useEffect, useMemo, useState } from "react";

import SchemaFieldList from "./SchemaFieldList";
import {
  canRepresentDefinition,
  rootDefinitionFromRows,
  rowsFromRootDefinition,
  type FieldRow,
  type RawDefinition,
} from "./schema-definition-types";

const EMPTY_DEFINITION: RawDefinition = { type: "object", properties: {} };

interface SchemaDefinitionBuilderProps {
  name: string;
  defaultValue?: Record<string, unknown>;
  errors?: Record<string, string>;
  onRepresentableChange?: (representable: boolean) => void;
}

export default function SchemaDefinitionBuilder({
  name,
  defaultValue,
  errors = {},
  onRepresentableChange,
}: SchemaDefinitionBuilderProps) {
  const initial = (defaultValue as RawDefinition | undefined) ?? EMPTY_DEFINITION;
  const representable = useMemo(() => canRepresentDefinition(initial), [initial]);
  const [rows, setRows] = useState<FieldRow[]>(() =>
    representable ? rowsFromRootDefinition(initial) : [],
  );

  useEffect(() => {
    onRepresentableChange?.(representable);
    // Only re-run if the representability verdict itself changes, not on
    // every keystroke inside `rows` (onRepresentableChange isn't expected to
    // be referentially stable across parent re-renders).
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [representable]);

  if (!representable) {
    return (
      <div className="space-y-2">
        <p className="text-critical text-sm">
          This schema&apos;s definition uses a shape the guided editor can&apos;t
          represent. Edit it directly via the API, or contact an admin.
        </p>
        <pre className="bg-muted overflow-auto rounded-xl p-4 text-xs">
          {JSON.stringify(initial, null, 2)}
        </pre>
      </div>
    );
  }

  const definition = rootDefinitionFromRows(rows);

  return (
    <div className="space-y-3">
      <SchemaFieldList rows={rows} onChange={setRows} path={[]} errors={errors} />
      <input
        type="hidden"
        name={name}
        value={JSON.stringify(definition)}
        readOnly
      />
      <details className="text-sm">
        <summary className="text-muted-foreground cursor-pointer select-none">
          Preview JSON
        </summary>
        <pre className="bg-muted mt-2 overflow-auto rounded-xl p-4 text-xs">
          {JSON.stringify(definition, null, 2)}
        </pre>
      </details>
    </div>
  );
}
