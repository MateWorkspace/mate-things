import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Dashboard — Mate Things",
};

export default function DashboardPage() {
  return (
    <main className="flex min-h-[calc(100vh-4rem)] items-center justify-center px-6 py-16 text-center">
      <div className="max-w-sm">
        <p className="font-display text-primary text-2xl sm:text-3xl">
          Nothing here yet
        </p>
        <p className="text-foreground/70 mt-3 text-sm">
          Your fleet of devices will show up here once the dashboard is built
          out.
        </p>
      </div>
    </main>
  );
}
