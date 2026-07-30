"use client";

import { usePathname, useRouter, useSearchParams } from "next/navigation";

import Tabs from "@/components/ui/tabs";

import { getNodeTabPanelId, NODE_TABS, type NodeTab } from "../_lib/node-tabs";

interface NodeTabsProps {
  activeTab: NodeTab;
  nodeId: string;
}

export default function NodeTabs({ activeTab, nodeId }: NodeTabsProps) {
  const pathname = usePathname();
  const router = useRouter();
  const searchParams = useSearchParams();
  const tabs = NODE_TABS.map((tab) => ({
    ...tab,
    panelId: getNodeTabPanelId(nodeId, tab.id),
  }));

  return (
    <Tabs
      activeTab={activeTab}
      ariaLabel="Node workspace"
      tabs={tabs}
      onChange={(tab) => {
        const nextSearchParams = new URLSearchParams(searchParams.toString());
        nextSearchParams.set("tab", tab);
        router.push(`${pathname}?${nextSearchParams.toString()}`);
      }}
    />
  );
}
