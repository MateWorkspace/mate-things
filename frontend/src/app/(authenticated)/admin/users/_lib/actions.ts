"use server";

import { revalidatePath } from "next/cache";
import { redirect } from "next/navigation";

import { ApiError } from "@/lib/api/client";
import {
  createUser,
  deleteUser,
  updateUser,
  updateUserPassword,
} from "@/lib/api/users";
import { requireSessionContext } from "@/lib/session";

import type { UserActionState } from "./state";

function denied(): UserActionState {
  return {
    status: "error",
    title: "Permission denied",
    message: "You do not have permission to make this change.",
  };
}

function failure(error: unknown): UserActionState {
  return {
    status: "error",
    title: error instanceof ApiError ? error.title : "Something went wrong",
    message:
      error instanceof ApiError
        ? error.message
        : "Please try again.",
  };
}

function required(values: Record<string, string>): Record<string, string> {
  return Object.fromEntries(
    Object.entries(values)
      .filter(([, value]) => !value)
      .map(([name]) => [name, "This field is required."]),
  );
}

export async function createUserAction(
  _previous: UserActionState,
  formData: FormData,
): Promise<UserActionState> {
  const session = await requireSessionContext();
  if (!session.permissions.has("user:add")) return denied();
  const roleId = String(formData.get("role_id") ?? "").trim();
  const name = String(formData.get("name") ?? "").trim();
  const bio = String(formData.get("bio") ?? "").trim();
  const username = String(formData.get("username") ?? "").trim();
  const password = String(formData.get("password") ?? "");
  const fieldErrors = required({ role_id: roleId, name, username, password });
  if (Object.keys(fieldErrors).length) {
    return {
      status: "error",
      title: "Check the user",
      message: "Complete the required fields.",
      fieldErrors,
    };
  }
  try {
    await createUser({ role_id: roleId, name, bio, username, password });
    revalidatePath("/admin/users");
    return {
      status: "success",
      title: "User created",
      message: `${name} can now sign in.`,
    };
  } catch (error) {
    return failure(error);
  }
}

export async function updateUserAction(
  _previous: UserActionState,
  formData: FormData,
): Promise<UserActionState> {
  const session = await requireSessionContext();
  if (!session.permissions.has("user:set")) return denied();
  const userId = String(formData.get("user_id") ?? "").trim();
  const roleId = String(formData.get("role_id") ?? "").trim();
  const name = String(formData.get("name") ?? "").trim();
  const bio = String(formData.get("bio") ?? "").trim();
  const username = String(formData.get("username") ?? "").trim();
  const fieldErrors = required({
    user_id: userId,
    role_id: roleId,
    name,
    username,
  });
  if (Object.keys(fieldErrors).length) {
    return {
      status: "error",
      title: "Check the user",
      message: "Complete the required fields.",
      fieldErrors,
    };
  }
  try {
    await updateUser(userId, { role_id: roleId, name, bio, username });
    revalidatePath(`/admin/users/${userId}`);
    revalidatePath("/admin/users");
    return {
      status: "success",
      title: "User updated",
      message: "Account details and role were saved.",
    };
  } catch (error) {
    return failure(error);
  }
}

export async function resetUserPasswordAction(
  _previous: UserActionState,
  formData: FormData,
): Promise<UserActionState> {
  const session = await requireSessionContext();
  if (!session.permissions.has("user_password:set")) return denied();
  const userId = String(formData.get("user_id") ?? "").trim();
  const password = String(formData.get("password") ?? "");
  const confirmation = String(formData.get("password_confirmation") ?? "");
  const fieldErrors = required({ user_id: userId, password });
  if (password !== confirmation)
    fieldErrors.password_confirmation = "Passwords do not match.";
  if (Object.keys(fieldErrors).length) {
    return {
      status: "error",
      title: "Check the password",
      message: "Correct the highlighted fields.",
      fieldErrors,
    };
  }
  try {
    await updateUserPassword(userId, password);
    return {
      status: "success",
      title: "Password reset",
      message: "The new password takes effect immediately.",
    };
  } catch (error) {
    return failure(error);
  }
}

export async function deleteUserAction(
  _previous: UserActionState,
  formData: FormData,
): Promise<UserActionState> {
  const session = await requireSessionContext();
  if (!session.permissions.has("user:remove")) return denied();
  const userId = String(formData.get("user_id") ?? "").trim();
  const username = String(formData.get("username") ?? "").trim();
  const confirmation = String(formData.get("confirmation") ?? "");
  if (!userId || !username || confirmation !== username) {
    return {
      status: "error",
      title: "Username does not match",
      message: "Enter the exact target username before deleting this account.",
      fieldErrors: {
        confirmation: "The confirmation must exactly match the username.",
      },
    };
  }
  try {
    await deleteUser(userId);
  } catch (error) {
    return failure(error);
  }
  if (session.user.id === userId) redirect("/auth/invalid-session");
  redirect("/admin/users");
}
