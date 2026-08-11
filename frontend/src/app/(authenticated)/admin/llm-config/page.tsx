import type { Metadata } from "next";

import Card from "@/components/ui/card";
import LocalDateTime from "@/components/ui/local-date-time";
import PageHeader from "@/components/ui/page-header";
import StatusBadge from "@/components/ui/status-badge";
import { getLlmConfig } from "@/lib/api/llm-config";
import { requirePermission } from "@/lib/session";

import LlmConfigForm from "./_components/LlmConfigForm";
import TestConnectionButton from "./_components/TestConnectionButton";

export const metadata: Metadata = { title: "LLM Config — Mate Things" };

export default async function LlmConfigPage() {
  const { permissions } = await requirePermission("llm_config:get");
  const config = await getLlmConfig();
  const canSet = permissions.has("llm_config:set");

  return (
    <main className="mx-auto w-full max-w-3xl space-y-6 px-4 py-6 sm:px-6 lg:px-8">
      <PageHeader
        title="LLM Config"
        description="The active provider, model, and credential used for AI-generated content across the app (e.g. the Infrared app's device-teaching pipeline)."
      />

      <Card className="p-5">
        <h2 className="font-display text-primary text-xl tracking-wide">
          Current configuration
        </h2>
        <dl className="mt-4 grid grid-cols-1 gap-4 text-sm sm:grid-cols-2">
          <div>
            <dt className="text-muted-foreground">Provider</dt>
            <dd className="mt-1 font-medium">
              {config.provider === "CLAUDE" ? "Claude" : "OpenAI"}
            </dd>
          </div>
          <div>
            <dt className="text-muted-foreground">Model</dt>
            <dd className="mt-1 font-mono">{config.model || "—"}</dd>
          </div>
          <div>
            <dt className="text-muted-foreground">Base URL</dt>
            <dd className="mt-1 font-mono break-all">
              {config.base_url || "Provider default"}
            </dd>
          </div>
          <div>
            <dt className="text-muted-foreground">API key</dt>
            <dd className="mt-1">
              <StatusBadge variant={config.api_key_set ? "success" : "warning"}>
                {config.api_key_set ? "Set" : "Not set"}
              </StatusBadge>
            </dd>
          </div>
          {config.updated_at ? (
            <div>
              <dt className="text-muted-foreground">Last updated</dt>
              <dd className="mt-1 font-medium">
                <LocalDateTime value={config.updated_at} />
              </dd>
            </div>
          ) : null}
        </dl>
        {canSet && config.api_key_set ? (
          <div className="mt-5">
            <TestConnectionButton />
          </div>
        ) : null}
      </Card>

      {canSet ? <LlmConfigForm config={config} /> : null}
    </main>
  );
}
