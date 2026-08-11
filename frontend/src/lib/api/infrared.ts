import "server-only";

import { apiFetch } from "@/lib/api/client";
import type { AuditFields, IdResponse } from "@/lib/api/types";

export type InfraredStateType = "ENUM" | "RANGE";

export interface InfraredDeviceTypeResponse extends AuditFields {
  id: string;
  name: string;
}

export interface InfraredStateResponse extends AuditFields {
  id: string;
  infrared_device_type_id: string;
  name: string;
  type: InfraredStateType;
}

export interface CreateInfraredStateRequest {
  name: string;
  type: InfraredStateType;
}

export async function listInfraredDeviceTypes(): Promise<
  InfraredDeviceTypeResponse[]
> {
  return apiFetch("/infrared/device-types");
}

export async function createInfraredDeviceType(
  name: string,
): Promise<IdResponse> {
  return apiFetch("/infrared/device-types", {
    method: "POST",
    body: { name },
  });
}

export async function deleteInfraredDeviceType(id: string): Promise<void> {
  return apiFetch(`/infrared/device-types/${id}`, { method: "DELETE" });
}

export async function listInfraredStates(
  deviceTypeId: string,
): Promise<InfraredStateResponse[]> {
  return apiFetch(`/infrared/device-types/${deviceTypeId}/states`);
}

export async function createInfraredState(
  deviceTypeId: string,
  request: CreateInfraredStateRequest,
): Promise<IdResponse> {
  return apiFetch(`/infrared/device-types/${deviceTypeId}/states`, {
    method: "POST",
    body: request,
  });
}

export async function deleteInfraredState(id: string): Promise<void> {
  return apiFetch(`/infrared/states/${id}`, { method: "DELETE" });
}
