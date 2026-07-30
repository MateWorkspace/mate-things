"use server";

import { cookies } from "next/headers";
import { redirect } from "next/navigation";

import { logout } from "@/lib/api/auth";
import { requireSession } from "@/lib/session";

const SIDEBAR_COLLAPSED_COOKIE = "mate_sidebar_collapsed";

export async function logoutAction(): Promise<void> {
  await logout();
  redirect("/login?signedOut=1");
}

export async function setSidebarCollapsedAction(
  collapsed: boolean,
): Promise<void> {
  await requireSession();
  const cookieStore = await cookies();

  cookieStore.set(SIDEBAR_COLLAPSED_COOKIE, collapsed ? "1" : "0", {
    httpOnly: true,
    maxAge: 60 * 60 * 24 * 365,
    path: "/",
    sameSite: "lax",
    secure: process.env.NODE_ENV === "production",
  });
}
