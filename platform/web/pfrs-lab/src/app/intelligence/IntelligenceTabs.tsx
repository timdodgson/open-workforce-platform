'use client';

import { INTELLIGENCE_TABS, type TabId } from './tab-registry';

interface IntelligenceTabsProps {
  activeTab: TabId;
  onTabChange: (tab: TabId) => void;
  onTabHover?: (tab: TabId) => void;
}

export default function IntelligenceTabs({ activeTab, onTabChange, onTabHover }: IntelligenceTabsProps) {
  return (
    <div className="flex flex-wrap gap-1 pb-1 border-b border-gray-700 mb-4">
      {INTELLIGENCE_TABS.map((tab) => (
        <button
          key={tab.id}
          onClick={() => onTabChange(tab.id)}
          onMouseEnter={() => onTabHover?.(tab.id)}
          onFocus={() => onTabHover?.(tab.id)}
          className={`flex items-center gap-1.5 px-3 py-2 text-xs rounded-t-lg whitespace-nowrap transition-colors ${
            activeTab === tab.id
              ? 'bg-gray-800 text-blue-400 border-b-2 border-blue-400'
              : 'text-gray-500 hover:text-gray-300 hover:bg-gray-800/50'
          }`}
        >
          <span>{tab.icon}</span>
          <span>{tab.label}</span>
        </button>
      ))}
    </div>
  );
}

export type { TabId } from './tab-registry';
