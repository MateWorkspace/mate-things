"use client";

import { useState } from "react";
import { Check, Copy } from "lucide-react";

export default function CodeBlock({
  value,
  label,
}: {
  value: string;
  label: string;
}) {
  const [copied, setCopied] = useState(false);

  return (
    <div className="relative">
      <button
        type="button"
        aria-label={`Copy ${label}`}
        className="border-border bg-background/80 focus-visible:ring-focus absolute top-2 right-2 inline-flex min-h-9 items-center gap-2 rounded-xl border px-3 text-xs font-semibold backdrop-blur focus-visible:ring-2 focus-visible:outline-none"
        onClick={async () => {
          await navigator.clipboard.writeText(value);
          setCopied(true);
          window.setTimeout(() => setCopied(false), 1500);
        }}
      >
        {copied ? (
          <Check aria-hidden="true" className="h-4 w-4" />
        ) : (
          <Copy aria-hidden="true" className="h-4 w-4" />
        )}
        {copied ? "Copied" : "Copy"}
      </button>
      <pre className="border-border bg-muted overflow-x-auto rounded-2xl border p-4 text-xs">
        <code>{value}</code>
      </pre>
    </div>
  );
}
