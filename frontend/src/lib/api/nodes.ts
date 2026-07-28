"use server";

import { apiFetch, buildQuery } from "@/lib/api/client";
import type { AuditFields, PageDataResponse, PageQuery } from "@/lib/api/types";

export interface NodeResponse extends AuditFields {
  id: string;
  node_class_id: string;
  device_id: string;
  device_info: string;
  name: string;
  firmware_id: string;
  description: string;
  is_connected: boolean;
  preferences: Record<string, unknown>;
}

export interface UpdateNodeRequest {
  node_class_id?: string;
  device_id?: string;
  name?: string;
  firmware_id?: string;
  description?: string;
}

export interface ListNodesQuery extends PageQuery {
  node_class_id?: string;
  firmware_id?: string;
}

// Nodes are created by the device itself over MQTT on registration, never
// via this HTTP API - there's no createNode here to mirror; the backend
// has no POST /nodes route.

export async function listNodes(
  query: ListNodesQuery = {},
): Promise<PageDataResponse<NodeResponse>> {
  return apiFetch(`/nodes${buildQuery(query)}`);
}

export async function getNodeByDeviceId(
  deviceId: string,
): Promise<NodeResponse> {
  return apiFetch(`/nodes/by-device/${encodeURIComponent(deviceId)}`);
}

export async function getNodeById(id: string): Promise<NodeResponse> {
  return apiFetch(`/nodes/${id}`);
}

export async function updateNode(
  id: string,
  request: UpdateNodeRequest,
): Promise<void> {
  return apiFetch(`/nodes/${id}`, { method: "PATCH", body: request });
}

export async function updateNodeFirmware(
  id: string,
  firmwareId: string,
): Promise<void> {
  return apiFetch(`/nodes/${id}/firmware`, {
    method: "PATCH",
    body: { firmware_id: firmwareId },
  });
}

export async function deleteNode(id: string): Promise<void> {
  return apiFetch(`/nodes/${id}`, { method: "DELETE" });
}
