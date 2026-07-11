import { cached } from '@/lib/cache';
import {
  loadLearningArtifact,
  loadPolicyArtifact,
  loadSummaryArtifact,
} from '@/lib/intelligence-artifacts';
import { getStorageProvider, type StorageProvider } from '@/lib/storage';
import { parseCounterfactualCSV } from '@/lib/parsers/intelligence';
import type { CounterfactualRow, PolicyVersion, WorkerModel } from '@/lib/types/intelligence';
import { DEFAULT_PAGE_LIMIT, RUN_BATCH_SIZE, SECTION_CACHE_TTL } from './constants';
import { emptyIngestAcc, ingestBatches } from './ingest';
import { loadContinuousLearningState, paginateSlice, summarizeCounterfactual } from './helpers';
import { resolveRunIdSets } from './resolve-runs';
import type {
  IntelligenceData,
  IntelligenceLoadOptions,
  IntelligencePageMeta,
  IntelligenceSection,
  IntelligenceSummary,
} from './types';

/** Load one intelligence slice — keeps Lambda responses under the 6MB cap. */
export async function loadIntelligenceSection(
  section: IntelligenceSection,
  storage?: StorageProvider,
  options?: IntelligenceLoadOptions,
): Promise<Partial<IntelligenceData> & { summary?: IntelligenceSummary } & IntelligencePageMeta> {
  const offset = options?.offset ?? 0;
  const limit = options?.limit ?? DEFAULT_PAGE_LIMIT;
  const cacheKey = `intel:${section}:${offset}:${limit}`;
  const ttl = SECTION_CACHE_TTL[section] ?? 60_000;
  return cached(
    cacheKey,
    () => loadIntelligenceSectionUncached(section, storage, offset, limit),
    ttl,
  );
}

async function loadIntelligenceSectionUncached(
  section: IntelligenceSection,
  storage: StorageProvider | undefined,
  offset: number,
  limit: number,
): Promise<Partial<IntelligenceData> & { summary?: IntelligenceSummary } & IntelligencePageMeta> {
  const store = storage ?? getStorageProvider();
  const { allRunIds, si2RunIds, learningRunIds, policyRunIds, counterfactualRunIds } =
    await resolveRunIdSets(store);

  if (section === 'summary') {
    const artifact = await loadSummaryArtifact(store);
    if (artifact) {
      return {
        summary: artifact.summary,
        totalRuns: artifact.totalRuns,
        artifactGeneratedAt: artifact.generatedAt,
        dataSource: 'artifact',
      };
    }
    const [modelContent, registryContent] = await Promise.all([
      store.readRootFile('worker_model.json'),
      store.readRootFile('policy_registry.json'),
    ]);
    let registryVersionCount = 0;
    if (registryContent) {
      try {
        const reg = JSON.parse(registryContent);
        registryVersionCount = (reg.versions || []).length;
      } catch { /* graceful */ }
    }
    return {
      summary: {
        totalRuns: allRunIds.length,
        runsScanned: new Set([...learningRunIds, ...policyRunIds]).size,
        si2RunIds,
        registryVersionCount,
        policyEvalCount: 0,
        hasModel: Boolean(modelContent),
      },
      dataSource: 'live',
    };
  }

  if (section === 'continuous-learning') {
    const policyArtifact = await loadPolicyArtifact(store);
    if (policyArtifact) {
      return {
        continuousLearning: policyArtifact.continuousLearning,
        policyLearningReports: policyArtifact.policyLearningReports,
        runsScanned: policyArtifact.runsScanned,
        totalRuns: policyArtifact.totalRuns,
        artifactGeneratedAt: policyArtifact.generatedAt,
        dataSource: 'artifact',
      };
    }
    const acc = emptyIngestAcc();
    const [state] = await Promise.all([
      loadContinuousLearningState(store),
      ingestBatches(store, policyRunIds, 'policy', acc),
    ]);
    return {
      continuousLearning: state,
      policyLearningReports: acc.policyLearningReports,
      runsScanned: policyRunIds.length,
      totalRuns: allRunIds.length,
      dataSource: 'live',
    };
  }

  if (section === 'promotion') {
    const policyArtifact = await loadPolicyArtifact(store);
    if (policyArtifact) {
      return {
        policyVersions: policyArtifact.policyVersions,
        artifactGeneratedAt: policyArtifact.generatedAt,
        dataSource: 'artifact',
      };
    }
    const content = await store.readRootFile('policy_registry.json');
    let policyVersions: PolicyVersion[] = [];
    if (content) {
      try { policyVersions = JSON.parse(content).versions || []; } catch { /* graceful */ }
    }
    return { policyVersions, dataSource: 'live' };
  }

  if (section === 'counterfactual') {
    const policyArtifact = await loadPolicyArtifact(store);
    if (policyArtifact?.counterfactual) {
      return {
        counterfactual: policyArtifact.counterfactual,
        runsScanned: policyArtifact.runsScanned,
        totalRuns: policyArtifact.totalRuns,
        artifactGeneratedAt: policyArtifact.generatedAt,
        dataSource: 'artifact',
      };
    }
    const allRows: CounterfactualRow[] = [];
    for (let i = 0; i < counterfactualRunIds.length; i += RUN_BATCH_SIZE) {
      const batch = counterfactualRunIds.slice(i, i + RUN_BATCH_SIZE);
      const batchRows = await Promise.all(
        batch.map(async (runId) => {
          const content = await store.readFile(runId, 'counterfactual_learning.csv');
          return content ? parseCounterfactualCSV(content, runId) : [];
        }),
      );
      for (const rows of batchRows) allRows.push(...rows);
    }
    return {
      counterfactual: summarizeCounterfactual(allRows),
      runsScanned: counterfactualRunIds.length,
      totalRuns: allRunIds.length,
      dataSource: 'live',
    };
  }

  const acc = emptyIngestAcc();

  if (section === 'learning') {
    const learningArtifact = await loadLearningArtifact(store);
    if (learningArtifact) {
      const page = paginateSlice(learningArtifact.learning, offset, limit);
      return {
        learning: page.items,
        decisionLearning: learningArtifact.decisionLearning,
        runsScanned: learningArtifact.runsScanned,
        totalRuns: learningArtifact.totalRuns,
        totalRows: page.totalRows,
        offset: page.offset,
        limit: page.limit,
        hasMore: page.hasMore,
        artifactGeneratedAt: learningArtifact.generatedAt,
        dataSource: 'artifact',
      };
    }
    await ingestBatches(store, learningRunIds, 'learning', acc);
    const page = paginateSlice(acc.learning, offset, limit);
    return {
      learning: page.items,
      decisionLearning: acc.decisionLearning,
      runsScanned: learningRunIds.length,
      totalRuns: allRunIds.length,
      totalRows: page.totalRows,
      offset: page.offset,
      limit: page.limit,
      hasMore: page.hasMore,
      dataSource: 'live',
    };
  }

  if (section === 'decisions') {
    const learningArtifact = await loadLearningArtifact(store);
    if (learningArtifact) {
      const page = paginateSlice(learningArtifact.decisions, offset, limit);
      return {
        decisions: page.items,
        decisionLearning: learningArtifact.decisionLearning,
        runsScanned: learningArtifact.runsScanned,
        totalRuns: learningArtifact.totalRuns,
        totalRows: page.totalRows,
        offset: page.offset,
        limit: page.limit,
        hasMore: page.hasMore,
        artifactGeneratedAt: learningArtifact.generatedAt,
        dataSource: 'artifact',
      };
    }
    await ingestBatches(store, learningRunIds, 'learning', acc);
    const page = paginateSlice(acc.decisions, offset, limit);
    return {
      decisions: page.items,
      decisionLearning: acc.decisionLearning,
      runsScanned: learningRunIds.length,
      totalRuns: allRunIds.length,
      totalRows: page.totalRows,
      offset: page.offset,
      limit: page.limit,
      hasMore: page.hasMore,
      dataSource: 'live',
    };
  }

  if (section === 'assist') {
    const policyArtifact = await loadPolicyArtifact(store);
    if (policyArtifact) {
      return {
        assistRecords: policyArtifact.assistRecords,
        runsScanned: policyArtifact.runsScanned,
        totalRuns: policyArtifact.totalRuns,
        artifactGeneratedAt: policyArtifact.generatedAt,
        dataSource: 'artifact',
      };
    }
    await ingestBatches(store, policyRunIds, 'policy', acc);
    return {
      assistRecords: acc.assistRecords,
      runsScanned: policyRunIds.length,
      totalRuns: allRunIds.length,
      dataSource: 'live',
    };
  }

  if (section === 'policies') {
    const policyArtifact = await loadPolicyArtifact(store);
    if (policyArtifact) {
      return {
        policyDecisions: policyArtifact.policyDecisions,
        policyLearningReports: policyArtifact.policyLearningReports,
        policyEvalCount: policyArtifact.policyEvalCount,
        registryVersionCount: policyArtifact.registryVersionCount,
        runsScanned: policyArtifact.runsScanned,
        totalRuns: policyArtifact.totalRuns,
        artifactGeneratedAt: policyArtifact.generatedAt,
        dataSource: 'artifact',
      };
    }
    await ingestBatches(store, policyRunIds, 'policy', acc);
    const registryContent = await store.readRootFile('policy_registry.json');
    let registryVersionCount = 0;
    if (registryContent) {
      try {
        const reg = JSON.parse(registryContent);
        registryVersionCount = (reg.versions || []).length;
      } catch { /* graceful */ }
    }
    return {
      policyDecisions: acc.policyDecisions,
      policyLearningReports: acc.policyLearningReports,
      policyEvalCount: acc.policyEvalCount,
      registryVersionCount,
      runsScanned: policyRunIds.length,
      totalRuns: allRunIds.length,
      dataSource: 'live',
    };
  }

  const policyArtifact = await loadPolicyArtifact(store);
  if (policyArtifact?.model) {
    return {
      model: policyArtifact.model,
      totalRuns: policyArtifact.totalRuns,
      artifactGeneratedAt: policyArtifact.generatedAt,
      dataSource: 'artifact',
    };
  }
  const modelContent = await store.readRootFile('worker_model.json');
  let model: WorkerModel | null = null;
  if (modelContent) {
    try { model = JSON.parse(modelContent); } catch { /* graceful */ }
  }
  return { model, totalRuns: allRunIds.length, dataSource: 'live' };
}
