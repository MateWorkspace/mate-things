import Button from "@/components/ui/button";
import type { UserResponse } from "@/lib/api/users";

interface ProfileViewProps {
  user: UserResponse;
  canEdit: boolean;
  onEdit: () => void;
}

export default function ProfileView({
  user,
  canEdit,
  onEdit,
}: ProfileViewProps) {
  return (
    <div className="space-y-5">
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
      </dl>

      {canEdit ? (
        <Button type="button" variant="secondary" onClick={onEdit}>
          Edit profile
        </Button>
      ) : null}
    </div>
  );
}
