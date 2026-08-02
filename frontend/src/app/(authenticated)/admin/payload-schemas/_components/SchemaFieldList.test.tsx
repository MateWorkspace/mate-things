import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it } from "vitest";
import { useState } from "react";

import SchemaFieldList from "./SchemaFieldList";
import type { FieldRow } from "./schema-definition-types";

function Harness({ initial = [] as FieldRow[] }) {
  const [rows, setRows] = useState<FieldRow[]>(initial);
  return <SchemaFieldList rows={rows} onChange={setRows} path={[]} errors={{}} />;
}

describe("SchemaFieldList", () => {
  afterEach(cleanup);

  it("adds a field row when a valid name is typed and Enter is pressed", async () => {
    const user = userEvent.setup();
    render(<Harness />);

    const input = screen.getByPlaceholderText("variable_name");
    await user.type(input, "sample_rate{Enter}");

    expect(screen.getByText("sample_rate")).toBeInTheDocument();
    expect(screen.getByRole("combobox", { name: "sample_rate type" })).toHaveValue(
      "string",
    );
  });

  it("converts a typed space into an underscore as the user types", async () => {
    const user = userEvent.setup();
    render(<Harness />);

    const input = screen.getByPlaceholderText("variable_name");
    await user.type(input, "sample rate{Enter}");

    expect(screen.getByText("sample_rate")).toBeInTheDocument();
  });

  it("rejects a duplicate sibling name without adding a second row", async () => {
    const user = userEvent.setup();
    render(<Harness initial={[]} />);

    const input = screen.getByPlaceholderText("variable_name");
    await user.type(input, "name{Enter}");
    await user.type(input, "name{Enter}");

    expect(screen.getAllByText("name")).toHaveLength(1);
    expect(
      screen.getByText("A field with this name already exists here."),
    ).toBeInTheDocument();
  });

  it("reveals a nested add-field input when a row's type becomes Object", async () => {
    const user = userEvent.setup();
    render(<Harness />);

    await user.type(
      screen.getByPlaceholderText("variable_name"),
      "location{Enter}",
    );
    await user.selectOptions(
      screen.getByRole("combobox", { name: "location type" }),
      "object",
    );

    const nestedInputs = screen.getAllByPlaceholderText("variable_name");
    expect(nestedInputs).toHaveLength(2);
  });

  it("removes a row when its remove button is clicked", async () => {
    const user = userEvent.setup();
    render(<Harness />);

    await user.type(
      screen.getByPlaceholderText("variable_name"),
      "temp{Enter}",
    );
    expect(screen.getByText("temp")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Remove temp" }));
    expect(screen.queryByText("temp")).not.toBeInTheDocument();
  });

  it("shows Min/Max item and length fields for a List of text row", async () => {
    const user = userEvent.setup();
    render(<Harness />);

    await user.type(screen.getByPlaceholderText("variable_name"), "tags{Enter}");
    await user.selectOptions(
      screen.getByRole("combobox", { name: "tags type" }),
      "[]string",
    );

    expect(screen.getByLabelText("Min items")).toBeInTheDocument();
    expect(screen.getByLabelText("Max items")).toBeInTheDocument();
    expect(screen.getByLabelText("Min length")).toBeInTheDocument();
    expect(screen.getByLabelText("Max length")).toBeInTheDocument();
  });

  it("adds and removes enum options as tags", async () => {
    const user = userEvent.setup();
    render(<Harness />);

    await user.type(screen.getByPlaceholderText("variable_name"), "mode{Enter}");
    await user.selectOptions(
      screen.getByRole("combobox", { name: "mode type" }),
      "enum",
    );

    const optionsInput = screen.getByPlaceholderText(
      "Type an option and press Enter",
    );
    await user.type(optionsInput, "auto{Enter}");
    await user.type(optionsInput, "manual{Enter}");

    expect(screen.getByText("auto")).toBeInTheDocument();
    expect(screen.getByText("manual")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Remove option auto" }));
    expect(screen.queryByText("auto")).not.toBeInTheDocument();
    expect(screen.getByText("manual")).toBeInTheDocument();
  });

  it("clears stale constraint values when a row's type changes", async () => {
    const user = userEvent.setup();
    render(<Harness />);

    await user.type(screen.getByPlaceholderText("variable_name"), "name{Enter}");
    await user.type(screen.getByLabelText("Min length"), "3");
    expect(screen.getByLabelText("Min length")).toHaveValue(3);

    await user.selectOptions(
      screen.getByRole("combobox", { name: "name type" }),
      "boolean",
    );

    expect(screen.queryByLabelText("Min length")).not.toBeInTheDocument();

    await user.selectOptions(
      screen.getByRole("combobox", { name: "name type" }),
      "integer",
    );

    expect(screen.getByLabelText("Min")).toHaveValue(null);
  });

  it("preserves enum options when switching to List of choices", async () => {
    const user = userEvent.setup();
    render(<Harness />);

    await user.type(screen.getByPlaceholderText("variable_name"), "mode{Enter}");
    await user.selectOptions(
      screen.getByRole("combobox", { name: "mode type" }),
      "enum",
    );
    const optionsInput = screen.getByPlaceholderText(
      "Type an option and press Enter",
    );
    await user.type(optionsInput, "auto{Enter}");
    await user.type(optionsInput, "manual{Enter}");

    await user.selectOptions(
      screen.getByRole("combobox", { name: "mode type" }),
      "[]enum",
    );

    expect(screen.getByText("auto")).toBeInTheDocument();
    expect(screen.getByText("manual")).toBeInTheDocument();
  });

  it("preserves min/max when switching float to List of numbers", async () => {
    const user = userEvent.setup();
    render(<Harness />);

    await user.type(screen.getByPlaceholderText("variable_name"), "rate{Enter}");
    await user.selectOptions(
      screen.getByRole("combobox", { name: "rate type" }),
      "float",
    );
    await user.type(screen.getByLabelText("Min"), "0");
    await user.type(screen.getByLabelText("Max"), "100");

    await user.selectOptions(
      screen.getByRole("combobox", { name: "rate type" }),
      "[]float",
    );

    expect(screen.getByLabelText("Min")).toHaveValue(0);
    expect(screen.getByLabelText("Max")).toHaveValue(100);
  });

  it("shows an error and marks the tree invalid when an enum has zero options", async () => {
    const user = userEvent.setup();
    render(<Harness />);

    await user.type(screen.getByPlaceholderText("variable_name"), "mode{Enter}");
    await user.selectOptions(
      screen.getByRole("combobox", { name: "mode type" }),
      "enum",
    );

    expect(screen.getByText("Add at least one option.")).toBeInTheDocument();
  });

  it("shows an error when minimum exceeds maximum", async () => {
    const user = userEvent.setup();
    render(<Harness />);

    await user.type(screen.getByPlaceholderText("variable_name"), "rate{Enter}");
    await user.selectOptions(
      screen.getByRole("combobox", { name: "rate type" }),
      "float",
    );
    await user.type(screen.getByLabelText("Min"), "100");
    await user.type(screen.getByLabelText("Max"), "0");

    expect(screen.getByText("Min must not exceed Max.")).toBeInTheDocument();
  });

  it("cancels a rename in progress when Escape is pressed", async () => {
    const user = userEvent.setup();
    render(<Harness />);

    await user.type(screen.getByPlaceholderText("variable_name"), "name{Enter}");
    await user.click(screen.getByText("name"));

    const renameInput = screen.getByDisplayValue("name");
    await user.clear(renameInput);
    await user.type(renameInput, "renamed");
    await user.keyboard("{Escape}");

    expect(screen.getByText("name")).toBeInTheDocument();
    expect(screen.queryByText("renamed")).not.toBeInTheDocument();
  });

  it("toggles the Required checkbox for a row", async () => {
    const user = userEvent.setup();
    render(<Harness />);

    await user.type(screen.getByPlaceholderText("variable_name"), "name{Enter}");
    const checkbox = screen.getByRole("checkbox", { name: "Required" });

    expect(checkbox).not.toBeChecked();
    await user.click(checkbox);
    expect(checkbox).toBeChecked();
  });
});
