"use client";
import Button from "@/components/ui/button";
import { ErrorState } from "@/components/ui/states";
export default function Error({ reset }: { reset: () => void }) {
  return (
    <main className="mx-auto max-w-5xl p-6">
      <ErrorState
        title="Node logs could not be loaded"
        description="Device events are temporarily unavailable."
        action={<Button onClick={reset}>Retry</Button>}
      />
    </main>
  );
}
