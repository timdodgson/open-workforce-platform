'use client';

import { useState, useEffect, useCallback, useRef } from 'react';
import { useSearchParams } from 'next/navigation';
import IntelligenceTabs, { TabId } from './IntelligenceTabs';
import OverviewTab from './OverviewTab';
import PolicyDecisionsTab from './PolicyDecisionsTab';
import SIValidationTab from './SIValidationTab';
import type { IntelligenceData, IntelligenceSummary } from '@/lib/intelligence-data';

import WorkerLearningDashboard from '../learning/WorkerLearningDashboard';
import WorkerDecisionDashboard from '../decisions/WorkerDecisionDashboard';
import FeatureImportanceDashboard from '../feature-importance/FeatureImportanceDashboard';
import { PredictionsTabClient, WhatIfTabClient } from './PredictionsTabClient';
import AssistDashboard from '../assist/AssistDashboard';
import Card from '@/components/Card';

const VALID_TABS: TabId[] = [
  'overview', 'learning', 'model', 'predictions', 'decisions', 'what-if',
  'validation', 'policies', 'si-validation',
];

type Section = 'summary' | 'learning' | 'decisions' | 'model' | 'assist' | 'policies';

const TAB_SECTION: Partial<Record<TabId, Section>> = {
  learning: 'learning',
  model: 'model',
  decisions: 'decisions',
  validation: 'assist',
  policies: 'policies',
};

const prefetched = new Set<Section>();

function prefetchSection(section: Section) {
  if (prefetched.has(section) || typeof window === 'undefined') return;
  prefetched.add(section);
  void fetch(`/api/intelligence?section=${section}`, { priority: 'low' } as RequestInit);
}

const emptyData: IntelligenceData = {
  learning: [],
  decisions: [],
  decisionLearning: [],
  model: null,
  predictionsData: null,
  assistRecords: [],
  policyDecisions: [],
  policyLearningReports: [],
  policyEvalCount: 0,
  registryVersionCount: 0,
  si2RunIds: [],
  runsScanned: 0,
  totalRuns: 0,
};

export default function IntelligenceShell() {
  const searchParams = useSearchParams();
  const tabParam = searchParams.get('tab') as TabId | null;
  const initialTab = tabParam && VALID_TABS.includes(tabParam) ? tabParam : 'overview';
  const [activeTab, setActiveTab] = useState<TabId>(initialTab);
  const [summary, setSummary] = useState<IntelligenceSummary | null>(null);
  const [data, setData] = useState<IntelligenceData>(emptyData);
  const [loadingSection, setLoadingSection] = useState<Section | null>(null);
  const [error, setError] = useState<string | null>(null);
  const loadedSections = useRef<Set<Section>>(new Set());

  useEffect(() => {
    if (tabParam && VALID_TABS.includes(tabParam) && tabParam !== activeTab) {
      setActiveTab(tabParam);
    }
  }, [tabParam, activeTab]);

  const fetchSection = useCallback(async (section: Section) => {
    if (loadedSections.current.has(section)) return;
    setLoadingSection(section);
    setError(null);
    try {
      const res = await fetch(`/api/intelligence?section=${section}`);
      if (!res.ok) throw new Error(`Failed to load ${section} (${res.status})`);
      const json = await res.json();
      if (section === 'summary' && json.summary) {
        setSummary(json.summary);
        setData((prev) => ({
          ...prev,
          totalRuns: json.summary.totalRuns,
          runsScanned: json.summary.runsScanned,
          si2RunIds: json.summary.si2RunIds,
          registryVersionCount: json.summary.registryVersionCount,
          policyEvalCount: json.summary.policyEvalCount,
        }));
      } else {
        setData((prev) => ({ ...prev, ...json }));
      }
      loadedSections.current.add(section);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to load data');
    } finally {
      setLoadingSection(null);
    }
  }, []);

  useEffect(() => {
    void fetchSection('summary');
  }, [fetchSection]);

  useEffect(() => {
    const section = TAB_SECTION[activeTab];
    if (section) void fetchSection(section);
  }, [activeTab, fetchSection]);

  const handleTabHover = useCallback((tab: TabId) => {
    const section = TAB_SECTION[tab];
    if (section) prefetchSection(section);
    if (tab === 'si-validation') prefetchSection('summary');
  }, []);

  const sectionForTab = TAB_SECTION[activeTab];
  const isLoading = sectionForTab ? loadingSection === sectionForTab : false;

  return (
    <div>
      <IntelligenceTabs activeTab={activeTab} onTabChange={setActiveTab} onTabHover={handleTabHover} />

      {summary && summary.totalRuns > 0 && (
        <p className="text-[10px] text-gray-600 mb-2">
          {data.learning.length > 0
            ? `Loaded ${data.learning.length} learning rows from ${data.runsScanned} of ${summary.totalRuns} runs`
            : `${summary.totalRuns} runs in storage`}
          {data.runsScanned > 0 && data.runsScanned < summary.totalRuns ? ' (newest)' : ''}.
        </p>
      )}

      {error && (
        <p className="text-xs text-red-400 mb-2">{error}</p>
      )}

      {isLoading && (
        <p className="text-xs text-gray-500 mb-2">Loading…</p>
      )}

      {activeTab === 'overview' && <OverviewTab />}

      {activeTab === 'learning' && (
        data.learning.length > 0
          ? <WorkerLearningDashboard records={data.learning} />
          : <EmptyState title="Learning" message="No worker learning data yet. Run experiments to generate worker_learning.csv." loading={isLoading} />
      )}

      {activeTab === 'model' && (
        data.model
          ? <FeatureImportanceDashboard model={data.model} />
          : <EmptyState title="Model" message="No trained model available. Run: python platform/ml/worker_model/train.py --data-dir platform/web/pfrs-lab/data/runs" loading={isLoading} />
      )}

      {activeTab === 'predictions' && <PredictionsTabClient />}

      {activeTab === 'decisions' && (
        data.decisions.length > 0
          ? <WorkerDecisionDashboard decisions={data.decisions} learning={data.decisionLearning} />
          : <EmptyState title="Decision Analysis" message="No worker decision data. Run tune-pfrs with --worker-decision-mode shadow or assist." loading={isLoading} />
      )}

      {activeTab === 'what-if' && <WhatIfTabClient />}

      {activeTab === 'validation' && (
        data.assistRecords.length > 0
          ? <AssistDashboard records={data.assistRecords} />
          : <EmptyState title="Assist Validation" message="No assist data. Use --worker-decision-mode assist or --policy-mode hybrid with --run-label." loading={isLoading} />
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
          completed={(summary?.si2RunIds ?? data.si2RunIds).length}
          totalExpected={240}
          si2RunIds={summary?.si2RunIds ?? data.si2RunIds}
        />
      )}
    </div>
  );
}

function EmptyState({ title, message, loading }: { title: string; message: string; loading?: boolean }) {
  return (
    <Card title={title}>
      <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500">
        <p className="text-xs">{loading ? 'Loading…' : message}</p>
      </div>
    </Card>
  );
}
