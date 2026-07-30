"use server";

import { apiFetch } from "@/lib/api/client";

export interface NodeConfigValueResponse {
  key: string;
  value: string;
  updated_at?: string;
}

export async function getNodeConfig(
  id: string,
): Promise<NodeConfigValueResponse[]> {
  return apiFetch(`/nodes/${id}/config`);
}

export async function setNodeConfig(
  id: string,
  key: string,
  value: string,
): Promise<void> {
  return apiFetch(`/nodes/${id}/config`, {
    method: "PUT",
    body: { key, value },
  });
}
