import { Search } from "lucide-react";
import type { Metadata } from "next";
import Link from "next/link";
import { redirect } from "next/navigation";

import CollectionToolbar from "@/components/collection/CollectionToolbar";
import Pagination from "@/components/collection/Pagination";
import RoleSearchCombobox from "@/components/roles/RoleSearchCombobox";
import Button from "@/components/ui/button";
import Input from "@/components/ui/input";
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
      <CollectionToolbar filterTitle="Filter users">
        <form
          action="/admin/users"
          className="flex flex-col gap-3 sm:flex-row sm:items-end"
        >
          <input type="hidden" name="limit" value={query.limit} />
          <label className="flex-1">
            <span className="text-foreground/70 mb-1.5 block text-xs font-semibold">
              Search
            </span>
            <Input
              name="search"
              type="search"
              defaultValue={query.search}
              placeholder="Search name or username"
              aria-label="Search users"
            />
          </label>
          {canReadRoles ? (
            <div className="sm:w-56">
              <RoleSearchCombobox
                name="role_id"
                defaultRoleId={roleId}
                defaultRoleName={selectedRole?.name}
              />
            </div>
          ) : null}
          <Button type="submit" className="gap-2">
            <Search className="size-4" aria-hidden="true" />
            Search
          </Button>
        </form>
      </CollectionToolbar>
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
              ? "No accounts match the current filters."
              : "Create the first managed account when a role is available."
          }
          action={
            query.search || roleId ? (
              <Link href="/admin/users" className="text-primary font-semibold">
                Clear filters
              </Link>
            ) : undefined
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
