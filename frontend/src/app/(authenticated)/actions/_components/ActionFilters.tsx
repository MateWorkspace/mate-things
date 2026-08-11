import FilterBar from "@/components/collection/FilterBar";
import NodeClassSearchCombobox from "@/components/node-classes/NodeClassSearchCombobox";
import Input from "@/components/ui/input";

interface ActionFiltersProps {
  canReadNodeClasses: boolean;
  search?: string;
  nodeClassId?: string;
  nodeClassName?: string;
}

export default function ActionFilters(props: ActionFiltersProps) {
  return (
    <FilterBar clearHref="/actions">
      <label>
        <span className="text-foreground/70 mb-1.5 block text-xs font-semibold">
          Search
        </span>
        <Input
          name="search"
          defaultValue={props.search}
          placeholder="Search actions"
          aria-label="Search actions"
        />
      </label>
      {props.canReadNodeClasses ? (
        <NodeClassSearchCombobox
          name="node_class_id"
          defaultNodeClassId={props.nodeClassId}
          defaultNodeClassName={props.nodeClassName ?? props.nodeClassId}
        />
      ) : (
        <label>
          <span className="text-foreground/70 mb-1.5 block text-xs font-semibold">
            Node class ID
          </span>
          <Input
            aria-label="Node class ID"
            name="node_class_id"
            readOnly
            value={props.nodeClassId ?? ""}
          />
        </label>
      )}
    </FilterBar>
  );
}
