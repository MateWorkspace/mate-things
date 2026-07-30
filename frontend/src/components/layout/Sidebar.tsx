import type { NavigationGroup } from "@/config/navigation";

import SidebarGroup from "./SidebarGroup";

interface SidebarProps {
  navigation: readonly NavigationGroup[];
  pathname: string;
  collapsed: boolean;
  idPrefix: string;
  onNavigate?: () => void;
}

export default function Sidebar({
  navigation,
  pathname,
  collapsed,
  idPrefix,
  onNavigate,
}: SidebarProps) {
  return (
    <nav aria-label="Primary navigation" className="space-y-6 p-3">
      {navigation.map((group) => (
        <SidebarGroup
          key={group.label}
          group={group}
          pathname={pathname}
          collapsed={collapsed}
          idPrefix={idPrefix}
          onNavigate={onNavigate}
        />
      ))}
    </nav>
  );
}
