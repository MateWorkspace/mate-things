"use client";

import { Check, Copy } from "lucide-react";
import { useState } from "react";

import Button from "@/components/ui/button";

const COPIED_RESET_MS = 2000;

export default function CopyButton({ value }: { value: string }) {
  const [copied, setCopied] = useState(false);

  async function handleCopy() {
    await navigator.clipboard.writeText(value);
    setCopied(true);
    setTimeout(() => setCopied(false), COPIED_RESET_MS);
  }

  return (
    <Button type="button" variant="secondary" onClick={handleCopy} className="gap-2">
      {copied ? (
        <>
          <Check aria-hidden="true" className="size-4" />
          Copied!
        </>
      ) : (
        <>
          <Copy aria-hidden="true" className="size-4" />
          Copy
        </>
      )}
    </Button>
  );
}
