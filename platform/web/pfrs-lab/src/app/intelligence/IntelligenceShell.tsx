'use client';

import { useState, useEffect, useCallback } from 'react';
import { useSearchParams, useRouter } from 'next/navigation';
import IntelligenceTabs, { TabId } from './IntelligenceTabs';
import OverviewTab from './OverviewTab';
import PolicyDecisionsTab from './PolicyDecisionsTab';
import SIValidationTab from './SIValidationTab';
import type { IntelligenceData, IntelligenceSummary } from '@/lib/intelligence-data';

const DEFAULT_PAGE_LIMIT = 100;

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

const PAGINATED_SECTIONS = new Set<Section>(['learning', 'decisions']);

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

const prefetched = new Set<string>();

function prefetchSection(section: Section, offset = 0) {
  const key = `${section}:${offset}`;
  if (prefetched.has(key) || typeof window === 'undefined') return;
  prefetched.add(key);
  const params = new URLSearchParams({ section });
  if (PAGINATED_SECTIONS.has(section)) {
    params.set('offset', String(offset));
    params.set('limit', String(DEFAULT_PAGE_LIMIT));
  }
  void fetch(`/api/intelligence?${params}`, { priority: 'low' } as RequestInit);
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
  const [hasMore, setHasMore] = useState<Partial<Record<Section, boolean>>>({});
  const [totalRows, setTotalRows] = useState<Partial<Record<Section, number>>>({});
  const [offsets, setOffsets] = useState<Partial<Record<Section, number>>>({});

  useEffect(() => {
    if (tabParam && VALID_TABS.includes(tabParam) && tabParam !== activeTab) {
      setActiveTab(tabParam);
    }
  }, [tabParam, activeTab]);

  const fetchSection = useCallback(async (section: Section, append = false) => {
    if (!append && loadedSections.has(section)) return;

    const offset = append ? (offsets[section] ?? 0) : 0;
    setLoadingSection(section);
    setError(null);

    try {
      const params = new URLSearchParams({ section });
      if (PAGINATED_SECTIONS.has(section)) {
        params.set('offset', String(offset));
        params.set('limit', String(DEFAULT_PAGE_LIMIT));
      }
      const res = await fetch(`/api/intelligence?${params}`);
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
        setLoadedSections((prev) => new Set(prev).add(section));
      } else if (section === 'learning') {
        setData((prev) => ({
          ...prev,
          ...json,
          learning: append ? [...prev.learning, ...(json.learning || [])] : (json.learning || []),
        }));
        setHasMore((prev) => ({ ...prev, learning: Boolean(json.hasMore) }));
        setTotalRows((prev) => ({ ...prev, learning: json.totalRows ?? json.learning?.length ?? 0 }));
        setOffsets((prev) => ({
          ...prev,
          learning: offset + (json.learning?.length ?? 0),
        }));
        if (!json.hasMore) setLoadedSections((prev) => new Set(prev).add(section));
      } else if (section === 'decisions') {
        setData((prev) => ({
          ...prev,
          ...json,
          decisions: append ? [...prev.decisions, ...(json.decisions || [])] : (json.decisions || []),
        }));
        setHasMore((prev) => ({ ...prev, decisions: Boolean(json.hasMore) }));
        setTotalRows((prev) => ({ ...prev, decisions: json.totalRows ?? json.decisions?.length ?? 0 }));
        setOffsets((prev) => ({
          ...prev,
          decisions: offset + (json.decisions?.length ?? 0),
        }));
        if (!json.hasMore) setLoadedSections((prev) => new Set(prev).add(section));
      } else {
        setData((prev) => ({ ...prev, ...json }));
        setLoadedSections((prev) => new Set(prev).add(section));
      }
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to load data');
    } finally {
      setLoadingSection(null);
    }
  }, [loadedSections, offsets]);

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
  const paginatedSection = sectionForTab && PAGINATED_SECTIONS.has(sectionForTab) ? sectionForTab : null;

  return (
    <div>
      <IntelligenceTabs activeTab={activeTab} onTabChange={handleTabChange} onTabHover={handleTabHover} />

      {summary && summary.totalRuns > 0 && (
        <p className="text-[10px] text-gray-500 mb-2">
          {paginatedSection && totalRows[paginatedSection]
            ? `Showing ${data[paginatedSection === 'learning' ? 'learning' : 'decisions'].length} of ${totalRows[paginatedSection]} rows from ${data.runsScanned || summary.runsScanned} of ${summary.totalRuns} runs`
            : data.learning.length > 0
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
        <>
          <IntelligenceTabPanel
            title="Worker Learning"
            loading={isLoading && data.learning.length === 0}
            empty={sectionLoaded && data.learning.length === 0}
            emptyMessage="No worker learning data yet. Run experiments to generate worker_learning.csv."
          >
            <WorkerLearningDashboard records={data.learning} />
          </IntelligenceTabPanel>
          {hasMore.learning && (
            <div className="text-center mb-4">
              <button
                type="button"
                onClick={() => void fetchSection('learning', true)}
                disabled={loadingSection === 'learning'}
                className="text-xs px-4 py-2 rounded bg-gray-800 text-gray-300 hover:bg-gray-700 disabled:opacity-50"
              >
                {loadingSection === 'learning' ? 'Loading…' : 'Load more'}
              </button>
            </div>
          )}
        </>
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
        <>
          <IntelligenceTabPanel
            title="Decision Analysis"
            loading={isLoading && data.decisions.length === 0}
            empty={sectionLoaded && data.decisions.length === 0}
            emptyMessage="No worker decision data. Run tune-pfrs with --worker-decision-mode shadow or assist."
          >
            <WorkerDecisionDashboard decisions={data.decisions} learning={data.decisionLearning} />
          </IntelligenceTabPanel>
          {hasMore.decisions && (
            <div className="text-center mb-4">
              <button
                type="button"
                onClick={() => void fetchSection('decisions', true)}
                disabled={loadingSection === 'decisions'}
                className="text-xs px-4 py-2 rounded bg-gray-800 text-gray-300 hover:bg-gray-700 disabled:opacity-50"
              >
                {loadingSection === 'decisions' ? 'Loading…' : 'Load more'}
              </button>
            </div>
          )}
        </>
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
