import type { ActionState } from "@/lib/forms/action-state";
import type { LlmConnectionStatus } from "@/lib/api/llm-config";

export type LlmConfigActionState = ActionState<string>;

export const EMPTY_LLM_CONFIG_STATE: LlmConfigActionState = {
  status: "idle",
};

export type TestConnectionState =
  | { status: "idle" }
  | { status: "result"; connectionStatus: LlmConnectionStatus }
  | { status: "error"; message: string };

export const EMPTY_TEST_CONNECTION_STATE: TestConnectionState = {
  status: "idle",
};
