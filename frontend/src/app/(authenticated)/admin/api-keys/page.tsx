import type { Metadata } from "next";
import { redirect } from "next/navigation";

import Pagination from "@/components/collection/Pagination";
import PageHeader from "@/components/ui/page-header";
import { EmptyState } from "@/components/ui/states";
import { listApiKeys, type ApiKeyStatus } from "@/lib/api/api-keys";
import {
  getOutOfRangePageRedirect,
  parsePageQuery,
} from "@/lib/collection-query";
import { requirePermission } from "@/lib/session";

import ApiKeyFilters from "./_components/ApiKeyFilters";
import ApiKeyTable from "./_components/ApiKeyTable";
import GenerateApiKeyDialog from "./_components/GenerateApiKeyDialog";

export const metadata: Metadata = { title: "API Keys — Mate Things" };
type RawSearchParams = Record<string, string | string[] | undefined>;

const STATUSES = new Set<ApiKeyStatus>(["active", "inactive", "any"]);

export default async function ApiKeysPage({
  searchParams,
}: {
  searchParams: Promise<RawSearchParams>;
}) {
  const [{ permissions }, raw] = await Promise.all([
    requirePermission("api_key:get"),
    searchParams,
  ]);
  const query = parsePageQuery(raw);
  const rawStatus = typeof raw.status === "string" ? raw.status : "";
  const status = STATUSES.has(rawStatus as ApiKeyStatus)
    ? (rawStatus as ApiKeyStatus)
    : "any";

  const apiKeys = await listApiKeys({ ...query, status });
  const target = getOutOfRangePageRedirect(
    "/admin/api-keys",
    raw,
    apiKeys.page,
  );
  if (target) {
    redirect(target);
  }

  return (
    <main className="mx-auto w-full max-w-7xl space-y-6 px-4 py-6 sm:px-6 lg:px-8">
      <PageHeader
        title="API Keys"
        description="Manage per-user API keys used for app-to-app authentication."
        actions={
          permissions.has("api_key:add") ? <GenerateApiKeyDialog /> : undefined
        }
      />
      <ApiKeyFilters search={query.search} status={status} />
      <p className="text-muted-foreground text-xs font-semibold tracking-wider uppercase">
        Showing {apiKeys.data.length} of {apiKeys.page.total_items} API keys
      </p>
      {apiKeys.data.length ? (
        <ApiKeyTable
          apiKeys={apiKeys.data}
          canManage={permissions.has("api_key:set")}
          canDelete={permissions.has("api_key:remove")}
        />
      ) : (
        <EmptyState
          title={
            query.search || status !== "any"
              ? "No matching API keys"
              : "No API keys yet"
          }
          description={
            query.search || status !== "any"
              ? "No keys match the current filters."
              : "Generate a key for a user to get started."
          }
        />
      )}
      <Pagination
        page={apiKeys.page}
        pathname="/admin/api-keys"
        searchParams={{
          limit: String(query.limit),
          search: query.search,
          status: status === "any" ? undefined : status,
        }}
      />
    </main>
  );
}
