import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import NodeConfigForm from "./NodeConfigForm";

vi.mock("../_lib/actions", () => ({
  saveNodeConfigAction: vi.fn(),
}));

const PARAMETERS = [
  { key: "mqtt_host", value_type: "string" },
  { key: "sample_rate", value_type: "uint32" },
  { key: "report_enabled", value_type: "bool" },
];

describe("NodeConfigForm", () => {
  afterEach(cleanup);

  it("maps the backend's supported parameter types to native controls", () => {
    render(
      <NodeConfigForm
        canSet
        nodeId="node-1"
        parameters={PARAMETERS}
        values={[
          { key: "mqtt_host", value: "broker.example.com" },
          { key: "sample_rate", value: "30" },
          { key: "report_enabled", value: "true" },
        ]}
      />,
    );

    expect(screen.getByLabelText("mqtt host")).toHaveAttribute("type", "text");
    expect(screen.getByLabelText("mqtt host")).not.toBeRequired();
    expect(screen.getByLabelText("sample rate")).toHaveAttribute(
      "type",
      "number",
    );
    expect(screen.getByLabelText("sample rate")).toBeRequired();
    expect(screen.getByLabelText("sample rate")).toHaveAttribute("min", "0");
    expect(screen.getByLabelText("sample rate")).toHaveAttribute(
      "max",
      "4294967295",
    );
    expect(screen.getByLabelText("report enabled")).toHaveRole("combobox");
    expect(screen.getByLabelText("report enabled")).toBeRequired();
    expect(screen.getAllByRole("button", { name: "Save value" })).toHaveLength(
      3,
    );
  });

  it("keeps values readable but removes mutation controls without node_config:set", () => {
    render(
      <NodeConfigForm
        canSet={false}
        nodeId="node-1"
        parameters={PARAMETERS}
        values={[{ key: "mqtt_host", value: "broker.example.com" }]}
      />,
    );

    expect(screen.getByDisplayValue("broker.example.com")).toBeDisabled();
    expect(
      screen.queryByRole("button", { name: "Save value" }),
    ).not.toBeInTheDocument();
  });

  it("distinguishes an unset value from a saved value without an update time", () => {
    render(
      <NodeConfigForm
        canSet
        nodeId="node-1"
        parameters={[
          { key: "saved_without_time", value_type: "string" },
          { key: "missing_value", value_type: "string" },
        ]}
        values={[{ key: "saved_without_time", value: "configured" }]}
      />,
    );

    expect(screen.getByText("Set — update time unavailable")).toBeVisible();
    expect(screen.getByText("Not set")).toBeVisible();
  });
});
