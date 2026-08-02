"use client";

import { useState } from "react";
import { Check, Copy } from "lucide-react";

interface JsonPayloadProps {
  value: unknown;
  summary?: string;
  hideToggle?: boolean;
}

export default function JsonPayload({
  value,
  summary = "View payload",
  hideToggle = false,
}: JsonPayloadProps) {
  const [copied, setCopied] = useState(false);
  const json = JSON.stringify(value, null, 2);

  const copyButton = (
    <button
      type="button"
      aria-label="Copy JSON payload"
      className="border-border bg-background/80 focus-visible:ring-focus absolute top-2 right-2 inline-flex min-h-9 items-center gap-2 rounded-xl border px-3 text-xs font-semibold backdrop-blur focus-visible:ring-2 focus-visible:outline-none"
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
  );

  const body = (
    <div className="relative">
      {copyButton}
      <pre className="bg-muted max-h-80 overflow-auto rounded-xl p-3 pr-24 text-xs whitespace-pre-wrap">
        {json}
      </pre>
    </div>
  );

  if (hideToggle) {
    return (
      <div className="border-border bg-background rounded-xl border p-3">
        {body}
      </div>
    );
  }

  return (
    <details className="border-border bg-background rounded-xl border p-3">
      <summary className="text-primary cursor-pointer text-sm font-semibold">
        {summary}
      </summary>
      <div className="mt-3">{body}</div>
    </details>
  );
}
