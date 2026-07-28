"use server";

import { apiFetch } from "@/lib/api/client";

export interface DispatchOtaRequest {
  firmware_id: string;
  firmware_url: string;
}

/**
 * Both dispatch routes live under /nodes/... in the backend's URL space but
 * are tagged "OTA" in swagger - kept here, not in nodes.ts, to match the
 * tag grouping.
 */
export async function dispatchOtaByNodeId(
  nodeId: string,
  request: DispatchOtaRequest,
): Promise<void> {
  return apiFetch(`/nodes/${nodeId}/ota`, { method: "POST", body: request });
}

export async function dispatchOtaByDeviceId(
  deviceId: string,
  request: DispatchOtaRequest,
): Promise<void> {
  return apiFetch(`/nodes/by-device/${encodeURIComponent(deviceId)}/ota`, {
    method: "POST",
    body: request,
  });
}
