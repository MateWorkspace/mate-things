"use server";

import { ApiError } from "@/lib/api/client";
import {
  acceptInfraredRecordRaw,
  deleteInfraredRecordCase,
  deleteInfraredRecordRaw,
  deleteInfraredRecordSession,
  deleteInfraredRecordState,
  deleteInfraredStateCoder,
  deleteInfraredTestCase,
  deleteInfraredTestCaseState,
  discardInfraredRecordRaw,
  recordInfraredTestCaseResult,
  retryInfraredRecordCase,
  transmitInfraredTestCase,
} from "@/lib/api/infrared";
import { requireSessionContext } from "@/lib/session";

import type { RecordSessionActionState } from "./state";

function permissionDenied(): RecordSessionActionState {
  return {
    status: "error",
    title: "Permission denied",
    message: "You do not have permission to make this change.",
  };
}

function actionError(error: unknown): RecordSessionActionState {
  if (error instanceof ApiError) {
    return { status: "error", title: error.title, message: error.message };
  }
  return {
    status: "error",
    title: "Something went wrong",
    message: "Please try again.",
  };
}

export async function acceptRawAction(
  _previousState: RecordSessionActionState,
  formData: FormData,
): Promise<RecordSessionActionState> {
  const { permissions } = await requireSessionContext();
  if (!permissions.has("infrared_record_session:set")) {
    return permissionDenied();
  }
  const caseId = String(formData.get("case_id") ?? "").trim();
  const rawId = String(formData.get("raw_id") ?? "").trim();
  try {
    await acceptInfraredRecordRaw(caseId, rawId);
  } catch (error) {
    return actionError(error);
  }
  return { status: "success", title: "Capture accepted", message: "" };
}

export async function discardRawAction(
  _previousState: RecordSessionActionState,
  formData: FormData,
): Promise<RecordSessionActionState> {
  const { permissions } = await requireSessionContext();
  if (!permissions.has("infrared_record_session:set")) {
    return permissionDenied();
  }
  const caseId = String(formData.get("case_id") ?? "").trim();
  const rawId = String(formData.get("raw_id") ?? "").trim();
  const reason =
    String(formData.get("reason") ?? "").trim() || "user discarded";
  try {
    await discardInfraredRecordRaw(caseId, rawId, reason);
  } catch (error) {
    return actionError(error);
  }
  return { status: "success", title: "Capture discarded", message: "" };
}

export async function retryCaseAction(
  _previousState: RecordSessionActionState,
  formData: FormData,
): Promise<RecordSessionActionState> {
  const { permissions } = await requireSessionContext();
  if (!permissions.has("infrared_record_session:set")) {
    return permissionDenied();
  }
  const caseId = String(formData.get("case_id") ?? "").trim();
  try {
    await retryInfraredRecordCase(caseId);
  } catch (error) {
    return actionError(error);
  }
  return { status: "success", title: "Case reset for retake", message: "" };
}

export async function transmitTestCaseAction(
  _previousState: RecordSessionActionState,
  formData: FormData,
): Promise<RecordSessionActionState> {
  const { permissions } = await requireSessionContext();
  if (!permissions.has("infrared_record_session:set")) {
    return permissionDenied();
  }
  const testCaseId = String(formData.get("test_case_id") ?? "").trim();
  try {
    await transmitInfraredTestCase(testCaseId);
  } catch (error) {
    return actionError(error);
  }
  return { status: "success", title: "Transmitted", message: "" };
}

export async function recordTestCaseResultAction(
  _previousState: RecordSessionActionState,
  formData: FormData,
): Promise<RecordSessionActionState> {
  const { permissions } = await requireSessionContext();
  if (!permissions.has("infrared_record_session:set")) {
    return permissionDenied();
  }
  const testCaseId = String(formData.get("test_case_id") ?? "").trim();
  const passed = formData.get("passed") === "true";
  try {
    await recordInfraredTestCaseResult(testCaseId, passed);
  } catch (error) {
    return actionError(error);
  }
  return { status: "success", title: "Result recorded", message: "" };
}

export async function deleteRecordSessionAction(
  _previousState: RecordSessionActionState,
  formData: FormData,
): Promise<RecordSessionActionState> {
  const { permissions } = await requireSessionContext();
  if (!permissions.has("infrared_record_session:delete")) {
    return permissionDenied();
  }
  const id = String(formData.get("id") ?? "").trim();
  try {
    await deleteInfraredRecordSession(id);
  } catch (error) {
    return actionError(error);
  }
  return { status: "success", title: "Session deleted", message: "" };
}

export async function deleteRecordCaseAction(
  _previousState: RecordSessionActionState,
  formData: FormData,
): Promise<RecordSessionActionState> {
  const { permissions } = await requireSessionContext();
  if (!permissions.has("infrared_record_session:delete")) {
    return permissionDenied();
  }
  const caseId = String(formData.get("case_id") ?? "").trim();
  try {
    await deleteInfraredRecordCase(caseId);
  } catch (error) {
    return actionError(error);
  }
  return { status: "success", title: "Case deleted", message: "" };
}

export async function deleteRecordStateAction(
  _previousState: RecordSessionActionState,
  formData: FormData,
): Promise<RecordSessionActionState> {
  const { permissions } = await requireSessionContext();
  if (!permissions.has("infrared_record_session:delete")) {
    return permissionDenied();
  }
  const id = String(formData.get("id") ?? "").trim();
  try {
    await deleteInfraredRecordState(id);
  } catch (error) {
    return actionError(error);
  }
  return { status: "success", title: "State deleted", message: "" };
}

export async function deleteRecordRawAction(
  _previousState: RecordSessionActionState,
  formData: FormData,
): Promise<RecordSessionActionState> {
  const { permissions } = await requireSessionContext();
  if (!permissions.has("infrared_record_session:delete")) {
    return permissionDenied();
  }
  const caseId = String(formData.get("case_id") ?? "").trim();
  const rawId = String(formData.get("raw_id") ?? "").trim();
  try {
    await deleteInfraredRecordRaw(caseId, rawId);
  } catch (error) {
    return actionError(error);
  }
  return { status: "success", title: "Capture deleted", message: "" };
}

export async function deleteStateCoderAction(
  _previousState: RecordSessionActionState,
  formData: FormData,
): Promise<RecordSessionActionState> {
  const { permissions } = await requireSessionContext();
  if (!permissions.has("infrared_record_session:delete")) {
    return permissionDenied();
  }
  const id = String(formData.get("id") ?? "").trim();
  try {
    await deleteInfraredStateCoder(id);
  } catch (error) {
    return actionError(error);
  }
  return { status: "success", title: "Coder deleted", message: "" };
}

export async function deleteTestCaseAction(
  _previousState: RecordSessionActionState,
  formData: FormData,
): Promise<RecordSessionActionState> {
  const { permissions } = await requireSessionContext();
  if (!permissions.has("infrared_record_session:delete")) {
    return permissionDenied();
  }
  const id = String(formData.get("id") ?? "").trim();
  try {
    await deleteInfraredTestCase(id);
  } catch (error) {
    return actionError(error);
  }
  return { status: "success", title: "Test case deleted", message: "" };
}

export async function deleteTestCaseStateAction(
  _previousState: RecordSessionActionState,
  formData: FormData,
): Promise<RecordSessionActionState> {
  const { permissions } = await requireSessionContext();
  if (!permissions.has("infrared_record_session:delete")) {
    return permissionDenied();
  }
  const testCaseId = String(formData.get("test_case_id") ?? "").trim();
  const stateId = String(formData.get("state_id") ?? "").trim();
  try {
    await deleteInfraredTestCaseState(testCaseId, stateId);
  } catch (error) {
    return actionError(error);
  }
  return { status: "success", title: "Test case state deleted", message: "" };
}
