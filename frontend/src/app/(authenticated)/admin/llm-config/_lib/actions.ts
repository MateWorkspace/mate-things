"use server";

import { ApiError } from "@/lib/api/client";
import {
  testLlmConnection,
  testLlmConnectionWithConfig,
  updateLlmConfig,
  type LlmProvider,
} from "@/lib/api/llm-config";
import { requireSessionContext } from "@/lib/session";

import type { LlmConfigActionState, TestConnectionState } from "./state";

const PROVIDERS = new Set<LlmProvider>(["CLAUDE", "OPENAI", "GEMINI"]);

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

interface ParsedLlmConfigFields {
  provider?: LlmProvider;
  model: string;
  apiKey: string;
  apiKeyWasSet: boolean;
  baseUrl: string;
}

/** Reads the fields both the Save and Test buttons submit from the same form. */
function parseLlmConfigFields(formData: FormData): ParsedLlmConfigFields {
  const rawProvider = String(formData.get("provider") ?? "").trim();
  const provider = PROVIDERS.has(rawProvider as LlmProvider)
    ? (rawProvider as LlmProvider)
    : undefined;

  return {
    provider,
    model: String(formData.get("model") ?? "").trim(),
    apiKey: String(formData.get("api_key") ?? ""),
    apiKeyWasSet: formData.get("api_key_was_set") === "true",
    baseUrl: String(formData.get("base_url") ?? "").trim(),
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

  const { provider, model, apiKey, apiKeyWasSet, baseUrl } =
    parseLlmConfigFields(formData);

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

/**
 * Tests the provider/model/base_url/api_key currently typed into the form,
 * without saving anything - some base URLs and keys don't support every
 * model a provider offers, so this lets that be checked before Save.
 */
export async function testLlmConfigFormAction(
  _previousState: TestConnectionState,
  formData: FormData,
): Promise<TestConnectionState> {
  const session = await requireSessionContext();
  if (!session.permissions.has("llm_config:set")) {
    return {
      status: "error",
      message: "You do not have permission to do this.",
    };
  }

  const { provider, model, apiKey, apiKeyWasSet, baseUrl } =
    parseLlmConfigFields(formData);

  if (!provider || !model || (!apiKey && !apiKeyWasSet)) {
    return {
      status: "error",
      message:
        "Choose a provider and model (and an API key, unless one is already saved) before testing.",
    };
  }

  try {
    const result = await testLlmConnectionWithConfig({
      provider,
      model,
      api_key: apiKey || undefined,
      base_url: baseUrl || undefined,
    });
    return { status: "result", connectionStatus: result.status };
  } catch (error) {
    return {
      status: "error",
      message: error instanceof ApiError ? error.message : "Please try again.",
    };
  }
}
