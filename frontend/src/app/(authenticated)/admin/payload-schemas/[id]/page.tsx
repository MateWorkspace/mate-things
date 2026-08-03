import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";

import PreferencesDialog from "@/components/preferences/PreferencesDialog";
import Card from "@/components/ui/card";
import LocalDateTime from "@/components/ui/local-date-time";
import PageHeader from "@/components/ui/page-header";
import StatusBadge from "@/components/ui/status-badge";
import { ApiError } from "@/lib/api/client";
import { getPayloadSchemaById } from "@/lib/api/payload-schemas";
import { requirePermission } from "@/lib/session";

import PayloadSchemaForm from "../_components/PayloadSchemaForm";

export const metadata: Metadata = {
  title: "Payload Schema details — Mate Things",
};

const REFERENCE_TIME = Date.now();

export default async function PayloadSchemaDetails({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const [{ id }, session] = await Promise.all([
    params,
    requirePermission("payload_schema:get"),
  ]);
  let schema;
  try {
    schema = await getPayloadSchemaById(id);
  } catch (error) {
    if (error instanceof ApiError && error.status === 404) notFound();
    throw error;
  }
  const now = REFERENCE_TIME;
  const state =
    new Date(schema.valid_from).valueOf() > now
      ? { label: "Future", variant: "warning" as const }
      : schema.valid_to && new Date(schema.valid_to).valueOf() <= now
        ? { label: "Expired", variant: "neutral" as const }
        : { label: "Current", variant: "success" as const };
  return (
    <main className="mx-auto w-full max-w-7xl space-y-6 px-4 py-6 sm:px-6 lg:px-8">
      <PageHeader
        title={`${schema.name} v${schema.version}`}
        description="Version definition, validity, preferences, and related actions."
        actions={
          <div className="flex flex-wrap gap-2">
            {session.permissions.has("payload_schema:set") ||
            session.permissions.has("payload_schema:remove") ? (
              <PayloadSchemaForm
                schema={schema}
                canEdit={session.permissions.has("payload_schema:set")}
                canDelete={session.permissions.has("payload_schema:remove")}
              />
            ) : null}
            <PreferencesDialog
              resource="payload_schema"
              id={schema.id}
              preferences={schema.preferences}
              permissions={[...session.permissions]}
            />
          </div>
        }
      />
      <section className="grid gap-5 lg:grid-cols-3">
        <Card>
          <div className="flex justify-between">
            <h2 className="font-display text-primary text-xl">Validity</h2>
            <StatusBadge variant={state.variant}>{state.label}</StatusBadge>
          </div>
          <dl className="mt-4 space-y-3 text-sm">
            <div>
              <dt className="text-muted-foreground">From</dt>
              <dd>
                <LocalDateTime value={schema.valid_from} />
              </dd>
            </div>
            <div>
              <dt className="text-muted-foreground">To</dt>
              <dd>
                {schema.valid_to ? (
                  <LocalDateTime value={schema.valid_to} />
                ) : (
                  "Open-ended"
                )}
              </dd>
            </div>
          </dl>
        </Card>
        <Card>
          <h2 className="font-display text-primary text-xl">Audit</h2>
          <dl className="mt-4 space-y-3 text-sm">
            <div>
              <dt className="text-muted-foreground">Created</dt>
              <dd>
                <LocalDateTime value={schema.created_at} />
              </dd>
            </div>
            {schema.updated_at ? (
              <div>
                <dt className="text-muted-foreground">Updated</dt>
                <dd>
                  <LocalDateTime value={schema.updated_at} />
                </dd>
              </div>
            ) : null}
          </dl>
        </Card>
        <Card>
          <h2 className="font-display text-primary text-xl">References</h2>
          <p className="text-muted-foreground mt-3 text-sm">
            Find Actions that validate against this exact schema version.
          </p>
          <Link
            href={`/actions?payload_schema_id=${schema.id}`}
            className="text-primary mt-4 inline-block font-semibold underline"
          >
            View filtered actions
          </Link>
        </Card>
      </section>
      <Card>
        <h2 className="font-display text-primary text-xl">Definition</h2>
        <pre className="border-border bg-muted mt-4 overflow-auto rounded-xl border p-4 font-mono text-sm leading-6">
          {JSON.stringify(schema.definition, null, 2)}
        </pre>
      </Card>
      <Card>
        <h2 className="font-display text-primary text-xl">Preferences</h2>
        <pre className="bg-muted mt-4 overflow-auto rounded-xl p-4 text-sm">
          {JSON.stringify(schema.preferences, null, 2)}
        </pre>
      </Card>
    </main>
  );
}
