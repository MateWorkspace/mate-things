"use client";

import { useId, useState } from "react";

import Button from "@/components/ui/button";
import Label from "@/components/ui/label";

interface JsonEditorProps {
  name: string;
  label: string;
  defaultValue: unknown;
  objectOnly?: boolean;
}

function formatJson(value: unknown): string {
  return JSON.stringify(value ?? {}, null, 2);
}

export default function JsonEditor({
  name,
  label,
  defaultValue,
  objectOnly = true,
}: JsonEditorProps) {
  const id = useId();
  const [value, setValue] = useState(() => formatJson(defaultValue));
  const [error, setError] = useState<string>();
  const validate = (candidate: string): unknown => {
    try {
      const parsed: unknown = JSON.parse(candidate);
      if (
        objectOnly &&
        (parsed === null || Array.isArray(parsed) || typeof parsed !== "object")
      ) {
        throw new Error("Enter a JSON object, not an array or primitive.");
      }
      setError(undefined);
      return parsed;
    } catch (parseError) {
      const message =
        parseError instanceof SyntaxError
          ? "Enter valid JSON before continuing."
          : parseError instanceof Error
            ? parseError.message
            : "Enter valid JSON before continuing.";
      setError(message);
      return undefined;
    }
  };

  return (
    <div>
      <Label htmlFor={id}>{label}</Label>
      <textarea
        id={id}
        name={name}
        value={value}
        rows={12}
        spellCheck={false}
        aria-invalid={Boolean(error)}
        aria-describedby={error ? `${id}-error` : undefined}
        className="border-control-border bg-muted text-foreground focus-visible:border-focus focus-visible:ring-focus w-full resize-y rounded-xl border px-3.5 py-3 font-mono text-sm leading-6 focus-visible:ring-2 focus-visible:outline-none"
        onBlur={() => validate(value)}
        onChange={(event) => {
          const next = event.target.value;
          setValue(next);
          validate(next);
        }}
      />
      {error ? (
        <p
          id={`${id}-error`}
          role="alert"
          className="text-critical mt-1.5 text-sm"
        >
          {error}
        </p>
      ) : null}
      <div className="mt-2 flex flex-wrap gap-2">
        <Button
          type="button"
          variant="secondary"
          onClick={() => validate(value)}
        >
          Parse
        </Button>
        <Button
          type="button"
          variant="secondary"
          onClick={() => {
            const parsed = validate(value);
            if (parsed !== undefined) setValue(formatJson(parsed));
          }}
        >
          Format
        </Button>
        <Button
          type="button"
          variant="secondary"
          onClick={() => {
            if (validate(value) !== undefined) {
              void navigator.clipboard?.writeText(value);
            }
          }}
        >
          Copy
        </Button>
      </div>
    </div>
  );
}
