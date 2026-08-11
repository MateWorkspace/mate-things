import type { Metadata } from "next";

import PageHeader from "@/components/ui/page-header";
import { EmptyState } from "@/components/ui/states";
import { requirePermission } from "@/lib/session";

export const metadata: Metadata = { title: "Infrared Command — Mate Things" };

export default async function InfraredCommandPage() {
  await requirePermission("infrared_record_session:get");

  return (
    <main className="mx-auto w-full max-w-7xl space-y-6 px-4 py-6 sm:px-6 lg:px-8">
      <PageHeader
        title="Command"
        description="Dispatch a taught infrared command to a device and confirm it worked."
      />
      <EmptyState
        title="Coming soon"
        description="Command dispatch isn't built yet. Set up device types and states in Infrared Settings first."
      />
    </main>
  );
}
