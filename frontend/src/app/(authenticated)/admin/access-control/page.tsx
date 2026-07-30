import type { Metadata } from "next";

import PageHeader from "@/components/ui/page-header";
import { EmptyState } from "@/components/ui/states";
import { listPermissions } from "@/lib/api/permissions";
import { getRolePermissions, listRoles } from "@/lib/api/roles";
import { requireAnyPermission } from "@/lib/session";

import AccessCreateActions from "./_components/AccessCreateActions";
import AccessTabs from "./_components/AccessTabs";
import PermissionCard from "./_components/PermissionCard";
import RoleCard from "./_components/RoleCard";
import RoleDetails from "./_components/RoleDetails";

export const metadata: Metadata = { title: "Access Control — Mate Things" };

export default async function AccessControlPage({
  searchParams,
}: {
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}) {
  const [session, raw] = await Promise.all([
    requireAnyPermission(["role:get", "permission:get", "role_permission:get"]),
    searchParams,
  ]);
  const canRoles = session.permissions.has("role:get");
  const canPermissions = session.permissions.has("permission:get");
  let tab: "roles" | "permissions" =
    String(raw.tab) === "permissions" ? "permissions" : "roles";
  if (tab === "roles" && !canRoles && canPermissions) tab = "permissions";
  if (tab === "permissions" && !canPermissions && canRoles) tab = "roles";
  const [rolesResult, permissionsResult] = await Promise.all([
    canRoles ? listRoles({ limit: 48 }) : Promise.resolve(null),
    canPermissions ? listPermissions({ limit: 48 }) : Promise.resolve(null),
  ]);
  const roleId = String(raw.role ?? "");
  const selectedRole = rolesResult?.data.find((role) => role.id === roleId);
  const assignments =
    selectedRole && session.permissions.has("role_permission:get")
      ? await getRolePermissions(selectedRole.id)
      : [];
  return (
    <main className="mx-auto w-full max-w-7xl space-y-6 px-4 py-6 sm:px-6 lg:px-8">
      <PageHeader
        title="Access Control"
        description="Manage roles, permissions, defaults, and assignments in one workspace."
        actions={
          <AccessCreateActions
            canAddRole={session.permissions.has("role:add")}
            canAddPermission={session.permissions.has("permission:add")}
            tab={tab}
          />
        }
      />
      <AccessTabs
        active={tab}
        showRoles={canRoles}
        showPermissions={canPermissions}
      />
      {tab === "roles" ? (
        selectedRole ? (
          <RoleDetails
            role={selectedRole}
            permissions={permissionsResult?.data ?? []}
            selected={assignments.map((item) => item.id)}
            grants={[...session.permissions]}
          />
        ) : rolesResult?.data.length ? (
          <section className="grid gap-5 sm:grid-cols-2 xl:grid-cols-3">
            {rolesResult.data.map((role) => (
              <RoleCard key={role.id} role={role} />
            ))}
          </section>
        ) : (
          <EmptyState
            title="No roles yet"
            description="Create a role to start composing access."
          />
        )
      ) : permissionsResult?.data.length ? (
        <section className="grid gap-5 sm:grid-cols-2 xl:grid-cols-3">
          {permissionsResult.data.map((permission) => (
            <PermissionCard
              key={permission.id}
              permission={permission}
              grants={[...session.permissions]}
            />
          ))}
        </section>
      ) : (
        <EmptyState
          title="No permissions yet"
          description="Create a permission to expose a backend capability."
        />
      )}
    </main>
  );
}
