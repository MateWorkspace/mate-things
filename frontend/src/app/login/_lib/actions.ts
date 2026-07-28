"use server";

import { redirect } from "next/navigation";

import { login } from "@/lib/api/auth";
import { ApiError } from "@/lib/api/client";

export interface LoginFormState {
  error?: string;
  username?: string;
}

export async function loginAction(
  _prevState: LoginFormState,
  formData: FormData,
): Promise<LoginFormState> {
  const username = String(formData.get("username") ?? "").trim();
  const password = String(formData.get("password") ?? "");

  if (!username || !password) {
    return { error: "Enter your username and password.", username };
  }

  try {
    await login({ username, password });
  } catch (err) {
    if (err instanceof ApiError) {
      return {
        error:
          err.status === 401
            ? "Incorrect username or password."
            : err.message,
        username,
      };
    }
    return { error: "Something went wrong. Try again.", username };
  }

  redirect("/dashboard");
}
