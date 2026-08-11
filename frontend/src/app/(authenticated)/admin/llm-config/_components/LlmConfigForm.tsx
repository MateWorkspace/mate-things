"use client";

import { useActionState, useId, useRef, useState } from "react";

import ActionMessage from "@/components/forms/ActionMessage";
import FieldError from "@/components/forms/FieldError";
import Button from "@/components/ui/button";
import Input from "@/components/ui/input";
import Label from "@/components/ui/label";
import Select from "@/components/ui/select";
import StatusBadge from "@/components/ui/status-badge";
import { useFirstInvalidField } from "@/hooks/use-first-invalid-field";
import { useRefreshAfterAction } from "@/hooks/use-refresh-after-action";
import type { LlmConfigResponse, LlmProvider } from "@/lib/api/llm-config";

import {
  testLlmConfigFormAction,
  updateLlmConfigAction,
} from "../_lib/actions";
import {
  EMPTY_LLM_CONFIG_STATE,
  EMPTY_TEST_CONNECTION_STATE,
} from "../_lib/state";

const MODEL_OPTIONS: Record<LlmProvider, readonly string[]> = {
  CLAUDE: [
    "claude-opus-5",
    "claude-sonnet-5",
    "claude-fable-5",
    "claude-haiku-4-5-20251001",
  ],
  OPENAI: ["gpt-5", "gpt-5-mini", "gpt-5-nano", "gpt-4.1", "gpt-4o"],
  GEMINI: [
    "gemini-2.5-pro",
    "gemini-2.5-flash",
    "gemini-2.5-flash-lite",
    "gemini-2.0-flash",
  ],
};

function normalizeProvider(provider: LlmProvider): LlmProvider {
  return provider === "OPENAI" || provider === "GEMINI" ? provider : "CLAUDE";
}

function modelOptionsFor(provider: LlmProvider, currentModel: string) {
  const options = MODEL_OPTIONS[normalizeProvider(provider)];
  return currentModel && !options.includes(currentModel)
    ? [currentModel, ...options]
    : options;
}

export default function LlmConfigForm({
  config,
}: {
  config: LlmConfigResponse;
}) {
  const [state, action, pending] = useActionState(
    updateLlmConfigAction,
    EMPTY_LLM_CONFIG_STATE,
  );
  const [testState, testAction, testPending] = useActionState(
    testLlmConfigFormAction,
    EMPTY_TEST_CONNECTION_STATE,
  );
  const fieldId = useId();
  const formRef = useRef<HTMLFormElement>(null);
  useFirstInvalidField(state, formRef);
  useRefreshAfterAction(state);

  const configProvider = normalizeProvider(config.provider);
  const [provider, setProvider] = useState<LlmProvider>(configProvider);
  const modelOptions = modelOptionsFor(
    provider,
    provider === configProvider ? config.model : "",
  );

  return (
    <form
      ref={formRef}
      action={action}
      onReset={(event) => event.preventDefault()}
      className="border-border bg-muted space-y-4 rounded-2xl border p-4"
    >
      <div>
        <Label htmlFor={`${fieldId}-provider`}>Provider</Label>
        <Select
          id={`${fieldId}-provider`}
          name="provider"
          value={provider}
          onChange={(event) => setProvider(event.target.value as LlmProvider)}
        >
          <option value="CLAUDE">Claude</option>
          <option value="OPENAI">OpenAI</option>
          <option value="GEMINI">Gemini</option>
        </Select>
        <FieldError>{state.fieldErrors?.provider}</FieldError>
      </div>

      <div>
        <Label htmlFor={`${fieldId}-model`}>Model</Label>
        <Select
          id={`${fieldId}-model`}
          name="model"
          key={provider}
          defaultValue={
            provider === configProvider ? config.model : modelOptions[0]
          }
          aria-invalid={Boolean(state.fieldErrors?.model)}
          aria-describedby={
            state.fieldErrors?.model ? `${fieldId}-model-error` : undefined
          }
        >
          {modelOptions.map((model) => (
            <option key={model} value={model}>
              {model}
            </option>
          ))}
        </Select>
        <FieldError id={`${fieldId}-model-error`}>
          {state.fieldErrors?.model}
        </FieldError>
      </div>

      <div>
        <Label htmlFor={`${fieldId}-base-url`}>Base URL (optional)</Label>
        <Input
          id={`${fieldId}-base-url`}
          name="base_url"
          defaultValue={config.base_url}
          placeholder="Leave blank for the provider's default endpoint"
        />
      </div>

      <div>
        <input
          type="hidden"
          name="api_key_was_set"
          value={String(config.api_key_set)}
        />
        <Label htmlFor={`${fieldId}-api-key`}>API key</Label>
        <Input
          id={`${fieldId}-api-key`}
          name="api_key"
          type="password"
          autoComplete="off"
          placeholder={
            config.api_key_set ? "•••••••• (currently set)" : "Not set"
          }
          aria-invalid={Boolean(state.fieldErrors?.api_key)}
          aria-describedby={`${fieldId}-api-key-hint`}
        />
        <p
          id={`${fieldId}-api-key-hint`}
          className="text-muted-foreground mt-1.5 text-xs"
        >
          {config.api_key_set
            ? "Write-only - it's never shown again after saving. Leave blank to keep the current key."
            : "Required the first time you configure this provider."}
        </p>
        <FieldError>{state.fieldErrors?.api_key}</FieldError>
      </div>

      <ActionMessage state={state} />

      <div className="flex flex-wrap items-center justify-end gap-3">
        {testState.status === "result" ? (
          <StatusBadge
            variant={
              testState.connectionStatus === "CONNECTED"
                ? "success"
                : "critical"
            }
          >
            {testState.connectionStatus === "CONNECTED"
              ? "Connected"
              : "Disconnected"}
          </StatusBadge>
        ) : testState.status === "error" ? (
          <span className="text-critical text-sm">{testState.message}</span>
        ) : null}
        <Button
          type="submit"
          formAction={testAction}
          variant="secondary"
          disabled={pending || testPending}
        >
          {testPending ? "Testing…" : "Test connection"}
        </Button>
        <Button type="submit" disabled={pending || testPending}>
          {pending ? "Saving…" : "Save configuration"}
        </Button>
      </div>
    </form>
  );
}
