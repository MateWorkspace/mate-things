import type { Metadata } from "next";

import PageHeader from "@/components/ui/page-header";
import { EmptyState } from "@/components/ui/states";
import { listInfraredDeviceTypes } from "@/lib/api/infrared";
import { requirePermission } from "@/lib/session";

import DeviceTypeCard from "./_components/DeviceTypeCard";
import DeviceTypeForm from "./_components/DeviceTypeForm";

export const metadata: Metadata = {
  title: "Infrared Settings — Mate Things",
  description: "Manage infrared device types and the states each one exposes.",
};

export default async function InfraredSettingsPage() {
  const { permissions } = await requirePermission("infrared_reference:get");
  const deviceTypes = await listInfraredDeviceTypes();
  const canAdd = permissions.has("infrared_reference:add");
  const canDelete = permissions.has("infrared_reference:delete");

  return (
    <main className="mx-auto w-full max-w-7xl space-y-6 px-4 py-6 sm:px-6 lg:px-8">
      <PageHeader
        title="Infrared Settings"
        description="Device types are the appliance categories."
        actions={canAdd ? <DeviceTypeForm /> : undefined}
      />
      <p className="text-muted-foreground text-xs font-semibold tracking-wider uppercase">
        {deviceTypes.length} device{" "}
        {deviceTypes.length === 1 ? "type" : "types"}
      </p>
      {deviceTypes.length ? (
        <section
          aria-label="Infrared device types"
          className="grid grid-cols-1 gap-5 sm:grid-cols-2 xl:grid-cols-3"
        >
          {deviceTypes.map((deviceType) => (
            <DeviceTypeCard
              key={deviceType.id}
              deviceType={deviceType}
              canDelete={canDelete}
            />
          ))}
        </section>
      ) : (
        <EmptyState
          title="No device types yet"
          description="Create a device type to start teaching the Infrared app new appliances."
        />
      )}
    </main>
  );
}
