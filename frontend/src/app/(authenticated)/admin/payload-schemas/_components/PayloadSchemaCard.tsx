import Link from "next/link";

import ResourceCard from "@/components/collection/ResourceCard";
import LocalDateTime from "@/components/ui/local-date-time";
import StatusBadge from "@/components/ui/status-badge";
import type { PayloadSchemaResponse } from "@/lib/api/payload-schemas";

const REFERENCE_TIME = Date.now();

function status(schema: PayloadSchemaResponse): {
  label: string;
  variant: "success" | "warning" | "neutral";
} {
  const now = REFERENCE_TIME;
  if (new Date(schema.valid_from).valueOf() > now)
    return { label: "Future", variant: "warning" };
  if (schema.valid_to && new Date(schema.valid_to).valueOf() <= now)
    return { label: "Expired", variant: "neutral" };
  return { label: "Current", variant: "success" };
}
export default function PayloadSchemaCard({
  schema,
}: {
  schema: PayloadSchemaResponse;
}) {
  const state = status(schema);
  return (
    <ResourceCard
      title={`${schema.name} v${schema.version}`}
      actions={<StatusBadge variant={state.variant}>{state.label}</StatusBadge>}
    >
      <dl className="space-y-3">
        <div>
          <dt className="text-muted-foreground">Valid from</dt>
          <dd>
            <LocalDateTime value={schema.valid_from} />
          </dd>
        </div>
        <div>
          <dt className="text-muted-foreground">Valid to</dt>
          <dd>
            {schema.valid_to ? (
              <LocalDateTime value={schema.valid_to} />
            ) : (
              "Open-ended"
            )}
          </dd>
        </div>
        <div>
          <dt className="text-muted-foreground">
            {schema.updated_at ? "Updated" : "Created"}
          </dt>
          <dd>
            <LocalDateTime
              value={schema.updated_at ?? schema.created_at}
              variant="date"
            />
          </dd>
        </div>
      </dl>
      <Link
        href={`/admin/payload-schemas/${schema.id}`}
        className="bg-primary text-surface mt-5 inline-flex min-h-11 w-full items-center justify-center rounded-xl text-sm font-semibold"
      >
        View schema
      </Link>
    </ResourceCard>
  );
}
