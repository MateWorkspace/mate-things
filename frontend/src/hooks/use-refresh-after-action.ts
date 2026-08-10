"use client";

import { startTransition, useEffect, useRef } from "react";
import { useRouter } from "next/navigation";

interface RefreshableActionState {
  status: string;
}

export function useRefreshAfterAction(state: RefreshableActionState): void {
  const router = useRouter();
  const refreshed = useRef<RefreshableActionState | null>(null);

  useEffect(() => {
    if (
      (state.status !== "success" && state.status !== "partial") ||
      refreshed.current === state
    ) {
      return;
    }

    refreshed.current = state;
    startTransition(() => router.refresh());
  }, [router, state]);
}
