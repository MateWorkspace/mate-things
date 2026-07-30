"use client";
import Button from "@/components/ui/button";
import { ErrorState } from "@/components/ui/states";
export default function Error({ reset }: { reset: () => void }) {
  return (
    <main className="mx-auto max-w-5xl p-6">
      <ErrorState
        title="Action history could not be loaded"
        description="Existing results are unavailable right now."
        action={<Button onClick={reset}>Retry</Button>}
      />
    </main>
  );
}
