'use client';

import { useState, useEffect, useCallback } from 'react';
import { useSearchParams, useRouter } from 'next/navigation';
import IntelligenceTabs, { TabId } from './IntelligenceTabs';
import OverviewTab from './OverviewTab';
import PolicyDecisionsTab from './PolicyDecisionsTab';
import SIValidationTab from './SIValidationTab';
import type { IntelligenceData, IntelligenceSummary } from '@/lib/intelligence-data';

import WorkerLearningDashboard from '../learning/WorkerLearningDashboard';
import WorkerDecisionDashboard from '../decisions/WorkerDecisionDashboard';
import FeatureImportanceDashboard from '../feature-importance/FeatureImportanceDashboard';
import { PredictionsTabClient, WhatIfTabClient } from './PredictionsTabClient';
import ContinuousLearningTab from './ContinuousLearningTab';
import PromotionTab from './PromotionTab';
import CounterfactualTab from './CounterfactualTab';
import AssistDashboard from '../assist/AssistDashboard';
import IntelligenceTabPanel from '@/components/IntelligenceTabPanel';

const VALID_TABS: TabId[] = [
  'overview', 'learning', 'continuous-learning', 'model', 'predictions', 'decisions',
  'counterfactual', 'what-if', 'validation', 'policies', 'promotion', 'si-validation',
];

type Section = 'summary' | 'learning' | 'decisions' | 'model' | 'assist' | 'policies'
  | 'continuous-learning' | 'promotion' | 'counterfactual';

const TAB_SECTION: Partial<Record<TabId, Section>> = {
  learning: 'learning',
  'continuous-learning': 'continuous-learning',
  model: 'model',
  decisions: 'decisions',
  counterfactual: 'counterfactual',
  validation: 'assist',
  policies: 'policies',
  promotion: 'promotion',
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
  const router = useRouter();
  const searchParams = useSearchParams();
  const tabParam = searchParams.get('tab') as TabId | null;
  const initialTab = tabParam && VALID_TABS.includes(tabParam) ? tabParam : 'overview';
  const [activeTab, setActiveTab] = useState<TabId>(initialTab);
  const [summary, setSummary] = useState<IntelligenceSummary | null>(null);
  const [data, setData] = useState<IntelligenceData>(emptyData);
  const [loadingSection, setLoadingSection] = useState<Section | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loadedSections, setLoadedSections] = useState<Set<Section>>(new Set());

  useEffect(() => {
    if (tabParam && VALID_TABS.includes(tabParam) && tabParam !== activeTab) {
      setActiveTab(tabParam);
    }
  }, [tabParam, activeTab]);

  const fetchSection = useCallback(async (section: Section) => {
    if (loadedSections.has(section)) return;
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
      setLoadedSections((prev) => new Set(prev).add(section));
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to load data');
    } finally {
      setLoadingSection(null);
    }
  }, [loadedSections]);

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

  const handleTabChange = useCallback((tab: TabId) => {
    setActiveTab(tab);
    const params = new URLSearchParams(searchParams.toString());
    if (tab === 'overview') {
      params.delete('tab');
    } else {
      params.set('tab', tab);
    }
    const qs = params.toString();
    router.replace(qs ? `/intelligence?${qs}` : '/intelligence', { scroll: false });
  }, [router, searchParams]);

  const sectionForTab = TAB_SECTION[activeTab];
  const isLoading = sectionForTab ? loadingSection === sectionForTab : false;
  const sectionLoaded = sectionForTab ? loadedSections.has(sectionForTab) : true;

  return (
    <div>
      <IntelligenceTabs activeTab={activeTab} onTabChange={handleTabChange} onTabHover={handleTabHover} />

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

      {activeTab === 'overview' && <OverviewTab />}

      {activeTab === 'learning' && (
        <IntelligenceTabPanel
          title="Worker Learning"
          loading={isLoading || !sectionLoaded}
          empty={data.learning.length === 0}
          emptyMessage="No worker learning data yet. Run experiments to generate worker_learning.csv."
        >
          <WorkerLearningDashboard records={data.learning} />
        </IntelligenceTabPanel>
      )}

      {activeTab === 'model' && (
        <IntelligenceTabPanel
          title="Model"
          loading={isLoading || !sectionLoaded}
          empty={!data.model}
          emptyMessage="No trained model available. Run: python platform/ml/worker_model/train.py --data-dir platform/web/pfrs-lab/data/runs"
        >
          {data.model && <FeatureImportanceDashboard model={data.model} />}
        </IntelligenceTabPanel>
      )}

      {activeTab === 'predictions' && <PredictionsTabClient />}

      {activeTab === 'decisions' && (
        <IntelligenceTabPanel
          title="Decision Analysis"
          loading={isLoading || !sectionLoaded}
          empty={data.decisions.length === 0}
          emptyMessage="No worker decision data. Run tune-pfrs with --worker-decision-mode shadow or assist."
        >
          <WorkerDecisionDashboard decisions={data.decisions} learning={data.decisionLearning} />
        </IntelligenceTabPanel>
      )}

      {activeTab === 'what-if' && <WhatIfTabClient />}

      {activeTab === 'validation' && (
        <IntelligenceTabPanel
          title="Assist Validation"
          loading={isLoading || !sectionLoaded}
          empty={data.assistRecords.length === 0}
          emptyMessage="No assist data. Use --worker-decision-mode assist or --policy-mode hybrid with --run-label."
        >
          <AssistDashboard records={data.assistRecords} />
        </IntelligenceTabPanel>
      )}

      {activeTab === 'policies' && (
        <PolicyDecisionsTab
          decisions={data.policyDecisions}
          learningReports={data.policyLearningReports}
          evalCount={data.policyEvalCount}
          registryVersionCount={data.registryVersionCount}
          loading={isLoading || !sectionLoaded}
        />
      )}

      {activeTab === 'si-validation' && (
        <SIValidationTab
          completed={(summary?.si2RunIds ?? data.si2RunIds).length}
          totalExpected={240}
          si2RunIds={summary?.si2RunIds ?? data.si2RunIds}
        />
      )}

      {activeTab === 'continuous-learning' && (
        <ContinuousLearningTab
          state={data.continuousLearning ?? null}
          reports={data.policyLearningReports}
          loading={isLoading || !sectionLoaded}
        />
      )}

      {activeTab === 'promotion' && (
        <PromotionTab versions={data.policyVersions ?? []} loading={isLoading || !sectionLoaded} />
      )}

      {activeTab === 'counterfactual' && (
        <CounterfactualTab summary={data.counterfactual ?? null} loading={isLoading || !sectionLoaded} />
      )}
    </div>
  );
}
