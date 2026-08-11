import FilterBar from "@/components/collection/FilterBar";
import RoleSearchCombobox from "@/components/roles/RoleSearchCombobox";
import Input from "@/components/ui/input";

interface UserFiltersProps {
  canReadRoles: boolean;
  search?: string;
  roleId?: string;
  roleName?: string;
}

export default function UserFilters(props: UserFiltersProps) {
  return (
    <FilterBar clearHref="/admin/users">
      <label>
        <span className="text-foreground/70 mb-1.5 block text-xs font-semibold">
          Search
        </span>
        <Input
          name="search"
          defaultValue={props.search}
          placeholder="Search name or username"
          aria-label="Search users"
        />
      </label>
      {props.canReadRoles ? (
        <RoleSearchCombobox
          name="role_id"
          defaultRoleId={props.roleId}
          defaultRoleName={props.roleName ?? props.roleId}
        />
      ) : null}
    </FilterBar>
  );
}
