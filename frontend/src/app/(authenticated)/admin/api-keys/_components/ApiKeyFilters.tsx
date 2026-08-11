import FilterBar from "@/components/collection/FilterBar";
import Input from "@/components/ui/input";
import Select from "@/components/ui/select";
import type { ApiKeyStatus } from "@/lib/api/api-keys";

interface ApiKeyFiltersProps {
  search?: string;
  status: ApiKeyStatus;
}

export default function ApiKeyFilters({ search, status }: ApiKeyFiltersProps) {
  return (
    <FilterBar clearHref="/admin/api-keys">
      <label>
        <span className="text-foreground/70 mb-1.5 block text-xs font-semibold">
          Search
        </span>
        <Input
          name="search"
          defaultValue={search}
          placeholder="Name or username"
          aria-label="Search API keys"
        />
      </label>
      <div>
        <label
          className="text-foreground/70 mb-1.5 block text-xs font-semibold"
          htmlFor="api-key-status"
        >
          Status
        </label>
        <Select id="api-key-status" name="status" defaultValue={status}>
          <option value="any">Any status</option>
          <option value="active">Active</option>
          <option value="inactive">Inactive</option>
        </Select>
      </div>
    </FilterBar>
  );
}
