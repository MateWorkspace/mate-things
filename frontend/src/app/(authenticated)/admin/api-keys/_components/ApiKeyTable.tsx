import LocalDateTime from "@/components/ui/local-date-time";
import StatusBadge, { type StatusVariant } from "@/components/ui/status-badge";
import type { ApiKeyResponse } from "@/lib/api/api-keys";

import DeleteApiKeyDialog from "./DeleteApiKeyDialog";
import RegenerateApiKeyDialog from "./RegenerateApiKeyDialog";
import RevokeApiKeyDialog from "./RevokeApiKeyDialog";

interface ApiKeyTableProps {
  apiKeys: readonly ApiKeyResponse[];
  canManage: boolean;
  canDelete: boolean;
}

function statusOf(
  apiKey: ApiKeyResponse,
): { label: string; variant: StatusVariant } {
  if (apiKey.revoked_at) {
    return { label: "Revoked", variant: "neutral" };
  }
  if (apiKey.expires_at && new Date(apiKey.expires_at) <= new Date()) {
    return { label: "Expired", variant: "critical" };
  }
  return { label: "Active", variant: "success" };
}

export default function ApiKeyTable({
  apiKeys,
  canManage,
  canDelete,
}: ApiKeyTableProps) {
  return (
    <div className="border-border overflow-x-auto rounded-2xl border">
      <table className="w-full text-left text-sm">
        <thead className="bg-muted text-muted-foreground text-xs font-semibold tracking-wider uppercase">
          <tr>
            <th scope="col" className="px-4 py-3 whitespace-nowrap">
              User
            </th>
            <th scope="col" className="px-4 py-3 whitespace-nowrap">
              Key
            </th>
            <th scope="col" className="px-4 py-3 whitespace-nowrap">
              Status
            </th>
            <th scope="col" className="px-4 py-3 whitespace-nowrap">
              Expires
            </th>
            <th scope="col" className="px-4 py-3 whitespace-nowrap">
              Created
            </th>
            {canManage || canDelete ? (
              <th scope="col" className="px-4 py-3 whitespace-nowrap">
                Actions
              </th>
            ) : null}
          </tr>
        </thead>
        <tbody className="divide-border divide-y">
          {apiKeys.map((apiKey) => {
            const status = statusOf(apiKey);

            return (
              <tr
                key={apiKey.id}
                className="hover:bg-highlight/20 transition-colors"
              >
                <td className="px-4 py-3 whitespace-nowrap">
                  {apiKey.user_name}{" "}
                  <span className="text-muted-foreground">
                    @{apiKey.user_username}
                  </span>
                </td>
                <td className="px-4 py-3 font-mono text-xs whitespace-nowrap">
                  mate_…{apiKey.key_last_four}
                </td>
                <td className="px-4 py-3 whitespace-nowrap">
                  <StatusBadge variant={status.variant}>
                    {status.label}
                  </StatusBadge>
                </td>
                <td className="px-4 py-3 whitespace-nowrap">
                  {apiKey.expires_at ? (
                    <LocalDateTime value={apiKey.expires_at} variant="date" />
                  ) : (
                    "Never"
                  )}
                </td>
                <td className="px-4 py-3 whitespace-nowrap">
                  <LocalDateTime value={apiKey.created_at} variant="date" />
                </td>
                {canManage || canDelete ? (
                  <td className="px-4 py-3 whitespace-nowrap">
                    <div className="flex flex-wrap gap-2">
                      {canManage ? (
                        <RegenerateApiKeyDialog apiKey={apiKey} />
                      ) : null}
                      {canManage && !apiKey.revoked_at ? (
                        <RevokeApiKeyDialog apiKey={apiKey} />
                      ) : null}
                      {canDelete ? (
                        <DeleteApiKeyDialog apiKey={apiKey} />
                      ) : null}
                    </div>
                  </td>
                ) : null}
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}
