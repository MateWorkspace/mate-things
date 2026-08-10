import "server-only";

import { apiFetch, buildQuery } from "@/lib/api/client";
import { collectAllPages } from "@/lib/api/collect-all-pages";
import type {
  AuditFields,
  IdResponse,
  PageDataResponse,
  PageQuery,
} from "@/lib/api/types";

export interface PayloadSchemaResponse extends AuditFields {
  id: string;
  name: string;
  version: number;
  definition: Record<string, unknown>;
  valid_from: string;
  valid_to?: string;
  preferences: Record<string, unknown>;
}

export interface CreatePayloadSchemaRequest {
  name: string;
  version: number;
  definition: Record<string, unknown>;
  valid_from?: string;
  valid_to?: string;
}

export interface UpdatePayloadSchemaRequest {
  name?: string;
  version?: number;
  definition?: Record<string, unknown>;
  valid_from?: string;
  valid_to?: string;
}

export interface ListPayloadSchemasQuery extends PageQuery {
  /** ISO 8601 timestamp - only schemas valid at this instant are returned. */
  valid_at?: string;
}

export async function listPayloadSchemas(
  query: ListPayloadSchemasQuery = {},
): Promise<PageDataResponse<PayloadSchemaResponse>> {
  return apiFetch(`/admin/payload-schemas${buildQuery(query)}`);
}

export async function listAllPayloadSchemas(
  query: Omit<ListPayloadSchemasQuery, "page"> = {},
): Promise<PayloadSchemaResponse[]> {
  return collectAllPages({
    fetchPage: async (page) => {
      const result = await listPayloadSchemas({ ...query, page });
      return {
        data: result.data,
        page: result.page.page,
        limit: result.page.limit,
        total: result.page.total_items,
      };
    },
    keyOf: (schema) => schema.id,
  });
}

export async function getPayloadSchemaById(
  id: string,
): Promise<PayloadSchemaResponse> {
  return apiFetch(`/admin/payload-schemas/${id}`);
}

export async function createPayloadSchema(
  request: CreatePayloadSchemaRequest,
): Promise<IdResponse> {
  return apiFetch("/admin/payload-schemas", {
    method: "POST",
    body: request,
  });
}

export async function updatePayloadSchema(
  id: string,
  request: UpdatePayloadSchemaRequest,
): Promise<void> {
  return apiFetch(`/admin/payload-schemas/${id}`, {
    method: "PATCH",
    body: request,
  });
}

export async function deletePayloadSchema(id: string): Promise<void> {
  return apiFetch(`/admin/payload-schemas/${id}`, { method: "DELETE" });
}
