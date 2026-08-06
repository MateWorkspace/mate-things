export interface ApiKeyActionState {
  status: "idle" | "success" | "error";
  title?: string;
  message?: string;
  fieldErrors?: Record<string, string>;
  key?: string;
}

export const EMPTY_API_KEY_STATE: ApiKeyActionState = { status: "idle" };
