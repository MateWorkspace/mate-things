"use server";

import { ApiError } from "@/lib/api/client";
import {
  testLlmConnection,
  updateLlmConfig,
  type LlmProvider,
} from "@/lib/api/llm-config";
import { requireSessionContext } from "@/lib/session";

import type { LlmConfigActionState, TestConnectionState } from "./state";

const PROVIDERS = new Set<LlmProvider>(["CLAUDE", "OPENAI"]);

function permissionDenied(): LlmConfigActionState {
  return {
    status: "error",
    title: "Permission denied",
    message: "You do not have permission to make this change.",
  };
}

function actionError(error: unknown): LlmConfigActionState {
  if (error instanceof ApiError) {
    return { status: "error", title: error.title, message: error.message };
  }

  return {
    status: "error",
    title: "Something went wrong",
    message: "Please try again.",
  };
}

export async function updateLlmConfigAction(
  _previousState: LlmConfigActionState,
  formData: FormData,
): Promise<LlmConfigActionState> {
  const session = await requireSessionContext();
  if (!session.permissions.has("llm_config:set")) {
    return permissionDenied();
  }

  const rawProvider = String(formData.get("provider") ?? "").trim();
  const provider = PROVIDERS.has(rawProvider as LlmProvider)
    ? (rawProvider as LlmProvider)
    : undefined;
  const model = String(formData.get("model") ?? "").trim();
  const apiKey = String(formData.get("api_key") ?? "");
  const apiKeyWasSet = formData.get("api_key_was_set") === "true";
  const baseUrl = String(formData.get("base_url") ?? "").trim();

  const fieldErrors: Record<string, string> = {};
  if (!provider) fieldErrors.provider = "Choose a provider.";
  if (!model) fieldErrors.model = "This field is required.";
  if (!apiKey && !apiKeyWasSet) {
    fieldErrors.api_key = "Required the first time you configure a provider.";
  }
  if (Object.keys(fieldErrors).length > 0) {
    return {
      status: "error",
      title: "Check the configuration",
      message: "Complete the required fields.",
      fieldErrors,
    };
  }

  try {
    await updateLlmConfig({
      provider: provider!,
      model,
      api_key: apiKey || undefined,
      base_url: baseUrl || undefined,
    });
  } catch (error) {
    return actionError(error);
  }

  return {
    status: "success",
    title: "LLM configuration saved",
    message: "The active provider, model, and credential have been updated.",
  };
}

export async function testLlmConnectionAction(
  _previousState: TestConnectionState,
  _formData: FormData,
): Promise<TestConnectionState> {
  const session = await requireSessionContext();
  if (!session.permissions.has("llm_config:set")) {
    return {
      status: "error",
      message: "You do not have permission to do this.",
    };
  }

  try {
    const result = await testLlmConnection();
    return { status: "result", connectionStatus: result.status };
  } catch (error) {
    return {
      status: "error",
      message: error instanceof ApiError ? error.message : "Please try again.",
    };
  }
}
