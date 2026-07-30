import type { Metadata } from "next";
import { Activity, CircleAlert, RadioTower, Server } from "lucide-react";

import PageHeader from "@/components/ui/page-header";
import { requireSessionContext } from "@/lib/session";

import DashboardRefresh from "./_components/DashboardRefresh";
import FleetMetricCard from "./_components/FleetMetricCard";
import RecentWarnings from "./_components/RecentWarnings";
import UrgentAttention from "./_components/UrgentAttention";
import {
  DASHBOARD_RECENT_WINDOW_HOURS,
  loadDashboardData,
} from "./_lib/dashboard-data";

export const metadata: Metadata = {
  title: "Fleet Overview — Mate Things",
  description:
    "Permission-aware fleet health, urgent operational issues, and recent device warnings.",
};

export default async function DashboardPage() {
  const { permissions } = await requireSessionContext();
  const data = await loadDashboardData(permissions);
  const disconnectedCount = data.nodes?.disconnected.length ?? 0;
  const hasMetrics =
    Boolean(data.nodes) ||
    data.failedActions !== undefined ||
    data.recentWarnings !== undefined;

  return (
    <main className="mx-auto w-full max-w-7xl space-y-8 px-4 py-6 sm:px-6 lg:px-8">
      <PageHeader
        title="Fleet overview"
        description="Live operational health, urgent issues, and recent warning signals across the resources you can access."
        actions={<DashboardRefresh loadedAt={data.loadedAt} />}
      />

      {hasMetrics ? (
        <section aria-labelledby="fleet-health-heading">
          <div>
            <h2
              id="fleet-health-heading"
              className="font-display text-primary text-2xl tracking-wide"
            >
              Operational health
            </h2>
            <p className="text-muted-foreground mt-1 text-sm">
              Log and action signals cover the last{" "}
              {DASHBOARD_RECENT_WINDOW_HOURS} hours.
            </p>
          </div>
          <div className="mt-4 grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
            {data.nodes ? (
              <FleetMetricCard
                description="Exact fleet inventory"
                href="/nodes"
                icon={Server}
                label="Total nodes"
                status="Inventory"
                statusVariant="info"
                value={data.nodes.total}
              />
            ) : null}
            {data.nodes ? (
              <FleetMetricCard
                description={`${data.nodes.sampled}-node page sample`}
                href="/nodes?connection=connected"
                icon={RadioTower}
                label="Connected"
                status="Sampled"
                statusVariant="success"
                value={data.nodes.connected}
              />
            ) : null}
            {data.nodes ? (
              <FleetMetricCard
                description={`${data.nodes.sampled}-node page sample`}
                href="/nodes?connection=disconnected"
                icon={CircleAlert}
                label="Disconnected"
                status={disconnectedCount > 0 ? "Attention" : "Clear"}
                statusVariant={disconnectedCount > 0 ? "critical" : "success"}
                value={disconnectedCount}
              />
            ) : null}
            {data.failedActions !== undefined ? (
              <FleetMetricCard
                description={`Failed or unresponded · ${DASHBOARD_RECENT_WINDOW_HOURS}h`}
                href={`/action-history?start=${encodeURIComponent(data.windowStartedAt)}`}
                icon={CircleAlert}
                label="Action issues"
                status={data.failedActions.length > 0 ? "Attention" : "Clear"}
                statusVariant={
                  data.failedActions.length > 0 ? "critical" : "success"
                }
                value={data.failedActions.length}
              />
            ) : null}
            {data.recentWarnings !== undefined ? (
              <FleetMetricCard
                description={`Warnings and errors · ${DASHBOARD_RECENT_WINDOW_HOURS}h`}
                href={`/node-logs?start=${encodeURIComponent(data.windowStartedAt)}`}
                icon={Activity}
                label="Log alerts"
                status={data.recentWarnings.length > 0 ? "Attention" : "Clear"}
                statusVariant={
                  data.recentWarnings.length > 0 ? "warning" : "success"
                }
                value={data.recentWarnings.length}
              />
            ) : null}
          </div>
        </section>
      ) : null}

      <UrgentAttention
        disconnectedNodes={data.nodes?.disconnected}
        failedActions={data.failedActions}
      />
      <RecentWarnings logs={data.recentWarnings} />
    </main>
  );
}
