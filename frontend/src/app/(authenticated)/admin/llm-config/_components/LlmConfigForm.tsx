"use client";

import { useActionState, useId, useRef } from "react";

import ActionMessage from "@/components/forms/ActionMessage";
import FieldError from "@/components/forms/FieldError";
import Button from "@/components/ui/button";
import Input from "@/components/ui/input";
import Label from "@/components/ui/label";
import Select from "@/components/ui/select";
import { useFirstInvalidField } from "@/hooks/use-first-invalid-field";
import { useRefreshAfterAction } from "@/hooks/use-refresh-after-action";
import type { LlmConfigResponse } from "@/lib/api/llm-config";

import { updateLlmConfigAction } from "../_lib/actions";
import { EMPTY_LLM_CONFIG_STATE } from "../_lib/state";

export default function LlmConfigForm({
  config,
}: {
  config: LlmConfigResponse;
}) {
  const [state, action, pending] = useActionState(
    updateLlmConfigAction,
    EMPTY_LLM_CONFIG_STATE,
  );
  const fieldId = useId();
  const formRef = useRef<HTMLFormElement>(null);
  useFirstInvalidField(state, formRef);
  useRefreshAfterAction(state);

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
          defaultValue={config.provider}
        >
          <option value="CLAUDE">Claude</option>
          <option value="OPENAI">OpenAI</option>
        </Select>
        <FieldError>{state.fieldErrors?.provider}</FieldError>
      </div>

      <div>
        <Label htmlFor={`${fieldId}-model`}>Model</Label>
        <Input
          id={`${fieldId}-model`}
          name="model"
          defaultValue={config.model}
          placeholder="claude-sonnet-5"
          required
          aria-invalid={Boolean(state.fieldErrors?.model)}
          aria-describedby={
            state.fieldErrors?.model ? `${fieldId}-model-error` : undefined
          }
        />
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
        <Label htmlFor={`${fieldId}-api-key`}>API key</Label>
        <Input
          id={`${fieldId}-api-key`}
          name="api_key"
          type="password"
          autoComplete="off"
          placeholder={
            config.api_key_set ? "•••••••• (currently set)" : "Not set"
          }
          required
          aria-invalid={Boolean(state.fieldErrors?.api_key)}
          aria-describedby={`${fieldId}-api-key-hint`}
        />
        <p
          id={`${fieldId}-api-key-hint`}
          className="text-muted-foreground mt-1.5 text-xs"
        >
          Write-only - it&apos;s never shown again after saving. Every save
          requires it, even one that only changes the model or base URL.
        </p>
        <FieldError>{state.fieldErrors?.api_key}</FieldError>
      </div>

      <ActionMessage state={state} />

      <div className="flex justify-end">
        <Button type="submit" disabled={pending}>
          {pending ? "Saving…" : "Save configuration"}
        </Button>
      </div>
    </form>
  );
}
