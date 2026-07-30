"use client";

import { useRef, type KeyboardEvent } from "react";

export interface Tab {
  id: string;
  label: string;
  panelId: string;
  disabled?: boolean;
}

interface TabsProps {
  tabs: readonly Tab[];
  activeTab: string;
  onChange: (id: string) => void;
  ariaLabel: string;
}

export function getTabId(panelId: string) {
  return `tab-${panelId}`;
}

export default function Tabs({
  tabs,
  activeTab,
  onChange,
  ariaLabel,
}: TabsProps) {
  const tabRefs = useRef(new Map<string, HTMLButtonElement>());

  const handleKeyDown = (
    event: KeyboardEvent<HTMLButtonElement>,
    currentTabId: string,
  ) => {
    const enabledTabs = tabs.filter((tab) => !tab.disabled);
    const currentIndex = enabledTabs.findIndex(
      (tab) => tab.id === currentTabId,
    );

    if (currentIndex === -1) {
      return;
    }

    let nextIndex: number | undefined;

    switch (event.key) {
      case "ArrowRight":
        nextIndex = (currentIndex + 1) % enabledTabs.length;
        break;
      case "ArrowLeft":
        nextIndex =
          (currentIndex - 1 + enabledTabs.length) % enabledTabs.length;
        break;
      case "Home":
        nextIndex = 0;
        break;
      case "End":
        nextIndex = enabledTabs.length - 1;
        break;
      default:
        return;
    }

    event.preventDefault();
    const nextTab = enabledTabs[nextIndex];
    tabRefs.current.get(nextTab.id)?.focus();
    onChange(nextTab.id);
  };

  return (
    <div
      role="tablist"
      aria-label={ariaLabel}
      className="border-border flex gap-1 overflow-x-auto border-b"
    >
      {tabs.map((tab) => {
        const isActive = tab.id === activeTab;

        return (
          <button
            key={tab.id}
            ref={(element) => {
              if (element) {
                tabRefs.current.set(tab.id, element);
              } else {
                tabRefs.current.delete(tab.id);
              }
            }}
            id={getTabId(tab.panelId)}
            type="button"
            role="tab"
            aria-selected={isActive}
            aria-controls={tab.panelId}
            tabIndex={isActive ? 0 : -1}
            disabled={tab.disabled}
            onClick={() => onChange(tab.id)}
            onKeyDown={(event) => handleKeyDown(event, tab.id)}
            className={`focus-visible:ring-focus focus-visible:ring-offset-background shrink-0 border-b-2 px-3 py-2.5 text-sm font-semibold transition-colors focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-50 ${
              isActive
                ? "border-primary text-primary"
                : "text-foreground/65 hover:text-foreground border-transparent"
            }`}
          >
            {tab.label}
          </button>
        );
      })}
    </div>
  );
}
