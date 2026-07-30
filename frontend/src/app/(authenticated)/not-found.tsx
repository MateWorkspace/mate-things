import Link from "next/link";

import { EmptyState } from "@/components/ui/states";

export default function AuthenticatedNotFound() {
  return (
    <main className="mx-auto flex min-h-[calc(100vh-4rem)] w-full max-w-7xl items-center px-4 py-6 sm:px-6 lg:px-8">
      <EmptyState
        title="Resource not found"
        description="The requested record may have been removed, renamed, or is no longer available to your account."
        action={
          <Link
            href="/dashboard"
            className="bg-primary text-surface focus-visible:ring-focus focus-visible:ring-offset-background inline-flex min-h-11 items-center justify-center rounded-xl px-4 py-2.5 text-sm font-semibold transition-opacity hover:opacity-90 focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none"
          >
            Return to dashboard
          </Link>
        }
      />
    </main>
  );
}
