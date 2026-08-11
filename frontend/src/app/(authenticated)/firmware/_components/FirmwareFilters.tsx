import FilterBar from "@/components/collection/FilterBar";
import NodeClassSearchCombobox from "@/components/node-classes/NodeClassSearchCombobox";
import Input from "@/components/ui/input";

interface FirmwareFiltersProps {
  canReadNodeClasses: boolean;
  search?: string;
  nodeClassId?: string;
  nodeClassName?: string;
}

export default function FirmwareFilters(props: FirmwareFiltersProps) {
  return (
    <FilterBar clearHref="/firmware">
      <label>
        <span className="text-foreground/70 mb-1.5 block text-xs font-semibold">
          Search
        </span>
        <Input
          name="search"
          defaultValue={props.search}
          placeholder="Search by firmware name"
          aria-label="Search firmware"
        />
      </label>
      {props.canReadNodeClasses ? (
        <NodeClassSearchCombobox
          name="node_class_id"
          defaultNodeClassId={props.nodeClassId}
          defaultNodeClassName={props.nodeClassName ?? props.nodeClassId}
        />
      ) : null}
    </FilterBar>
  );
}
