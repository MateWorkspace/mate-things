import FilterBar from "@/components/collection/FilterBar";
import FirmwareSearchCombobox from "@/components/firmwares/FirmwareSearchCombobox";
import NodeClassSearchCombobox from "@/components/node-classes/NodeClassSearchCombobox";
import Input from "@/components/ui/input";

interface NodeFiltersProps {
  search?: string;
  nodeClassId?: string;
  nodeClassName?: string;
  firmwareId?: string;
  firmwareName?: string;
  showClassFilter: boolean;
  showFirmwareFilter: boolean;
}

export default function NodeFilters({
  search,
  nodeClassId,
  nodeClassName,
  firmwareId,
  firmwareName,
  showClassFilter,
  showFirmwareFilter,
}: NodeFiltersProps) {
  return (
    <FilterBar clearHref="/nodes">
      <label>
        <span className="text-foreground/70 mb-1.5 block text-xs font-semibold">
          Search
        </span>
        <Input
          name="search"
          defaultValue={search}
          placeholder="Name or device ID"
          aria-label="Search nodes"
        />
      </label>
      {showClassFilter ? (
        <NodeClassSearchCombobox
          name="node_class_id"
          defaultNodeClassId={nodeClassId}
          defaultNodeClassName={nodeClassName ?? nodeClassId}
        />
      ) : null}
      {showFirmwareFilter ? (
        <FirmwareSearchCombobox
          name="firmware_id"
          defaultFirmwareId={firmwareId}
          defaultFirmwareName={firmwareName ?? firmwareId}
        />
      ) : null}
    </FilterBar>
  );
}
