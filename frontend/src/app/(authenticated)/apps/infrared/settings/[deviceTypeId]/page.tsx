import type { Metadata } from "next";
import { notFound } from "next/navigation";

import PageHeader from "@/components/ui/page-header";
import { EmptyState } from "@/components/ui/states";
import {
  listInfraredDeviceTypes,
  listInfraredStates,
} from "@/lib/api/infrared";
import { requirePermission } from "@/lib/session";

import DeleteStateDialog from "./_components/DeleteStateDialog";
import StateForm from "./_components/StateForm";

interface DeviceTypeDetailPageProps {
  params: Promise<{ deviceTypeId: string }>;
}

export async function generateMetadata({
  params,
}: DeviceTypeDetailPageProps): Promise<Metadata> {
  const { deviceTypeId } = await params;
  const deviceTypes = await listInfraredDeviceTypes();
  const deviceType = deviceTypes.find((item) => item.id === deviceTypeId);

  return {
    title: deviceType
      ? `${deviceType.name} — Mate Things`
      : "Device Type — Mate Things",
  };
}

export default async function DeviceTypeDetailPage({
  params,
}: DeviceTypeDetailPageProps) {
  const [{ permissions }, { deviceTypeId }] = await Promise.all([
    requirePermission("infrared_reference:get"),
    params,
  ]);
  const [deviceTypes, states] = await Promise.all([
    listInfraredDeviceTypes(),
    listInfraredStates(deviceTypeId),
  ]);
  const deviceType = deviceTypes.find((item) => item.id === deviceTypeId);

  if (!deviceType) {
    notFound();
  }

  const canAdd = permissions.has("infrared_reference:add");
  const canDelete = permissions.has("infrared_reference:delete");

  return (
    <main className="mx-auto w-full max-w-4xl space-y-6 px-4 py-6 sm:px-6 lg:px-8">
      <PageHeader
        title={deviceType.name}
        description="States are the controllable dimensions devices of this type expose, e.g. POWER, MODE, TEMPERATURE, etc. Each device is taught concrete option/range values per state when it's recorded."
        actions={
          canAdd ? <StateForm deviceTypeId={deviceType.id} /> : undefined
        }
      />
      {states.length ? (
        <ul className="space-y-3">
          {states.map((state) => (
            <li
              key={state.id}
              className="border-border bg-surface flex flex-wrap items-center justify-between gap-3 rounded-2xl border p-4"
            >
              <div className="flex items-center gap-3">
                <span className="font-semibold">{state.name}</span>
                <span className="bg-muted text-muted-foreground rounded-full px-2.5 py-1 text-xs font-semibold tracking-wide">
                  {state.type}
                </span>
              </div>
              {canDelete ? <DeleteStateDialog state={state} /> : null}
            </li>
          ))}
        </ul>
      ) : (
        <EmptyState
          title="No states yet"
          description="Add a state to define what this device type can be taught to control."
        />
      )}
    </main>
  );
}
