import Card from "@/components/ui/card";
import StatusBadge from "@/components/ui/status-badge";
import type { NodeResponse } from "@/lib/api";

interface NodeOverviewProps {
  node: NodeResponse;
}

const DATE_FORMATTER = new Intl.DateTimeFormat("en", {
  dateStyle: "medium",
  timeStyle: "short",
  timeZone: "UTC",
});

function Detail({
  label,
  children,
}: {
  label: string;
  children: React.ReactNode;
}) {
  return (
    <div className="flex flex-col gap-1 sm:flex-row sm:justify-between sm:gap-6">
      <dt className="text-muted-foreground text-sm">{label}</dt>
      <dd className="min-w-0 text-sm font-medium break-words sm:text-right">
        {children}
      </dd>
    </div>
  );
}

export default function NodeOverview({ node }: NodeOverviewProps) {
  const freshness = node.updated_at ?? node.created_at;

  return (
    <section aria-labelledby="node-identity-heading">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2
            id="node-identity-heading"
            className="font-display text-primary text-2xl tracking-wide"
          >
            Node identity
          </h2>
          <p className="text-foreground/70 mt-1 text-sm">
            Backend-reported identity, assignment, and lifecycle details.
          </p>
        </div>
        <StatusBadge variant={node.is_connected ? "success" : "critical"}>
          {node.is_connected ? "Connected" : "Disconnected"}
        </StatusBadge>
      </div>

      <div className="mt-4 grid gap-4 lg:grid-cols-2">
        <Card>
          <h3 className="font-display text-foreground text-lg tracking-wide">
            Device
          </h3>
          <dl className="mt-4 space-y-3">
            <Detail label="Device ID">
              <span className="font-mono">{node.device_id}</span>
            </Detail>
            <Detail label="Device info">
              {node.device_info || "Not reported"}
            </Detail>
            <Detail label="Description">
              {node.description || "No description provided"}
            </Detail>
          </dl>
        </Card>

        <Card>
          <h3 className="font-display text-foreground text-lg tracking-wide">
            Assignment and audit
          </h3>
          <dl className="mt-4 space-y-3">
            <Detail label="Node class ID">
              {node.node_class_id ? (
                <span className="font-mono">{node.node_class_id}</span>
              ) : (
                "Unassigned"
              )}
            </Detail>
            <Detail label="Firmware ID">
              {node.firmware_id ? (
                <span className="font-mono">{node.firmware_id}</span>
              ) : (
                "Unassigned"
              )}
            </Detail>
            <Detail label={node.updated_at ? "Updated" : "Registered"}>
              <time dateTime={freshness}>
                {DATE_FORMATTER.format(new Date(freshness))} UTC
              </time>
            </Detail>
          </dl>
        </Card>
      </div>
    </section>
  );
}
