import "server-only";

import { apiFetch, buildQuery } from "@/lib/api/client";
import type {
  AuditFields,
  IdResponse,
  PageDataResponse,
  PageQuery,
} from "@/lib/api/types";

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

export type InfraredRecordingState =
  | "DRAFT"
  | "CASES_GENERATING"
  | "RECORDING"
  | "ANALYZING"
  | "FUNCTION_GENERATING"
  | "TEST_CASES_GENERATING"
  | "TESTING"
  | "COMPLETED"
  | "FAILED";

export interface InfraredRecordSessionListItemResponse {
  id: string;
  recording_state: InfraredRecordingState;
  is_completed: boolean;
  brand: string;
  model: string;
  device_type_name: string;
  created_at: string;
}

export interface ListInfraredRecordSessionsQuery extends PageQuery {
  recording_state?: InfraredRecordingState;
  infrared_device_type_id?: string;
  created_at_start?: string;
  created_at_end?: string;
}

export async function listInfraredRecordSessions(
  query: ListInfraredRecordSessionsQuery = {},
): Promise<PageDataResponse<InfraredRecordSessionListItemResponse>> {
  return apiFetch(`/infrared/record-sessions${buildQuery(query)}`);
}

export interface InfraredRecordSessionResponse extends AuditFields {
  id: string;
  node_id: string;
  infrared_device_id: string;
  recording_state: InfraredRecordingState;
  current_record_case_id?: string;
  is_completed: boolean;
}

export async function getInfraredRecordSession(
  id: string,
): Promise<InfraredRecordSessionResponse> {
  return apiFetch(`/infrared/record-sessions/${id}`);
}

export interface InfraredDeviceResponse extends AuditFields {
  id: string;
  infrared_device_type_id: string;
  brand: string;
  model: string;
}

export async function getInfraredDevice(
  id: string,
): Promise<InfraredDeviceResponse> {
  return apiFetch(`/infrared/devices/${id}`);
}

export interface StartRecordSessionDefinition {
  infrared_state_id: string;
  options?: string[];
  minimum?: number;
  maximum?: number;
  step?: number;
}

export interface StartRecordSessionRequest {
  node_id: string;
  infrared_device_type_id: string;
  brand: string;
  model: string;
  definitions: StartRecordSessionDefinition[];
}

export async function startInfraredRecordSession(
  request: StartRecordSessionRequest,
): Promise<IdResponse> {
  return apiFetch("/infrared/record-sessions", {
    method: "POST",
    body: request,
  });
}

export interface InfraredStateDeviceRecordStateResponse {
  id: string;
  infrared_state_id: string;
  state_value: string;
  created_at: string;
  deleted_at?: string;
  deleted_by?: string;
}

export interface InfraredStateDeviceRecordRawResponse {
  id: string;
  status: "CAPTURED" | "ACCEPTED" | "DISCARDED";
  discarded_reason?: string;
  created_at: string;
  deleted_at?: string;
  deleted_by?: string;
}

export interface InfraredStateDeviceRecordCaseResponse {
  id: string;
  step: number;
  description: string;
  status: "PENDING" | "ACTIVE" | "ACCEPTED";
  states: InfraredStateDeviceRecordStateResponse[];
  raw: InfraredStateDeviceRecordRawResponse[];
  created_at: string;
  deleted_at?: string;
  deleted_by?: string;
}

export async function listInfraredRecordSessionCases(
  sessionId: string,
): Promise<InfraredStateDeviceRecordCaseResponse[]> {
  return apiFetch(`/infrared/record-sessions/${sessionId}/cases`);
}

export async function acceptInfraredRecordRaw(
  caseId: string,
  rawId: string,
): Promise<void> {
  return apiFetch(`/infrared/record-cases/${caseId}/raw/${rawId}/accept`, {
    method: "POST",
  });
}

export async function discardInfraredRecordRaw(
  caseId: string,
  rawId: string,
  reason: string,
): Promise<void> {
  return apiFetch(`/infrared/record-cases/${caseId}/raw/${rawId}/discard`, {
    method: "POST",
    body: { reason },
  });
}

export async function retryInfraredRecordCase(caseId: string): Promise<void> {
  return apiFetch(`/infrared/record-cases/${caseId}/retry`, {
    method: "POST",
  });
}

export async function setInfraredRecordSessionCurrentCase(
  sessionId: string,
  caseId: string,
): Promise<void> {
  return apiFetch(
    `/infrared/record-sessions/${sessionId}/cases/${caseId}/current`,
    { method: "POST" },
  );
}

export interface InfraredStateCoderResponse {
  id: string;
  encoder_source: string;
  decoder_source: string;
  summary_readme: string;
  detail_readme: string;
  status: "UNVERIFIED" | "ACTIVE" | "SUPERSEDED";
  created_at: string;
  deleted_at?: string;
  deleted_by?: string;
}

export async function getInfraredRecordSessionCoder(
  sessionId: string,
): Promise<InfraredStateCoderResponse> {
  return apiFetch(`/infrared/record-sessions/${sessionId}/coder`);
}

export interface InfraredTestCaseStateResponse {
  id: string;
  infrared_state_id: string;
  state_value: string;
  created_at: string;
  deleted_at?: string;
  deleted_by?: string;
}

export interface InfraredTestCaseResponse {
  id: string;
  step: number;
  description: string;
  status: "PENDING" | "PASSED" | "FAILED";
  states: InfraredTestCaseStateResponse[];
  created_at: string;
  deleted_at?: string;
  deleted_by?: string;
}

export async function listInfraredRecordSessionTestCases(
  sessionId: string,
): Promise<InfraredTestCaseResponse[]> {
  return apiFetch(`/infrared/record-sessions/${sessionId}/test-cases`);
}

export async function transmitInfraredTestCase(
  testCaseId: string,
): Promise<void> {
  return apiFetch(`/infrared/test-cases/${testCaseId}/transmit`, {
    method: "POST",
  });
}

export async function recordInfraredTestCaseResult(
  testCaseId: string,
  passed: boolean,
): Promise<void> {
  return apiFetch(`/infrared/test-cases/${testCaseId}/result`, {
    method: "POST",
    body: { passed },
  });
}

export async function deleteInfraredRecordSession(id: string): Promise<void> {
  return apiFetch(`/infrared/record-sessions/${id}`, { method: "DELETE" });
}

export async function deleteInfraredRecordCase(caseId: string): Promise<void> {
  return apiFetch(`/infrared/record-cases/${caseId}`, { method: "DELETE" });
}

export async function deleteInfraredRecordState(id: string): Promise<void> {
  return apiFetch(`/infrared/record-states/${id}`, { method: "DELETE" });
}

export async function deleteInfraredRecordRaw(
  caseId: string,
  rawId: string,
): Promise<void> {
  return apiFetch(`/infrared/record-cases/${caseId}/raw/${rawId}`, {
    method: "DELETE",
  });
}

export async function deleteInfraredStateCoder(id: string): Promise<void> {
  return apiFetch(`/infrared/state-coders/${id}`, { method: "DELETE" });
}

export async function deleteInfraredTestCase(id: string): Promise<void> {
  return apiFetch(`/infrared/test-cases/${id}`, { method: "DELETE" });
}

export async function deleteInfraredTestCaseState(
  testCaseId: string,
  stateId: string,
): Promise<void> {
  return apiFetch(`/infrared/test-cases/${testCaseId}/states/${stateId}`, {
    method: "DELETE",
  });
}
