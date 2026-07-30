import { UserRound } from "lucide-react";

import Button from "@/components/ui/button";
import type { UserResponse } from "@/lib/api/users";

interface ProfileViewProps {
  user: UserResponse;
  permissions: readonly string[];
  canEdit: boolean;
  onEdit: () => void;
}

export default function ProfileView({
  user,
  permissions,
  canEdit,
  onEdit,
}: ProfileViewProps) {
  return (
    <div className="space-y-5">
      <div className="flex items-center gap-3">
        <span
          role="img"
          aria-label="Anonymous profile icon"
          className="bg-surface text-primary flex size-12 shrink-0 items-center justify-center rounded-full"
        >
          <UserRound aria-hidden="true" className="size-6" />
        </span>
        <div className="min-w-0">
          <p className="truncate font-semibold">{user.name}</p>
          <p className="text-foreground/60 truncate text-sm">
            @{user.username}
          </p>
        </div>
      </div>

      <dl className="grid gap-4 sm:grid-cols-2">
        <div>
          <dt className="text-foreground/60 text-xs font-semibold tracking-wide uppercase">
            Name
          </dt>
          <dd className="mt-1 text-sm">{user.name}</dd>
        </div>
        <div>
          <dt className="text-foreground/60 text-xs font-semibold tracking-wide uppercase">
            Username
          </dt>
          <dd className="mt-1 text-sm">{user.username}</dd>
        </div>
        <div className="sm:col-span-2">
          <dt className="text-foreground/60 text-xs font-semibold tracking-wide uppercase">
            Bio
          </dt>
          <dd className="mt-1 text-sm">{user.bio || "No biography added."}</dd>
        </div>
        <div className="sm:col-span-2">
          <dt className="text-foreground/60 text-xs font-semibold tracking-wide uppercase">
            Role identifier
          </dt>
          <dd className="mt-1 text-sm break-all">{user.role_id}</dd>
        </div>
        <div className="sm:col-span-2">
          <dt className="text-foreground/60 text-xs font-semibold tracking-wide uppercase">
            Effective permissions
          </dt>
          <dd className="mt-2">
            {permissions.length > 0 ? (
              <ul className="flex flex-wrap gap-2">
                {permissions.map((permission) => (
                  <li
                    key={permission}
                    className="bg-muted text-foreground rounded-lg px-2.5 py-1 text-xs"
                  >
                    {permission}
                  </li>
                ))}
              </ul>
            ) : (
              <span className="text-foreground/60 text-sm">
                No effective permissions.
              </span>
            )}
          </dd>
        </div>
      </dl>

      {canEdit ? (
        <Button type="button" variant="secondary" onClick={onEdit}>
          Edit profile
        </Button>
      ) : null}
    </div>
  );
}
