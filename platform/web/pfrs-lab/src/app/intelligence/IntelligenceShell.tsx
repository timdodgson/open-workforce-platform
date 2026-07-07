'use client';

import { useState, useEffect } from 'react';
import { useSearchParams } from 'next/navigation';
import IntelligenceTabs, { TabId } from './IntelligenceTabs';
import OverviewTab from './OverviewTab';
import type { IntelligenceData } from './page';

// Import existing dashboard components directly.
import WorkerLearningDashboard from '../learning/WorkerLearningDashboard';
import WorkerDecisionDashboard from '../decisions/WorkerDecisionDashboard';
import FeatureImportanceDashboard from '../feature-importance/FeatureImportanceDashboard';
import WhatIfLab from '../what-if/WhatIfLab';
import AssistDashboard from '../assist/AssistDashboard';
import Card from '@/components/Card';

const VALID_TABS: TabId[] = ['overview', 'learning', 'model', 'predictions', 'decisions', 'what-if', 'validation'];

interface Props {
  data: IntelligenceData;
}

export default function IntelligenceShell({ data }: Props) {
  const searchParams = useSearchParams();
  const tabParam = searchParams.get('tab') as TabId | null;
  const initialTab = tabParam && VALID_TABS.includes(tabParam) ? tabParam : 'overview';
  const [activeTab, setActiveTab] = useState<TabId>(initialTab);

  // Sync tab with URL param on changes.
  useEffect(() => {
    if (tabParam && VALID_TABS.includes(tabParam) && tabParam !== activeTab) {
      setActiveTab(tabParam);
    }
  }, [tabParam]);

  return (
    <div>
      <IntelligenceTabs activeTab={activeTab} onTabChange={setActiveTab} />

      {activeTab === 'overview' && <OverviewTab />}

      {activeTab === 'learning' && (
        data.learning.length > 0
          ? <WorkerLearningDashboard records={data.learning} />
          : <EmptyState title="Learning" message="No worker learning data yet. Run experiments to generate worker_learning.csv." />
      )}

      {activeTab === 'model' && (
        data.model
          ? <FeatureImportanceDashboard model={data.model} />
          : <EmptyState title="Model" message="No trained model available. Run the ML pipeline to generate worker_model.json." />
      )}

      {activeTab === 'predictions' && (
        data.predictions.length > 0
          ? <WhatIfLab predictions={data.predictions} />
          : <EmptyState title="Predictions" message="No prediction data yet. Generate predictions with the ML pipeline." />
      )}

      {activeTab === 'decisions' && (
        data.decisions.length > 0
          ? <WorkerDecisionDashboard decisions={data.decisions} learning={data.decisionLearning} />
          : <EmptyState title="Decision Analysis" message="No worker decision data yet. Run experiments with --worker-decision-mode shadow." />
      )}

      {activeTab === 'what-if' && (
        data.predictions.length > 0
          ? <WhatIfLab predictions={data.predictions} />
          : <EmptyState title="What-If Lab" message="No prediction data for simulation. Generate predictions first." />
      )}

      {activeTab === 'validation' && (
        data.assistRecords.length > 0
          ? <AssistDashboard records={data.assistRecords} />
          : <EmptyState title="Assist Validation" message="No assist mode data yet. Run experiments with --worker-decision-mode shadow|assist|adaptive." />
      )}
    </div>
  );
}

function EmptyState({ title, message }: { title: string; message: string }) {
  return (
    <Card title={title}>
      <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500">
        <p className="text-xs">{message}</p>
      </div>
    </Card>
  );
}
