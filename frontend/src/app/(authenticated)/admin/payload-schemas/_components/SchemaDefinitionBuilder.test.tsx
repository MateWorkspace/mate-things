import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";

import SchemaDefinitionBuilder from "./SchemaDefinitionBuilder";

describe("SchemaDefinitionBuilder", () => {
  afterEach(cleanup);

  it("starts empty with a hidden input carrying an empty object definition", () => {
    const { container } = render(<SchemaDefinitionBuilder name="definition" />);
    const hidden = container.querySelector(
      'input[type="hidden"][name="definition"]',
    ) as HTMLInputElement;
    expect(JSON.parse(hidden.value)).toEqual({ type: "object", properties: {} });
  });

  it("pre-populates rows from an existing representable definition", () => {
    render(
      <SchemaDefinitionBuilder
        name="definition"
        defaultValue={{
          type: "object",
          properties: { sample_rate: { type: "float" } },
        }}
      />,
    );
    expect(screen.getByText("sample_rate")).toBeInTheDocument();
  });

  it("updates the hidden input's JSON as fields are added", async () => {
    const user = userEvent.setup();
    const { container } = render(<SchemaDefinitionBuilder name="definition" />);

    await user.type(screen.getByPlaceholderText("variable_name"), "name{Enter}");

    const hidden = container.querySelector(
      'input[type="hidden"][name="definition"]',
    ) as HTMLInputElement;
    expect(JSON.parse(hidden.value)).toEqual({
      type: "object",
      properties: { name: { type: "string" } },
    });
  });

  it("shows a read-only fallback and reports unrepresentable for a definition it can't parse", () => {
    const onRepresentableChange = vi.fn();
    render(
      <SchemaDefinitionBuilder
        name="definition"
        defaultValue={{ type: "date" }}
        onRepresentableChange={onRepresentableChange}
      />,
    );

    expect(
      screen.getByText(/guided editor can't represent/i),
    ).toBeInTheDocument();
    expect(screen.queryByPlaceholderText("variable_name")).not.toBeInTheDocument();
    expect(onRepresentableChange).toHaveBeenCalledWith(false);
  });

  it("reports representable for a normal definition", () => {
    const onRepresentableChange = vi.fn();
    render(
      <SchemaDefinitionBuilder
        name="definition"
        onRepresentableChange={onRepresentableChange}
      />,
    );
    expect(onRepresentableChange).toHaveBeenCalledWith(true);
  });
});
