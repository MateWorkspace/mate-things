import Link from "next/link";

import ResourceCard from "@/components/collection/ResourceCard";
import StatusBadge from "@/components/ui/status-badge";
import type { RoleResponse } from "@/lib/api/roles";

export default function RoleCard({
  role,
  permissionCount,
}: {
  role: RoleResponse;
  permissionCount?: number;
}) {
  return (
    <ResourceCard
      title={role.name}
      actions={
        role.is_default ? (
          <StatusBadge variant="success">Default</StatusBadge>
        ) : undefined
      }
    >
      <p className="text-foreground/70 line-clamp-3 min-h-15">
        {role.description || "No role description."}
      </p>
      <Link
        href={`/admin/access-control?tab=roles&role=${role.id}`}
        className="bg-primary text-surface mt-5 inline-flex min-h-11 w-full items-center justify-center rounded-xl px-4 text-sm font-semibold"
      >
        Manage role
      </Link>
    </ResourceCard>
  );
}
