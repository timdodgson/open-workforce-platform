'use client';

import { useState } from 'react';

export type TabId =
  | 'overview' | 'learning' | 'model' | 'predictions' | 'decisions' | 'what-if'
  | 'validation' | 'policies' | 'si-validation'
  | 'continuous-learning' | 'promotion' | 'counterfactual';

interface Tab {
  id: TabId;
  label: string;
  icon: string;
}

const TABS: Tab[] = [
  { id: 'overview', label: 'Overview', icon: '🧠' },
  { id: 'learning', label: 'Worker Learning', icon: '📊' },
  { id: 'continuous-learning', label: 'Policy Learning', icon: '🔄' },
  { id: 'model', label: 'Model', icon: '🔬' },
  { id: 'predictions', label: 'Predictions', icon: '🧪' },
  { id: 'decisions', label: 'Decisions', icon: '🎯' },
  { id: 'counterfactual', label: 'Counterfactual', icon: '🔀' },
  { id: 'what-if', label: 'What-If', icon: '⚗️' },
  { id: 'validation', label: 'Assist Val.', icon: '✅' },
  { id: 'policies', label: 'Policies', icon: '📋' },
  { id: 'promotion', label: 'Promotion', icon: '🚀' },
  { id: 'si-validation', label: 'SI Val.', icon: '🧪' },
];

interface IntelligenceTabsProps {
  activeTab: TabId;
  onTabChange: (tab: TabId) => void;
  onTabHover?: (tab: TabId) => void;
}

export default function IntelligenceTabs({ activeTab, onTabChange, onTabHover }: IntelligenceTabsProps) {
  return (
    <div className="flex gap-1 overflow-x-auto pb-1 border-b border-gray-700 mb-4">
      {TABS.map(tab => (
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

export { TABS };
