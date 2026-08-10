import type { Metadata } from "next";
import { notFound } from "next/navigation";

import PreferencesDialog from "@/components/preferences/PreferencesDialog";
import Card from "@/components/ui/card";
import LocalDateTime from "@/components/ui/local-date-time";
import PageHeader from "@/components/ui/page-header";
import { ApiError } from "@/lib/api/client";
import { listAllRoles } from "@/lib/api/roles";
import { getUserById, getUserPermissions } from "@/lib/api/users";
import { requirePermission } from "@/lib/session";

import PasswordResetForm from "../_components/PasswordResetForm";
import UserForm from "../_components/UserForm";

export const metadata: Metadata = { title: "User details — Mate Things" };

export default async function UserDetails({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const [{ id }, session] = await Promise.all([
    params,
    requirePermission("user:get"),
  ]);
  let user;
  try {
    user = await getUserById(id);
  } catch (error) {
    if (error instanceof ApiError && error.status === 404) notFound();
    throw error;
  }
  const [roles, effective] = await Promise.all([
    session.permissions.has("role:get") ? listAllRoles() : Promise.resolve([]),
    session.permissions.has("user_permission:get")
      ? getUserPermissions(id)
      : Promise.resolve(null),
  ]);
  const role = roles.find((item) => item.id === user.role_id);
  return (
    <main className="mx-auto w-full max-w-7xl space-y-6 px-4 py-6 sm:px-6 lg:px-8">
      <PageHeader
        title={user.name}
        description={`@${user.username}`}
        actions={
          <div className="flex flex-wrap gap-2">
            {session.permissions.has("user:set") &&
            session.permissions.has("role:get") ? (
              <UserForm
                user={user}
                roles={roles}
                canDelete={session.permissions.has("user:remove")}
                currentUserId={session.user.id}
              />
            ) : null}
            {session.permissions.has("user_password:set") ? (
              <PasswordResetForm user={user} />
            ) : null}
            <PreferencesDialog
              resource="user"
              id={user.id}
              preferences={user.preferences}
              permissions={[...session.permissions]}
            />
          </div>
        }
      />
      <section className="grid gap-5 lg:grid-cols-2">
        <Card>
          <h2 className="font-display text-primary text-xl">Account</h2>
          <dl className="mt-4 space-y-3 text-sm">
            <div>
              <dt className="text-muted-foreground">Role</dt>
              <dd className="font-semibold">{role?.name ?? user.role_id}</dd>
            </div>
            <div>
              <dt className="text-muted-foreground">Bio</dt>
              <dd>{user.bio || "No bio provided."}</dd>
            </div>
            <div>
              <dt className="text-muted-foreground">User ID</dt>
              <dd className="font-mono break-all">{user.id}</dd>
            </div>
          </dl>
        </Card>
        <Card>
          <h2 className="font-display text-primary text-xl">Audit context</h2>
          <dl className="mt-4 space-y-3 text-sm">
            <div>
              <dt className="text-muted-foreground">Created</dt>
              <dd>
                <LocalDateTime value={user.created_at} />
              </dd>
            </div>
            {user.updated_at ? (
              <div>
                <dt className="text-muted-foreground">Updated</dt>
                <dd>
                  <LocalDateTime value={user.updated_at} />
                </dd>
              </div>
            ) : null}
          </dl>
        </Card>
      </section>
      <Card>
        <h2 className="font-display text-primary text-xl">
          Effective permissions
        </h2>
        {effective ? (
          <div className="mt-4 flex flex-wrap gap-2">
            {effective.map((permission) => (
              <span
                key={permission.id}
                className="border-border bg-muted rounded-full border px-3 py-1 font-mono text-xs"
              >
                {permission.name}
              </span>
            ))}
          </div>
        ) : (
          <p className="text-muted-foreground mt-3 text-sm">
            user_permission:get is required to inspect this account&apos;s
            effective permissions.
          </p>
        )}
      </Card>
      <Card>
        <h2 className="font-display text-primary text-xl">Preferences</h2>
        <pre className="bg-muted mt-4 overflow-auto rounded-xl p-4 text-sm">
          {JSON.stringify(user.preferences, null, 2)}
        </pre>
      </Card>
    </main>
  );
}
