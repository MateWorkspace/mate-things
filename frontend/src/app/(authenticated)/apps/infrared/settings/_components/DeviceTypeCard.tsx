import Link from "next/link";

import ResourceCard from "@/components/collection/ResourceCard";
import LocalDateTime from "@/components/ui/local-date-time";
import type { InfraredDeviceTypeResponse } from "@/lib/api/infrared";

import DeleteDeviceTypeDialog from "./DeleteDeviceTypeDialog";

interface DeviceTypeCardProps {
  deviceType: InfraredDeviceTypeResponse;
  canDelete: boolean;
}

export default function DeviceTypeCard({
  deviceType,
  canDelete,
}: DeviceTypeCardProps) {
  const freshness = deviceType.updated_at ?? deviceType.created_at;

  return (
    <ResourceCard
      title={deviceType.name}
      actions={
        canDelete ? <DeleteDeviceTypeDialog deviceType={deviceType} /> : null
      }
    >
      <dl>
        <div className="flex items-start justify-between gap-4">
          <dt className="text-muted-foreground">
            {deviceType.updated_at ? "Updated" : "Created"}
          </dt>
          <dd className="text-right font-medium">
            <LocalDateTime value={freshness} />
          </dd>
        </div>
      </dl>

      <Link
        className="bg-primary text-surface focus-visible:ring-focus focus-visible:ring-offset-background mt-5 inline-flex min-h-11 w-full items-center justify-center rounded-xl px-4 py-2.5 text-sm font-semibold transition-opacity hover:opacity-90 focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none"
        href={`/apps/infrared/settings/${deviceType.id}`}
      >
        Manage states
      </Link>
    </ResourceCard>
  );
}
