import type { Metadata } from "next";

import PageHeader from "@/components/ui/page-header";
import { listInfraredDeviceTypes } from "@/lib/api/infrared";
import { listAllNodes } from "@/lib/api/nodes";
import { requirePermission } from "@/lib/session";

import NewRecordSessionForm from "./_components/NewRecordSessionForm";

export const metadata: Metadata = { title: "Record New Device — Mate Things" };

export default async function NewRecordSessionPage() {
  const { permissions } = await requirePermission(
    "infrared_record_session:add",
  );

  const [nodes, deviceTypes] = await Promise.all([
    permissions.has("node:get") ? listAllNodes() : Promise.resolve([]),
    permissions.has("infrared_reference:get")
      ? listInfraredDeviceTypes()
      : Promise.resolve([]),
  ]);

  return (
    <main className="mx-auto w-full max-w-3xl space-y-8 px-4 py-6 sm:px-6 lg:px-8">
      <PageHeader
        title="Record New Device"
        description="Capture a remote's infrared signals to teach the Infrared app how to control this device."
      />
      <NewRecordSessionForm nodes={nodes} initialDeviceTypes={deviceTypes} />
    </main>
  );
}
