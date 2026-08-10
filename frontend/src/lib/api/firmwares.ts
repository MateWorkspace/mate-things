import "server-only";

import { apiFetch, apiRequest, buildQuery } from "@/lib/api/client";
import { collectAllPages } from "@/lib/api/collect-all-pages";
import type { AuditFields, PageDataResponse, PageQuery } from "@/lib/api/types";

export interface FirmwareResponse extends AuditFields {
  id: string;
  node_class_id: string;
  name: string;
  size: number;
  checksum: string;
  binary_path: string;
  preferences: Record<string, unknown>;
}

export interface FirmwareCreateResponse {
  id: string;
  size: number;
  checksum: string;
  binary_path: string;
}

export interface FirmwareBinaryStatResponse {
  binary_path: string;
  size: number;
  checksum: string;
}

export interface FirmwareConfigParameter {
  key: string;
  value_type: string;
}

export type FirmwareConfigParameterResponse = FirmwareConfigParameter;

export type FirmwareConfigSchemaItem = FirmwareConfigParameter;

export type ReplaceFirmwareBinaryInput = {
  id: string;
  binary: File;
  configSchema?: FirmwareConfigParameter[];
};

export interface UpdateFirmwareRequest {
  node_class_id?: string;
  name?: string;
}

export interface ListFirmwaresQuery extends PageQuery {
  node_class_id?: string;
}

export async function listFirmwares(
  query: ListFirmwaresQuery = {},
): Promise<PageDataResponse<FirmwareResponse>> {
  return apiFetch(`/firmwares${buildQuery(query)}`);
}

const FIRMWARE_OPTION_PAGE_LIMIT = 48;

export async function listAllFirmwares(): Promise<FirmwareResponse[]> {
  return collectAllPages({
    fetchPage: async (page) => {
      const result = await listFirmwares({
        page,
        limit: FIRMWARE_OPTION_PAGE_LIMIT,
      });
      return {
        data: result.data,
        page: result.page.page,
        limit: result.page.limit,
        total: result.page.total_items,
      };
    },
    keyOf: (firmware) => firmware.id,
  });
}

export async function listFirmwaresByNodeClassId(
  nodeClassId: string,
  query: PageQuery = {},
): Promise<PageDataResponse<FirmwareResponse>> {
  return apiFetch(`/node-classes/${nodeClassId}/firmwares${buildQuery(query)}`);
}

export async function listAvailableFirmwaresByNodeId(
  nodeId: string,
  query: PageQuery = {},
): Promise<PageDataResponse<FirmwareResponse>> {
  return apiFetch(`/nodes/${nodeId}/firmwares/available${buildQuery(query)}`);
}

export async function getFirmwareByName(
  name: string,
): Promise<FirmwareResponse> {
  return apiFetch(`/firmwares/by-name/${encodeURIComponent(name)}`);
}

export async function getFirmwareById(id: string): Promise<FirmwareResponse> {
  return apiFetch(`/firmwares/${id}`);
}

export async function getFirmwareConfigParameters(
  id: string,
): Promise<FirmwareConfigParameterResponse[]> {
  return apiFetch(`/firmwares/${id}/config-parameters`);
}

/**
 * Creates a firmware row and uploads its binary in one call - the backend
 * expects multipart/form-data (node_class_id, name, file), not JSON.
 */
export async function createFirmware(
  nodeClassId: string,
  name: string,
  file: File | Blob,
  configSchema: FirmwareConfigSchemaItem[],
): Promise<FirmwareCreateResponse> {
  const body = new FormData();
  body.set("node_class_id", nodeClassId);
  body.set("name", name);
  body.set("file", file);
  body.set("config_schema", JSON.stringify(configSchema));

  return apiFetch("/firmwares", { method: "POST", body });
}

export async function updateFirmware(
  id: string,
  request: UpdateFirmwareRequest,
): Promise<void> {
  return apiFetch(`/firmwares/${id}`, { method: "PATCH", body: request });
}

/** Replaces an existing firmware row's binary content. */
export async function replaceFirmwareBinary(
  input: ReplaceFirmwareBinaryInput,
): Promise<FirmwareBinaryStatResponse> {
  const body = new FormData();
  body.set("file", input.binary);
  if (input.configSchema !== undefined) {
    body.append("config_schema", JSON.stringify(input.configSchema));
  }

  return apiFetch(`/firmwares/${input.id}/binary`, { method: "PUT", body });
}

export async function getFirmwareBinaryStatByName(
  name: string,
): Promise<FirmwareBinaryStatResponse> {
  return apiFetch(`/firmwares/by-name/${encodeURIComponent(name)}/binary/stat`);
}

/** True if a firmware binary with this name exists (HEAD request, no body). */
export async function firmwareBinaryExistsByName(
  name: string,
): Promise<boolean> {
  const response = await apiRequest(
    `/firmwares/by-name/${encodeURIComponent(name)}/binary`,
    { method: "HEAD" },
  );
  return response.ok;
}

/**
 * Resolves the short-lived MinIO presigned download URL for a firmware
 * binary, by id. The backend responds 302 with a Location header rather
 * than a JSON body or the binary itself - apiFetch would either throw
 * (JSON parse of a redirect) or silently follow the redirect and download
 * the binary, neither of which is what a caller wants here, so this uses
 * apiRequest directly and reads the Location header.
 */
export async function getFirmwareBinaryUrlById(id: string): Promise<string> {
  const response = await apiRequest(`/firmwares/${id}/binary`);
  const location = response.headers.get("Location");

  if (!location) {
    throw new Error(`No Location header on firmware binary redirect for ${id}`);
  }

  return location;
}

export async function getFirmwareBinaryUrlByName(
  name: string,
): Promise<string> {
  const response = await apiRequest(
    `/firmwares/by-name/${encodeURIComponent(name)}/binary`,
  );
  const location = response.headers.get("Location");

  if (!location) {
    throw new Error(
      `No Location header on firmware binary redirect for ${name}`,
    );
  }

  return location;
}

export async function deleteFirmware(
  id: string,
  expectedName: string,
): Promise<void> {
  return apiFetch(`/firmwares/${id}`, {
    method: "DELETE",
    body: { expected_name: expectedName },
  });
}
