export interface AccessActionState {
  status: "idle" | "success" | "error";
  title?: string;
  message?: string;
  fieldErrors?: Record<string, string>;
}

export const EMPTY_ACCESS_STATE: AccessActionState = { status: "idle" };
