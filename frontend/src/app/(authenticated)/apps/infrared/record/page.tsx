import type { Metadata } from "next";

import PageHeader from "@/components/ui/page-header";
import { EmptyState } from "@/components/ui/states";
import { requirePermission } from "@/lib/session";

export const metadata: Metadata = { title: "Infrared Record — Mate Things" };

export default async function InfraredRecordPage() {
  await requirePermission("infrared_record_session:get");

  return (
    <main className="mx-auto w-full max-w-7xl space-y-6 px-4 py-6 sm:px-6 lg:px-8">
      <PageHeader
        title="Record"
        description="Teach a node to control a physical device by capturing its remote's infrared signals."
      />
      <EmptyState
        title="Coming soon"
        description="Recording sessions aren't built yet. Set up device types and states in Infrared Settings first."
      />
    </main>
  );
}
