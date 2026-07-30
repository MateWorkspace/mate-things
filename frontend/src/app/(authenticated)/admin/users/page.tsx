import { Search } from "lucide-react";
import type { Metadata } from "next";
import Link from "next/link";
import { redirect } from "next/navigation";

import CollectionToolbar from "@/components/collection/CollectionToolbar";
import Pagination from "@/components/collection/Pagination";
import Button from "@/components/ui/button";
import Input from "@/components/ui/input";
import PageHeader from "@/components/ui/page-header";
import { EmptyState } from "@/components/ui/states";
import { listRoles } from "@/lib/api/roles";
import { listUsers } from "@/lib/api/users";
import {
  getOutOfRangePageRedirect,
  parsePageQuery,
} from "@/lib/collection-query";
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
  const [users, rolesResult] = await Promise.all([
    listUsers(query),
    permissions.has("role:get")
      ? listRoles({ limit: 48 })
      : Promise.resolve(null),
  ]);
  const target = getOutOfRangePageRedirect("/admin/users", raw, users.page);
  if (target) redirect(target);
  const roles = rolesResult?.data ?? [];
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
        <form action="/admin/users" className="flex flex-col gap-3 sm:flex-row">
          <input type="hidden" name="limit" value={query.limit} />
          <Input
            name="search"
            type="search"
            defaultValue={query.search}
            placeholder="Search name or username"
            aria-label="Search users"
          />
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
          title={query.search ? "No matching users" : "No users yet"}
          description={
            query.search
              ? "No accounts match the current search."
              : "Create the first managed account when a role is available."
          }
          action={
            query.search ? (
              <Link href="/admin/users" className="text-primary font-semibold">
                Clear search
              </Link>
            ) : undefined
          }
        />
      )}
      <Pagination
        page={users.page}
        pathname="/admin/users"
        searchParams={{ limit: String(query.limit), search: query.search }}
      />
    </main>
  );
}
