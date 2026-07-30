export interface UserActionState {
  status: "idle" | "success" | "error";
  title?: string;
  message?: string;
  fieldErrors?: Record<string, string>;
}

export const EMPTY_USER_STATE: UserActionState = { status: "idle" };
