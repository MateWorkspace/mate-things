"use client";

import Button from "@/components/ui/button";

export default function Error({ reset }: { error: Error; reset: () => void }) {
  return (
    <main className="mx-auto max-w-3xl p-6">
      <section className="border-critical/40 rounded-2xl border p-6 text-center">
        <h1 className="font-display text-2xl">Unable to load users</h1>
        <p className="text-muted-foreground my-3">
          The administration service did not return this collection.
        </p>
        <Button onClick={reset}>Retry</Button>
      </section>
    </main>
  );
}
