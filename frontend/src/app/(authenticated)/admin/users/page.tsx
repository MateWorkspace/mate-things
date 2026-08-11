import type { Metadata } from "next";
import { redirect } from "next/navigation";

import Pagination from "@/components/collection/Pagination";
import PageHeader from "@/components/ui/page-header";
import { EmptyState } from "@/components/ui/states";
import { getOptionalById } from "@/lib/api/optional";
import { getRoleById, listAllRoles } from "@/lib/api/roles";
import { listUsers } from "@/lib/api/users";
import {
  getOutOfRangePageRedirect,
  parsePageQuery,
} from "@/lib/collection-query";
import { firstQueryValue } from "@/lib/query";
import { requirePermission } from "@/lib/session";

import UserCard from "./_components/UserCard";
import UserFilters from "./_components/UserFilters";
import UserForm from "./_components/UserForm";

export const metadata: Metadata = { title: "Users — Mate Things" };
type Raw = Record<string, string | string[] | undefined>;

export default async function UsersPage({
  searchParams,
}: {
  searchParams: Promise<Raw>;
}) {
  const [{ permissions }, raw] = await Promise.all([
    requirePermission("user:get"),
    searchParams,
  ]);
  const query = parsePageQuery(raw);
  const canReadRoles = permissions.has("role:get");
  const roleId = firstQueryValue(raw.role_id)?.trim() || undefined;
  const [users, roles, selectedRole] = await Promise.all([
    listUsers({ ...query, role_id: roleId }),
    canReadRoles ? listAllRoles() : Promise.resolve([]),
    canReadRoles && roleId
      ? getOptionalById(() => getRoleById(roleId))
      : Promise.resolve(null),
  ]);
  const target = getOutOfRangePageRedirect("/admin/users", raw, users.page);
  if (target) redirect(target);
  return (
    <main className="mx-auto w-full max-w-7xl space-y-6 px-4 py-6 sm:px-6 lg:px-8">
      <PageHeader
        title="Users"
        description="Manage operator identities, roles, credentials, and access context."
        actions={
          permissions.has("user:add") && permissions.has("role:get") ? (
            <UserForm roles={roles} />
          ) : undefined
        }
      />
      <UserFilters
        canReadRoles={canReadRoles}
        search={query.search}
        roleId={roleId}
        roleName={selectedRole?.name}
      />
      <p className="text-muted-foreground text-xs font-semibold tracking-wider uppercase">
        Showing {users.data.length} of {users.page.total_items} users
      </p>
      {users.data.length ? (
        <section
          aria-label="User collection"
          className="grid grid-cols-1 gap-5 sm:grid-cols-2 xl:grid-cols-3"
        >
          {users.data.map((user) => (
            <UserCard
              key={user.id}
              user={user}
              role={roles.find((role) => role.id === user.role_id)}
            />
          ))}
        </section>
      ) : (
        <EmptyState
          title={query.search || roleId ? "No matching users" : "No users yet"}
          description={
            query.search || roleId
              ? "Clear or change the filters to broaden the results."
              : "Create the first managed account when a role is available."
          }
        />
      )}
      <Pagination
        page={users.page}
        pathname="/admin/users"
        searchParams={{
          limit: String(query.limit),
          search: query.search,
          role_id: roleId,
        }}
      />
    </main>
  );
}
