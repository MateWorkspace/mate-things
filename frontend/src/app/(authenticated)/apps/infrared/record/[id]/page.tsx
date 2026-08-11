import type { Metadata } from "next";
import { notFound } from "next/navigation";

import PageHeader from "@/components/ui/page-header";
import {
  getInfraredDevice,
  getInfraredRecordSession,
  getInfraredRecordSessionCoder,
  listInfraredDeviceTypes,
  listInfraredRecordSessionCases,
  listInfraredRecordSessionTestCases,
  listInfraredStates,
} from "@/lib/api/infrared";
import { getOptionalById } from "@/lib/api/optional";
import { requirePermission } from "@/lib/session";

import RecordCaseList from "./_components/RecordCaseList";
import RecordCoderPanel from "./_components/RecordCoderPanel";
import RecordSessionOverview from "./_components/RecordSessionOverview";
import RecordSessionWizard from "./_components/RecordSessionWizard";
import RecordTestCaseList from "./_components/RecordTestCaseList";

export const metadata: Metadata = { title: "Record Session — Mate Things" };

export default async function RecordSessionDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  const { permissions } = await requirePermission(
    "infrared_record_session:get",
  );

  const session = await getOptionalById(() => getInfraredRecordSession(id));
  if (!session) {
    notFound();
  }

  const canReadReferences = permissions.has("infrared_reference:get");
  const [cases, coder, testCases, device] = await Promise.all([
    listInfraredRecordSessionCases(id),
    getOptionalById(() => getInfraredRecordSessionCoder(id)),
    // 404s with "infrared_state_coder not found" until a coder exists.
    getOptionalById(() => listInfraredRecordSessionTestCases(id)),
    canReadReferences
      ? getInfraredDevice(session.infrared_device_id)
      : Promise.resolve(null),
  ]);
  // Case state names come from the device type's state definitions, not
  // the record-session API - only fetchable once the device (and so its
  // device type id) is known.
  const stateDefinitions = device
    ? await listInfraredStates(device.infrared_device_type_id)
    : [];

  // Terminal sessions (COMPLETED/FAILED) must not expose mutation controls.
  const canMutate =
    permissions.has("infrared_record_session:set") && !session.is_completed;
  const canDelete = permissions.has("infrared_record_session:delete");

  if (!session.is_completed) {
    return (
      <main className="mx-auto w-full max-w-5xl space-y-8 px-4 py-6 sm:px-6 lg:px-8">
        <PageHeader
          title="Record Session"
          description={`Session ${session.id}`}
        />
        <RecordSessionWizard
          session={session}
          cases={cases}
          testCases={testCases ?? []}
          canMutate={canMutate}
          stateDefinitions={stateDefinitions}
        />
      </main>
    );
  }

  const deviceTypes = canReadReferences ? await listInfraredDeviceTypes() : [];

  const deviceTypeName = device
    ? (deviceTypes.find((dt) => dt.id === device.infrared_device_type_id)
        ?.name ?? "Unknown")
    : null;

  return (
    <main className="mx-auto w-full max-w-5xl space-y-8 px-4 py-6 sm:px-6 lg:px-8">
      <PageHeader
        title="Record Session"
        description={`Session ${session.id}`}
      />
      <RecordSessionOverview
        session={session}
        brand={device?.brand ?? null}
        model={device?.model ?? null}
        deviceTypeName={deviceTypeName}
        canDelete={canDelete}
      />
      <section className="space-y-3">
        <h2 className="font-display text-primary text-xl tracking-wide">
          Cases
        </h2>
        <RecordCaseList
          cases={cases}
          canMutate={canMutate}
          stateDefinitions={stateDefinitions}
        />
      </section>
      {coder ? (
        <section className="space-y-3">
          <h2 className="font-display text-primary text-xl tracking-wide">
            Coder
          </h2>
          <RecordCoderPanel coder={coder} canDelete={canDelete} />
        </section>
      ) : null}
      {testCases?.length ? (
        <section className="space-y-3">
          <h2 className="font-display text-primary text-xl tracking-wide">
            Test cases
          </h2>
          <RecordTestCaseList
            testCases={testCases}
            canMutate={canMutate}
            stateDefinitions={stateDefinitions}
          />
        </section>
      ) : null}
    </main>
  );
}
