import "server-only";

import { apiFetch } from "@/lib/api/client";

export type LlmProvider = "CLAUDE" | "OPENAI" | "GEMINI";
export type LlmConnectionStatus = "CONNECTED" | "DISCONNECTED";

export interface LlmConfigResponse {
  provider: LlmProvider;
  model: string;
  base_url?: string;
  api_key_set: boolean;
  updated_at?: string;
  updated_by?: string;
}

export interface UpdateLlmConfigRequest {
  provider: LlmProvider;
  model: string;
  api_key?: string;
  base_url?: string;
}

/** Same shape as UpdateLlmConfigRequest, minus persistence — omitting
 * api_key falls back to whatever key is currently saved. */
export interface TestLlmConnectionWithConfigRequest {
  provider: LlmProvider;
  model: string;
  api_key?: string;
  base_url?: string;
}

export interface LlmConnectionStatusResponse {
  status: LlmConnectionStatus;
}

export async function getLlmConfig(): Promise<LlmConfigResponse> {
  return apiFetch("/admin/llm-config");
}

export async function updateLlmConfig(
  request: UpdateLlmConfigRequest,
): Promise<void> {
  return apiFetch("/admin/llm-config", { method: "PUT", body: request });
}

export async function testLlmConnection(): Promise<LlmConnectionStatusResponse> {
  return apiFetch("/admin/llm-config/test", { method: "POST" });
}

export async function testLlmConnectionWithConfig(
  request: TestLlmConnectionWithConfigRequest,
): Promise<LlmConnectionStatusResponse> {
  return apiFetch("/admin/llm-config/test-connection", {
    method: "POST",
    body: request,
  });
}
