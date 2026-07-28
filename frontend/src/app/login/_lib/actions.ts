"use server";

import { redirect } from "next/navigation";

import { login } from "@/lib/api/auth";
import { ApiError } from "@/lib/api/client";

export interface LoginFormState {
  title?: string;
  message?: string;
  username?: string;
}

export async function loginAction(
  _prevState: LoginFormState,
  formData: FormData,
): Promise<LoginFormState> {
  const username = String(formData.get("username") ?? "").trim();
  const password = String(formData.get("password") ?? "");

  if (!username || !password) {
    return {
      title: "Invalid Format",
      message: "Enter your username and password.",
      username,
    };
  }

  try {
    await login({ username, password });
  } catch (err) {
    if (err instanceof ApiError) {
      return { title: err.title, message: err.message, username };
    }
    return {
      title: "Something Went Wrong",
      message: "Please try again.",
      username,
    };
  }

  redirect("/dashboard");
}
