"use client";

import Button from "@/components/ui/button";
import { ErrorState } from "@/components/ui/states";

interface FirmwareDetailErrorProps {
  unstable_retry: () => void;
}

export default function FirmwareDetailError({
  unstable_retry,
}: FirmwareDetailErrorProps) {
  return (
    <main className="mx-auto flex min-h-[calc(100vh-4rem)] w-full max-w-7xl items-center px-4 py-6 sm:px-6 lg:px-8">
      <ErrorState
        title="Firmware details unavailable"
        description="This firmware workspace could not be loaded."
        action={
          <Button type="button" onClick={unstable_retry}>
            Retry
          </Button>
        }
      />
    </main>
  );
}
