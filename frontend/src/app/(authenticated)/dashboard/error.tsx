"use client";

import Button from "@/components/ui/button";
import { ErrorState } from "@/components/ui/states";

interface DashboardErrorProps {
  unstable_retry: () => void;
}

export default function DashboardError({
  unstable_retry,
}: DashboardErrorProps) {
  return (
    <main className="mx-auto flex min-h-[calc(100vh-4rem)] w-full max-w-7xl items-center px-4 py-6 sm:px-6 lg:px-8">
      <ErrorState
        title="Dashboard unavailable"
        description="The fleet overview could not be loaded. Try again to refresh this page."
        action={
          <Button type="button" onClick={unstable_retry}>
            Try again
          </Button>
        }
      />
    </main>
  );
}
