import FilterBar from "@/components/collection/FilterBar";
import Input from "@/components/ui/input";

interface PayloadSchemaFiltersProps {
  search?: string;
  stillValid: boolean;
}

export default function PayloadSchemaFilters({
  search,
  stillValid,
}: PayloadSchemaFiltersProps) {
  return (
    <FilterBar clearHref="/admin/payload-schemas">
      <label>
        <span className="text-foreground/70 mb-1.5 block text-xs font-semibold">
          Search
        </span>
        <Input
          name="search"
          defaultValue={search}
          placeholder="Search schema names"
          aria-label="Search schemas"
        />
      </label>
      <label className="flex items-center gap-2 self-end pb-2.5">
        <input
          type="checkbox"
          name="still_valid"
          value="1"
          defaultChecked={stillValid}
          className="accent-primary size-4"
        />
        <span className="text-foreground/70 text-xs font-semibold">
          Still valid only
        </span>
      </label>
    </FilterBar>
  );
}
