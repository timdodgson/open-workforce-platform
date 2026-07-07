'use client';

import { useState, useEffect } from 'react';
import { useSearchParams } from 'next/navigation';
import IntelligenceTabs, { TabId } from './IntelligenceTabs';
import OverviewTab from './OverviewTab';
import PolicyDecisionsTab from './PolicyDecisionsTab';
import SIValidationTab from './SIValidationTab';
import type { IntelligenceData } from '@/lib/intelligence-data';

import WorkerLearningDashboard from '../learning/WorkerLearningDashboard';
import WorkerDecisionDashboard from '../decisions/WorkerDecisionDashboard';
import FeatureImportanceDashboard from '../feature-importance/FeatureImportanceDashboard';
import PredictionExplorer from '../predictions/PredictionExplorer';
import WhatIfLab from '../what-if/WhatIfLab';
import AssistDashboard from '../assist/AssistDashboard';
import Card from '@/components/Card';

const VALID_TABS: TabId[] = [
  'overview', 'learning', 'model', 'predictions', 'decisions', 'what-if',
  'validation', 'policies', 'si-validation',
];

interface Props {
  data: IntelligenceData;
}

export default function IntelligenceShell({ data }: Props) {
  const searchParams = useSearchParams();
  const tabParam = searchParams.get('tab') as TabId | null;
  const initialTab = tabParam && VALID_TABS.includes(tabParam) ? tabParam : 'overview';
  const [activeTab, setActiveTab] = useState<TabId>(initialTab);

  useEffect(() => {
    if (tabParam && VALID_TABS.includes(tabParam) && tabParam !== activeTab) {
      setActiveTab(tabParam);
    }
  }, [tabParam, activeTab]);

  return (
    <div>
      <IntelligenceTabs activeTab={activeTab} onTabChange={setActiveTab} />

      {data.totalRuns > 0 && (
        <p className="text-[10px] text-gray-600 mb-2">
          Loaded {data.learning.length} learning rows from {data.runsScanned} of {data.totalRuns} runs
          {data.runsScanned < data.totalRuns ? ' (newest)' : ''}.
        </p>
      )}

      {activeTab === 'overview' && <OverviewTab />}

      {activeTab === 'learning' && (
        data.learning.length > 0
          ? <WorkerLearningDashboard records={data.learning} />
          : <EmptyState title="Learning" message="No worker learning data yet. Run experiments to generate worker_learning.csv." />
      )}

      {activeTab === 'model' && (
        data.model
          ? <FeatureImportanceDashboard model={data.model} />
          : <EmptyState title="Model" message="No trained model available. Run: python platform/ml/worker_model/train.py --data-dir platform/web/pfrs-lab/data/runs" />
      )}

      {activeTab === 'predictions' && (
        data.predictionsData && data.predictionsData.predictions.length > 0
          ? <PredictionExplorer data={data.predictionsData} />
          : <EmptyState title="Predictions" message="No prediction data. Generate worker_predictions.json via platform/ml/worker_model/predict.py" />
      )}

      {activeTab === 'decisions' && (
        data.decisions.length > 0
          ? <WorkerDecisionDashboard decisions={data.decisions} learning={data.decisionLearning} />
          : <EmptyState title="Decision Analysis" message="No worker decision data. Run tune-pfrs with --worker-decision-mode shadow or assist." />
      )}

      {activeTab === 'what-if' && (
        data.predictionsData && data.predictionsData.predictions.length > 0
          ? <WhatIfLab predictions={data.predictionsData.predictions} />
          : <EmptyState title="What-If Lab" message="No prediction data for simulation. Generate worker_predictions.json first." />
      )}

      {activeTab === 'validation' && (
        data.assistRecords.length > 0
          ? <AssistDashboard records={data.assistRecords} />
          : <EmptyState title="Assist Validation" message="No assist data. Use --worker-decision-mode assist or --policy-mode hybrid with --run-label." />
      )}

      {activeTab === 'policies' && (
        <PolicyDecisionsTab
          decisions={data.policyDecisions}
          learningReports={data.policyLearningReports}
          evalCount={data.policyEvalCount}
          registryVersionCount={data.registryVersionCount}
        />
      )}

      {activeTab === 'si-validation' && (
        <SIValidationTab
          completed={data.si2RunIds.length}
          totalExpected={240}
          si2RunIds={data.si2RunIds}
        />
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
