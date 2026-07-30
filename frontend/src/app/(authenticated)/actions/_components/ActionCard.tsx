import Link from "next/link";

import ResourceCard from "@/components/collection/ResourceCard";
import type { ActionResponse } from "@/lib/api/actions";

interface ActionCardProps {
  action: ActionResponse;
  nodeClassName?: string;
}

export default function ActionCard({ action, nodeClassName }: ActionCardProps) {
  return (
    <ResourceCard
      title={action.name}
      summary={nodeClassName ?? `Class ${action.node_class_id}`}
    >
      <p className="text-foreground/75 min-h-12">
        {action.description || "No description provided."}
      </p>
      <dl className="mt-4 space-y-2">
        <div className="flex justify-between gap-3">
          <dt className="text-muted-foreground">Schema</dt>
          <dd>
            {action.payload_schema_name} · v{action.payload_schema_version}
          </dd>
        </div>
        <div className="flex justify-between gap-3">
          <dt className="text-muted-foreground">Updated</dt>
          <dd>
            <time dateTime={action.updated_at ?? action.created_at}>
              {new Date(
                action.updated_at ?? action.created_at,
              ).toLocaleString()}
            </time>
          </dd>
        </div>
      </dl>
      <Link
        href={`/actions/${action.id}`}
        className="bg-primary text-surface mt-5 inline-flex min-h-11 w-full items-center justify-center rounded-xl px-4 text-sm font-semibold"
      >
        View and dispatch
      </Link>
    </ResourceCard>
  );
}
