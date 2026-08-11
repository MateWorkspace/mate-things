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
} from "@/lib/api/infrared";
import { getOptionalById } from "@/lib/api/optional";
import { requirePermission } from "@/lib/session";

import RecordCaseList from "./_components/RecordCaseList";
import RecordCoderPanel from "./_components/RecordCoderPanel";
import RecordSessionOverview from "./_components/RecordSessionOverview";
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

  const [cases, coder, testCases, device, deviceTypes] = await Promise.all([
    listInfraredRecordSessionCases(id),
    getOptionalById(() => getInfraredRecordSessionCoder(id)),
    // 404s with "infrared_state_coder not found" until a coder exists.
    getOptionalById(() => listInfraredRecordSessionTestCases(id)),
    getInfraredDevice(session.infrared_device_id),
    listInfraredDeviceTypes(),
  ]);

  const deviceTypeName =
    deviceTypes.find((dt) => dt.id === device.infrared_device_type_id)?.name ??
    "Unknown";
  const canSet = permissions.has("infrared_record_session:set");
  const canDelete = permissions.has("infrared_record_session:delete");

  return (
    <main className="mx-auto w-full max-w-5xl space-y-8 px-4 py-6 sm:px-6 lg:px-8">
      <PageHeader
        title="Record Session"
        description={`Session ${session.id}`}
      />
      <RecordSessionOverview
        session={session}
        brand={device.brand}
        model={device.model}
        deviceTypeName={deviceTypeName}
        canDelete={canDelete}
      />
      <section className="space-y-3">
        <h2 className="font-display text-primary text-xl tracking-wide">
          Cases
        </h2>
        <RecordCaseList cases={cases} canMutate={canSet} />
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
          <RecordTestCaseList testCases={testCases} canMutate={canSet} />
        </section>
      ) : null}
    </main>
  );
}
