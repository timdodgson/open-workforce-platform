'use client';

import { useState } from 'react';
import IntelligenceTabs, { TabId } from './IntelligenceTabs';
import OverviewTab from './OverviewTab';
import nextDynamic from 'next/dynamic';

// Lazy-load tab contents to avoid loading all data on initial render.
const LearningTab = nextDynamic(() => import('./tabs/LearningTab'), { ssr: false });
const ModelTab = nextDynamic(() => import('./tabs/ModelTab'), { ssr: false });
const PredictionsTab = nextDynamic(() => import('./tabs/PredictionsTab'), { ssr: false });
const DecisionsTab = nextDynamic(() => import('./tabs/DecisionsTab'), { ssr: false });
const WhatIfTab = nextDynamic(() => import('./tabs/WhatIfTab'), { ssr: false });
const ValidationTab = nextDynamic(() => import('./tabs/ValidationTab'), { ssr: false });

export default function IntelligenceShell() {
  const [activeTab, setActiveTab] = useState<TabId>('overview');

  return (
    <div>
      <IntelligenceTabs activeTab={activeTab} onTabChange={setActiveTab} />
      {activeTab === 'overview' && <OverviewTab />}
      {activeTab === 'learning' && <LearningTab />}
      {activeTab === 'model' && <ModelTab />}
      {activeTab === 'predictions' && <PredictionsTab />}
      {activeTab === 'decisions' && <DecisionsTab />}
      {activeTab === 'what-if' && <WhatIfTab />}
      {activeTab === 'validation' && <ValidationTab />}
    </div>
  );
}
