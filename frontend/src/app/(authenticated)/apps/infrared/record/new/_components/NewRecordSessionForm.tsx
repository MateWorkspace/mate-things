"use client";

import { useActionState, useEffect, useId, useState } from "react";

import ActionMessage from "@/components/forms/ActionMessage";
import Button from "@/components/ui/button";
import Input from "@/components/ui/input";
import Label from "@/components/ui/label";
import Select from "@/components/ui/select";
import type {
  InfraredDeviceTypeResponse,
  InfraredStateResponse,
  InfraredStateType,
  StartRecordSessionDefinition,
} from "@/lib/api/infrared";
import type { NodeResponse } from "@/lib/api/nodes";

import {
  listStatesForDeviceTypeAction,
  startRecordSessionAction,
} from "../_lib/actions";
import { EMPTY_NEW_RECORD_SESSION_STATE } from "../_lib/state";
import StringListEditor from "../../_components/StringListEditor";
import WizardStepper from "../../_components/WizardStepper";
import { RECORD_WIZARD_STEPS } from "../../_lib/wizard-steps";
import DeviceTypeCreateDialog from "./DeviceTypeCreateDialog";
import StateCreateDialog from "./StateCreateDialog";

interface StateDraft {
  stateId: string;
  name: string;
  type: InfraredStateType;
  options: string[];
  minimum: string;
  maximum: string;
  step: string;
}

function draftFromState(state: InfraredStateResponse): StateDraft {
  return {
    stateId: state.id,
    name: state.name,
    type: state.type,
    options: state.type === "ENUM" ? [""] : [],
    minimum: "",
    maximum: "",
    step: "",
  };
}

export default function NewRecordSessionForm({
  nodes,
  initialDeviceTypes,
}: {
  nodes: readonly NodeResponse[];
  initialDeviceTypes: readonly InfraredDeviceTypeResponse[];
}) {
  const [wizardStep, setWizardStep] = useState<1 | 2>(1);
  const [deviceTypes, setDeviceTypes] = useState(initialDeviceTypes);
  const [nodeId, setNodeId] = useState("");
  const [deviceTypeId, setDeviceTypeId] = useState("");
  const [brand, setBrand] = useState("");
  const [model, setModel] = useState("");
  const [drafts, setDrafts] = useState<StateDraft[]>([]);
  const [statesLoading, setStatesLoading] = useState(false);
  const [stepError, setStepError] = useState<string | null>(null);
  const fieldId = useId();

  const [state, formAction, pending] = useActionState(
    startRecordSessionAction,
    EMPTY_NEW_RECORD_SESSION_STATE,
  );

  useEffect(() => {
    if (!deviceTypeId) {
      return;
    }
    let cancelled = false;
    setStatesLoading(true);
    listStatesForDeviceTypeAction(deviceTypeId)
      .then((states) => {
        if (!cancelled) {
          setDrafts(states.map(draftFromState));
        }
      })
      .catch(() => {
        if (!cancelled) {
          setStepError("Could not load this device type's states. Try again.");
        }
      })
      .finally(() => {
        if (!cancelled) {
          setStatesLoading(false);
        }
      });
    return () => {
      cancelled = true;
    };
  }, [deviceTypeId]);

  function goToStep2() {
    if (!nodeId || !deviceTypeId || !brand.trim() || !model.trim()) {
      setStepError("Fill in every field before continuing.");
      return;
    }
    setStepError(null);
    setWizardStep(2);
  }

  function buildDefinitions(): StartRecordSessionDefinition[] {
    return drafts.map((draft) =>
      draft.type === "ENUM"
        ? {
            infrared_state_id: draft.stateId,
            options: draft.options.map((o) => o.trim()).filter(Boolean),
          }
        : {
            infrared_state_id: draft.stateId,
            minimum: Number(draft.minimum),
            maximum: Number(draft.maximum),
            step: Number(draft.step),
          },
    );
  }

  function handleSubmit(formData: FormData) {
    if (drafts.length === 0) {
      setStepError("Add at least one state value before starting.");
      return;
    }
    const invalidEnum = drafts.some(
      (d) =>
        d.type === "ENUM" && d.options.filter((o) => o.trim()).length === 0,
    );
    const invalidRange = drafts.some(
      (d) =>
        d.type === "RANGE" &&
        (d.minimum === "" || d.maximum === "" || d.step === ""),
    );
    if (invalidEnum || invalidRange) {
      setStepError("Complete every state's values before starting.");
      return;
    }
    setStepError(null);
    formData.set(
      "payload",
      JSON.stringify({
        node_id: nodeId,
        infrared_device_type_id: deviceTypeId,
        brand: brand.trim(),
        model: model.trim(),
        definitions: buildDefinitions(),
      }),
    );
    formAction(formData);
  }

  return (
    <div className="space-y-6">
      <WizardStepper
        steps={RECORD_WIZARD_STEPS}
        currentIndex={wizardStep - 1}
      />
      {wizardStep === 1 ? (
        <div className="space-y-4">
          <div>
            <Label htmlFor={`${fieldId}-node`}>Node</Label>
            <Select
              id={`${fieldId}-node`}
              value={nodeId}
              onChange={(event) => setNodeId(event.target.value)}
            >
              <option value="">Select a node</option>
              {nodes.map((node) => (
                <option key={node.id} value={node.id}>
                  {node.name || node.device_id}
                </option>
              ))}
            </Select>
          </div>
          <div>
            <Label htmlFor={`${fieldId}-device-type`}>Device type</Label>
            <div className="flex items-center gap-2">
              <Select
                id={`${fieldId}-device-type`}
                value={deviceTypeId}
                onChange={(event) => setDeviceTypeId(event.target.value)}
              >
                <option value="">Select a device type</option>
                {deviceTypes.map((dt) => (
                  <option key={dt.id} value={dt.id}>
                    {dt.name}
                  </option>
                ))}
              </Select>
              <DeviceTypeCreateDialog
                onCreated={(item) => {
                  setDeviceTypes((prev) => [
                    ...prev,
                    {
                      id: item.id,
                      name: item.name,
                      created_at: new Date().toISOString(),
                    },
                  ]);
                  setDeviceTypeId(item.id);
                }}
              />
            </div>
          </div>
          <div>
            <Label htmlFor={`${fieldId}-brand`}>Brand</Label>
            <Input
              id={`${fieldId}-brand`}
              value={brand}
              onChange={(event) => setBrand(event.target.value)}
            />
          </div>
          <div>
            <Label htmlFor={`${fieldId}-model`}>Model</Label>
            <Input
              id={`${fieldId}-model`}
              value={model}
              onChange={(event) => setModel(event.target.value)}
            />
          </div>
          {stepError ? (
            <p className="text-critical text-sm">{stepError}</p>
          ) : null}
          <div className="flex justify-end">
            <Button type="button" onClick={goToStep2}>
              Continue
            </Button>
          </div>
        </div>
      ) : (
        <div className="space-y-4">
          <StateCreateDialog
            deviceTypeId={deviceTypeId}
            onCreated={(item, type) => {
              setDrafts((prev) => [
                ...prev,
                {
                  stateId: item.id,
                  name: item.name,
                  type,
                  options: type === "ENUM" ? [""] : [],
                  minimum: "",
                  maximum: "",
                  step: "",
                },
              ]);
            }}
          />
          <form action={handleSubmit} className="space-y-4">
            {statesLoading ? (
              <p className="text-foreground/70 text-sm">Loading states…</p>
            ) : (
              drafts.map((draft, index) => (
                <div
                  key={draft.stateId}
                  className="border-border rounded-2xl border p-4"
                >
                  <p className="font-semibold">{draft.name}</p>
                  {draft.type === "ENUM" ? (
                    <div className="mt-2">
                      <StringListEditor
                        values={draft.options}
                        placeholder="Option value"
                        onChange={(options) => {
                          const next = [...drafts];
                          next[index] = { ...draft, options };
                          setDrafts(next);
                        }}
                      />
                    </div>
                  ) : (
                    <div className="mt-2 grid grid-cols-3 gap-2">
                      <Input
                        type="number"
                        placeholder="Minimum"
                        value={draft.minimum}
                        onChange={(event) => {
                          const next = [...drafts];
                          next[index] = {
                            ...draft,
                            minimum: event.target.value,
                          };
                          setDrafts(next);
                        }}
                      />
                      <Input
                        type="number"
                        placeholder="Maximum"
                        value={draft.maximum}
                        onChange={(event) => {
                          const next = [...drafts];
                          next[index] = {
                            ...draft,
                            maximum: event.target.value,
                          };
                          setDrafts(next);
                        }}
                      />
                      <Input
                        type="number"
                        placeholder="Step"
                        value={draft.step}
                        onChange={(event) => {
                          const next = [...drafts];
                          next[index] = { ...draft, step: event.target.value };
                          setDrafts(next);
                        }}
                      />
                    </div>
                  )}
                </div>
              ))
            )}
            <ActionMessage state={state} />
            {stepError ? (
              <p className="text-critical text-sm">{stepError}</p>
            ) : null}
            <div className="flex justify-between">
              <Button
                type="button"
                variant="secondary"
                onClick={() => setWizardStep(1)}
              >
                Back
              </Button>
              <Button type="submit" disabled={pending}>
                {pending ? "Starting…" : "Start recording"}
              </Button>
            </div>
          </form>
        </div>
      )}
    </div>
  );
}
