"use client";

import { useEffect } from "react";

import Button from "@/components/ui/button";
import { ErrorState } from "@/components/ui/states";

interface AuthenticatedErrorProps {
  error: Error & { digest?: string };
  unstable_retry: () => void;
}

export default function AuthenticatedError({
  error,
  unstable_retry,
}: AuthenticatedErrorProps) {
  useEffect(() => {
    console.error("Authenticated route failed", error.digest ?? error.name);
  }, [error]);

  return (
    <main className="mx-auto flex min-h-[calc(100vh-4rem)] w-full max-w-7xl items-center px-4 py-6 sm:px-6 lg:px-8">
      <ErrorState
        title="Workspace unavailable"
        description="This page could not be loaded. Your session and existing data have not been changed."
        action={
          <Button type="button" onClick={unstable_retry}>
            Try again
          </Button>
        }
      />
    </main>
  );
}
