import { UserRound } from "lucide-react";

import Button from "@/components/ui/button";
import type { UserResponse } from "@/lib/api/users";

interface ProfileViewProps {
  user: UserResponse;
  roleName?: string;
  canEdit: boolean;
  onEdit: () => void;
}

export default function ProfileView({
  user,
  roleName,
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
          <p className="text-muted-foreground truncate text-sm">
            @{user.username}
          </p>
        </div>
      </div>

      <dl className="grid gap-4 sm:grid-cols-2">
        <div>
          <dt className="text-muted-foreground text-xs font-semibold tracking-wide uppercase">
            Name
          </dt>
          <dd className="mt-1 text-sm">{user.name}</dd>
        </div>
        <div>
          <dt className="text-muted-foreground text-xs font-semibold tracking-wide uppercase">
            Username
          </dt>
          <dd className="mt-1 text-sm">{user.username}</dd>
        </div>
        <div className="sm:col-span-2">
          <dt className="text-muted-foreground text-xs font-semibold tracking-wide uppercase">
            Bio
          </dt>
          <dd className="mt-1 text-sm">{user.bio || "No biography added."}</dd>
        </div>
        <div className="sm:col-span-2">
          <dt className="text-muted-foreground text-xs font-semibold tracking-wide uppercase">
            Role
          </dt>
          <dd className="mt-1 text-sm">{roleName ?? "Unavailable"}</dd>
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
