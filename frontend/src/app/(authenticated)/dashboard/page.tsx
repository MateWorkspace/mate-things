import type { Metadata } from "next";

import PageHeader from "@/components/ui/page-header";
import { EmptyState } from "@/components/ui/states";
import { requireSessionContext } from "@/lib/session";

import FleetMetricSkeleton from "./_components/FleetMetricSkeleton";

export const metadata: Metadata = {
  title: "Dashboard — Mate Things",
};

const METRIC_POSITIONS = [
  { title: "Fleet metrics", permissions: ["node:get"] },
  { title: "Firmware metrics", permissions: ["firmware:get"] },
  {
    title: "Action metrics",
    permissions: ["action:get", "action_log:get"],
  },
  {
    title: "Observability metrics",
    permissions: ["telemetry_record:get", "node_log:get"],
  },
] as const;

export default async function DashboardPage() {
  const { permissions } = await requireSessionContext();
  const visibleMetricPositions = METRIC_POSITIONS.filter((position) =>
    position.permissions.some((permission) => permissions.has(permission)),
  );

  return (
    <main className="mx-auto w-full max-w-7xl space-y-8 px-4 py-6 sm:px-6 lg:px-8">
      <PageHeader
        title="Fleet overview"
        description="Fleet metrics will appear here when dashboard data is connected."
      />

      {visibleMetricPositions.length > 0 ? (
        <section aria-label="Fleet metrics">
          <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
            {visibleMetricPositions.map((position) => (
              <FleetMetricSkeleton
                key={position.title}
                title={position.title}
              />
            ))}
          </div>
        </section>
      ) : null}

      <section aria-labelledby="urgent-attention-heading">
        <h2
          id="urgent-attention-heading"
          className="font-display text-primary text-xl tracking-wide"
        >
          Urgent attention
        </h2>
        <div className="mt-4">
          <EmptyState
            title="No urgent attention items"
            description="Urgent fleet conditions will appear here when dashboard data is connected."
          />
        </div>
      </section>

      <section aria-labelledby="recent-warnings-heading">
        <h2
          id="recent-warnings-heading"
          className="font-display text-primary text-xl tracking-wide"
        >
          Recent warnings
        </h2>
        <div className="mt-4">
          <EmptyState
            title="No recent warnings"
            description="Recent fleet warnings will appear here when dashboard data is connected."
          />
        </div>
      </section>
    </main>
  );
}
