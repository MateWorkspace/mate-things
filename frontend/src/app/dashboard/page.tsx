import type { Metadata } from "next";
import Image from "next/image";

import hat from "@/assets/hat.svg";
import { requireSession } from "@/lib/session";

import LogoutButton from "./_components/LogoutButton";

export const metadata: Metadata = {
  title: "Dashboard — Mate Things",
};

export default async function DashboardPage() {
  const session = await requireSession();

  return (
    <div className="flex min-h-screen flex-col bg-background">
      <header className="flex items-center justify-between gap-4 border-b border-ink/10 bg-surface px-6 py-4 sm:px-10">
        <div className="flex items-center gap-3">
          <Image src={hat} alt="" className="h-8 w-8 sm:h-9 sm:w-9" />
          <div>
            <p className="font-display text-lg tracking-wide text-primary sm:text-xl">
              Dashboard
            </p>
            <p className="text-xs text-primary/70 sm:text-sm">
              Signed in as {session.name}
            </p>
          </div>
        </div>
        <LogoutButton />
      </header>

      <main className="flex flex-1 items-center justify-center px-6 py-16 text-center">
        <div className="max-w-sm">
          <p className="font-display text-2xl text-primary sm:text-3xl">
            Nothing here yet
          </p>
          <p className="mt-3 text-sm text-foreground/70">
            Your fleet of devices will show up here once the dashboard is
            built out.
          </p>
        </div>
      </main>
    </div>
  );
}
