import { describe, expect, it } from "vitest";

import { INITIAL_ACTION_STATE, type ActionState } from "./action-state";

describe("ActionState", () => {
  it("provides one idle state for every action form", () => {
    expect(INITIAL_ACTION_STATE).toEqual({ status: "idle" });
  });

  it("retains typed field errors and feature-specific payload extensions", () => {
    type ApiKeyState = ActionState<"user_id" | "expires_at"> & {
      key?: string;
    };
    type ScopedDeleteState = ActionState & { count?: number };
    type DispatchState = ActionState & { executionId?: string };

    const apiKeyState: ApiKeyState = {
      status: "success",
      key: "secret-once",
    };
    const deleteState: ScopedDeleteState = { status: "success", count: 4 };
    const dispatchState: DispatchState = {
      status: "success",
      executionId: "execution-1",
    };
    const errorState: ApiKeyState = {
      status: "error",
      fieldErrors: { user_id: "Choose a user." },
    };

    expect(apiKeyState.key).toBe("secret-once");
    expect(deleteState.count).toBe(4);
    expect(dispatchState.executionId).toBe("execution-1");
    expect(errorState.fieldErrors?.user_id).toBe("Choose a user.");
  });
});
