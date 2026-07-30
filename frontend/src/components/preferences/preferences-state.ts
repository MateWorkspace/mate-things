export interface PreferencesActionState {
  status: "idle" | "success" | "error";
  message?: string;
  fieldErrors?: Record<string, string>;
}

export const EMPTY_PREFERENCES_STATE: PreferencesActionState = {
  status: "idle",
};
