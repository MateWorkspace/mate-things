import Link from "next/link";

import ResourceCard from "@/components/collection/ResourceCard";
import type { RoleResponse } from "@/lib/api/roles";
import type { UserResponse } from "@/lib/api/users";

export default function UserCard({
  user,
  role,
}: {
  user: UserResponse;
  role?: RoleResponse;
}) {
  const freshness = user.updated_at ?? user.created_at;
  return (
    <ResourceCard title={user.name} summary={`@${user.username}`}>
      <p className="text-foreground/70 line-clamp-3 min-h-15">
        {user.bio || "No bio provided."}
      </p>
      <dl className="mt-4 space-y-2">
        <div className="flex justify-between gap-3">
          <dt className="text-muted-foreground">Role</dt>
          <dd className="font-semibold">{role?.name ?? user.role_id}</dd>
        </div>
        <div className="flex justify-between gap-3">
          <dt className="text-muted-foreground">
            {user.updated_at ? "Updated" : "Created"}
          </dt>
          <dd>
            <time dateTime={freshness}>
              {new Date(freshness).toLocaleDateString("en", {
                timeZone: "UTC",
              })}
            </time>
          </dd>
        </div>
      </dl>
      <Link
        href={`/admin/users/${user.id}`}
        className="bg-primary text-surface focus-visible:ring-focus mt-5 inline-flex min-h-11 w-full items-center justify-center rounded-xl px-4 py-2.5 text-sm font-semibold focus-visible:ring-2 focus-visible:outline-none"
      >
        Manage user
      </Link>
    </ResourceCard>
  );
}
