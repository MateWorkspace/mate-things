"use client";

import { useRouter } from "next/navigation";
import type { ReactNode } from "react";

import { useSmartRefresh } from "@/hooks/use-smart-refresh";

import RefreshControl from "./RefreshControl";

interface RefreshBoundaryProps {
  intervalMs?: number;
  updatedAt?: string;
  suspended?: boolean;
  children: ReactNode;
}

export default function RefreshBoundary({
  intervalMs = 30_000,
  updatedAt,
  suspended = false,
  children,
}: RefreshBoundaryProps) {
  const router = useRouter();
  const state = useSmartRefresh({
    intervalMs,
    suspended,
    updatedAt: updatedAt ? new Date(updatedAt) : new Date(),
    onRefresh: () => router.refresh(),
  });

  return (
    <section className="space-y-4">
      <RefreshControl state={state} />
      {children}
    </section>
  );
}
