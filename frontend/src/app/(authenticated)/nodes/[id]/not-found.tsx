import Link from "next/link";

import { EmptyState } from "@/components/ui/states";

export default function NotFound() {
  return (
    <main className="mx-auto flex min-h-[calc(100vh-4rem)] w-full max-w-7xl items-center px-4 py-6 sm:px-6 lg:px-8">
      <EmptyState
        title="Node not found"
        description="This device may have been removed or the link is no longer valid."
        action={
          <Link
            className="bg-primary text-surface focus-visible:ring-focus focus-visible:ring-offset-background inline-flex min-h-11 items-center justify-center rounded-xl px-4 py-2.5 text-sm font-semibold transition-opacity hover:opacity-90 focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none"
            href="/nodes"
          >
            Return to nodes
          </Link>
        }
      />
    </main>
  );
}
