import type { Metadata } from "next";

import RefreshBoundary from "@/components/refresh/RefreshBoundary";
import PageHeader from "@/components/ui/page-header";
import { EmptyState } from "@/components/ui/states";
import { listBroadcastSessions } from "@/lib/api/telemetry";
import { requirePermission } from "@/lib/session";

import BroadcastSessionCard from "./_components/BroadcastSessionCard";

export const metadata: Metadata = {
  title: "Broadcast Sessions — Mate Things",
};

export default async function BroadcastSessionsPage() {
  await requirePermission("broadcast_session:get");
  const result = await listBroadcastSessions();

  return (
    <main className="mx-auto w-full max-w-7xl space-y-6 px-4 py-6 sm:px-6 lg:px-8">
      <PageHeader
        title="Broadcast Sessions"
        description="Live telemetry websocket connections currently subscribed to broadcast updates."
      />
      <p className="text-muted-foreground text-xs font-semibold tracking-wider uppercase">
        {result.total_items} active session{result.total_items === 1 ? "" : "s"}
      </p>
      <RefreshBoundary intervalMs={10_000}>
        {result.data.length ? (
          <section
            aria-label="Broadcast sessions"
            className="grid grid-cols-1 gap-5 sm:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4"
          >
            {result.data.map((session) => (
              <BroadcastSessionCard key={session.id} session={session} />
            ))}
          </section>
        ) : (
          <EmptyState
            title="No active sessions"
            description="Open a live telemetry stream to see it appear here."
          />
        )}
      </RefreshBoundary>
    </main>
  );
}
