"use client";

import { useState } from "react";
import { Check, Copy } from "lucide-react";

interface JsonPayloadProps {
  value: unknown;
  summary?: string;
}

export default function JsonPayload({
  value,
  summary = "View payload",
}: JsonPayloadProps) {
  const [copied, setCopied] = useState(false);
  const json = JSON.stringify(value, null, 2);

  return (
    <details className="border-border bg-background rounded-xl border p-3">
      <summary className="text-primary cursor-pointer text-sm font-semibold">
        {summary}
      </summary>
      <div className="mt-3 flex justify-end">
        <button
          type="button"
          aria-label="Copy JSON payload"
          className="border-border focus-visible:ring-focus inline-flex min-h-10 items-center gap-2 rounded-xl border px-3 text-xs font-semibold focus-visible:ring-2 focus-visible:outline-none"
          onClick={async () => {
            await navigator.clipboard.writeText(json);
            setCopied(true);
            window.setTimeout(() => setCopied(false), 1500);
          }}
        >
          {copied ? (
            <Check aria-hidden="true" className="size-4" />
          ) : (
            <Copy aria-hidden="true" className="size-4" />
          )}
          {copied ? "Copied" : "Copy"}
        </button>
      </div>
      <pre className="bg-muted mt-2 max-h-80 overflow-auto rounded-xl p-3 text-xs whitespace-pre-wrap">
        {json}
      </pre>
    </details>
  );
}
