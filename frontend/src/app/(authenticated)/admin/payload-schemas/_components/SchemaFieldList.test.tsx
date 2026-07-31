import { cleanup, render, screen, within } from "@testing-library/react";
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

  it("toggles the Required checkbox for a row", async () => {
    const user = userEvent.setup();
    render(<Harness />);

    await user.type(screen.getByPlaceholderText("variable_name"), "name{Enter}");
    const row = screen.getByText("name").closest("div")!;
    const checkbox = within(row.parentElement as HTMLElement).getByRole(
      "checkbox",
      { name: "Required" },
    );

    expect(checkbox).not.toBeChecked();
    await user.click(checkbox);
    expect(checkbox).toBeChecked();
  });
});
