'use client';

import { useState } from 'react';

export type TabId = 'overview' | 'learning' | 'model' | 'predictions' | 'decisions' | 'what-if' | 'validation' | 'policies' | 'si-validation';

interface Tab {
  id: TabId;
  label: string;
  icon: string;
}

const TABS: Tab[] = [
  { id: 'overview', label: 'Overview', icon: '🧠' },
  { id: 'learning', label: 'Learning', icon: '📊' },
  { id: 'model', label: 'Model', icon: '🔬' },
  { id: 'predictions', label: 'Predictions', icon: '🧪' },
  { id: 'decisions', label: 'Decision Analysis', icon: '🎯' },
  { id: 'what-if', label: 'What-If Lab', icon: '⚗️' },
  { id: 'validation', label: 'Assist Validation', icon: '✅' },
  { id: 'policies', label: 'Policies', icon: '📋' },
  { id: 'si-validation', label: 'SI Validation', icon: '🧪' },
];

interface IntelligenceTabsProps {
  activeTab: TabId;
  onTabChange: (tab: TabId) => void;
}

export default function IntelligenceTabs({ activeTab, onTabChange }: IntelligenceTabsProps) {
  return (
    <div className="flex gap-1 overflow-x-auto pb-1 border-b border-gray-700 mb-4">
      {TABS.map(tab => (
        <button
          key={tab.id}
          onClick={() => onTabChange(tab.id)}
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
