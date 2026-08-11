import FilterBar from "@/components/collection/FilterBar";
import Input from "@/components/ui/input";

interface NodeClassFiltersProps {
  search?: string;
}

export default function NodeClassFilters({ search }: NodeClassFiltersProps) {
  return (
    <FilterBar clearHref="/node-classes">
      <label>
        <span className="text-foreground/70 mb-1.5 block text-xs font-semibold">
          Search
        </span>
        <Input
          name="search"
          defaultValue={search}
          placeholder="Search by name or description"
          aria-label="Search node classes"
        />
      </label>
    </FilterBar>
  );
}
