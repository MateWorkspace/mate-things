"use server";

import { refresh } from "next/cache";

import { ApiError } from "@/lib/api/client";
import { updateProfile, updateProfilePassword } from "@/lib/api/profile";
import { requireSessionContext } from "@/lib/session";

export interface FormActionState {
  status: "idle" | "success" | "error";
  title?: string;
  message?: string;
  fieldErrors?: Record<string, string>;
}

function permissionDenied(): FormActionState {
  return {
    status: "error",
    title: "Permission denied",
    message: "You do not have permission to make this change.",
  };
}

function actionError(error: unknown): FormActionState {
  if (error instanceof ApiError) {
    const validationDetails =
      error.status === 400 ? error.details?.trim() : undefined;

    return {
      status: "error",
      title: error.title,
      message: validationDetails || error.message,
    };
  }

  return {
    status: "error",
    title: "Something went wrong",
    message: "Please try again.",
  };
}

export async function saveProfileAction(
  _previousState: FormActionState,
  formData: FormData,
): Promise<FormActionState> {
  const session = await requireSessionContext();
  if (!session.permissions.has("profile:set")) {
    return permissionDenied();
  }

  const name = String(formData.get("name") ?? "").trim();
  const username = String(formData.get("username") ?? "").trim();
  const bio = String(formData.get("bio") ?? "").trim();
  const fieldErrors: Record<string, string> = {};

  if (!name) {
    fieldErrors.name = "Enter your name.";
  }
  if (!username) {
    fieldErrors.username = "Enter your username.";
  }
  if (Object.keys(fieldErrors).length > 0) {
    return {
      status: "error",
      title: "Check your details",
      message: "Complete the required fields.",
      fieldErrors,
    };
  }

  try {
    await updateProfile({ name, username, bio });
    refresh();
  } catch (error) {
    return actionError(error);
  }

  return {
    status: "success",
    title: "Profile updated",
    message: "Your profile details have been saved.",
  };
}

export async function changePasswordAction(
  _previousState: FormActionState,
  formData: FormData,
): Promise<FormActionState> {
  const session = await requireSessionContext();
  if (!session.permissions.has("profile_security:set")) {
    return permissionDenied();
  }

  const currentPassword = String(formData.get("current_password") ?? "");
  const newPassword = String(formData.get("new_password") ?? "");
  const confirmPassword = String(formData.get("confirm_password") ?? "");

  if (newPassword !== confirmPassword) {
    return {
      status: "error",
      title: "Passwords do not match",
      message: "Confirm your new password and try again.",
      fieldErrors: {
        confirm_password: "The confirmation must match the new password.",
      },
    };
  }

  try {
    await updateProfilePassword({
      current_password: currentPassword,
      new_password: newPassword,
    });
  } catch (error) {
    return actionError(error);
  }

  return {
    status: "success",
    title: "Password changed",
    message: "Your password has been updated.",
  };
}
