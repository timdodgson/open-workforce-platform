import { cached, invalidateAll } from '@/lib/cache';
import {
  loadLearningArtifact,
  loadPolicyArtifact,
  loadSummaryArtifact,
  writeIntelligenceArtifacts,
} from '@/lib/intelligence-artifacts';
import { getStorageProvider, type StorageProvider } from '@/lib/storage';
import {
  detectDomain,
  parseCounterfactualCSV,
  parseDecisionCSV,
  parseLearningCSV,
  parsePolicyDecisionsCSV,
  parsePolicyLearningReport,
  parsePortfolioAssistCSV,
  parseSearchAssistCSV,
  parseWorkerAssistAsDecisions,
  parseWorkerAssistCSV,
} from '@/lib/parsers/intelligence';
import type {
  ContinuousLearningState,
  CounterfactualSummary,
  DecisionRecord,
  LearningRecord,
  PolicyDecisionRecord,
  PolicyLearningReport,
  PolicyVersion,
  UnifiedAssistRecord,
  WorkerModel,
} from '@/lib/types/intelligence';

export type {
  ContinuousLearningState,
  CounterfactualRow,
  CounterfactualSummary,
  DecisionRecord,
  LearningRecord,
  PolicyDecisionRecord,
  PolicyLearningReport,
  PolicyVersion,
  UnifiedAssistRecord,
  WorkerAssistRecord,
  SearchAssistRecord,
  PortfolioAssistRecord,
  WorkerModel,
} from '@/lib/types/intelligence';

/** Keep S3 reads low — Lambda/ALB timeout risk on large scans. */
const MAX_LEARNING_RUNS = 80;
const MAX_POLICY_RUNS = 80;
const MAX_COUNTERFACTUAL_RUNS = 80;
const RUN_BATCH_SIZE = 8;
export const DEFAULT_PAGE_LIMIT = 100;

export interface IntelligenceLoadOptions {
  offset?: number;
  limit?: number;
}

export interface IntelligencePageMeta {
  totalRows?: number;
  offset?: number;
  limit?: number;
  hasMore?: boolean;
  artifactGeneratedAt?: string;
}

function paginateSlice<T>(items: T[], offset: number, limit: number) {
  const slice = items.slice(offset, offset + limit);
  return {
    items: slice,
    totalRows: items.length,
    offset,
    limit,
    hasMore: offset + limit < items.length,
  };
}

async function resolveRunIdSets(store: StorageProvider) {
  const allRunIds = await store.listRuns();
  const si2RunIds = allRunIds.filter((id) => id.startsWith('si2-') || id.startsWith('val-'));
  const newest = [...allRunIds].sort().reverse();
  const learningRunIds = newest.slice(0, MAX_LEARNING_RUNS);
  const policyRunIds = [...si2RunIds]
    .sort((a, b) => {
      const rank = (id: string) => (id.startsWith('val-deep-') ? 0 : id.startsWith('val-') ? 1 : 2);
      return rank(a) - rank(b) || b.localeCompare(a);
    })
    .slice(0, MAX_POLICY_RUNS);
  const counterfactualRunIds = newest.slice(0, MAX_COUNTERFACTUAL_RUNS);
  return { allRunIds, si2RunIds, learningRunIds, policyRunIds, counterfactualRunIds };
}

export interface IntelligenceData {
  learning: LearningRecord[];
  decisions: DecisionRecord[];
  decisionLearning: LearningRecord[];
  model: WorkerModel | null;
  predictionsData: null;
  assistRecords: UnifiedAssistRecord[];
  policyDecisions: PolicyDecisionRecord[];
  policyLearningReports: PolicyLearningReport[];
  policyEvalCount: number;
  registryVersionCount: number;
  si2RunIds: string[];
  runsScanned: number;
  totalRuns: number;
  continuousLearning?: ContinuousLearningState | null;
  policyVersions?: PolicyVersion[];
  counterfactual?: CounterfactualSummary | null;
}

type IngestAcc = {
  learning: LearningRecord[];
  decisions: DecisionRecord[];
  decisionLearning: LearningRecord[];
  assistRecords: UnifiedAssistRecord[];
  policyDecisions: PolicyDecisionRecord[];
  policyLearningReports: PolicyLearningReport[];
  policyEvalCount: number;
};

type IngestMode = 'learning' | 'policy';

async function readRunFiles(
  storage: StorageProvider,
  runId: string,
  filenames: string[],
): Promise<Record<string, string | null>> {
  const entries = await Promise.all(
    filenames.map(async (name) => [name, await storage.readFile(runId, name)] as const),
  );
  return Object.fromEntries(entries);
}

async function ingestRun(
  storage: StorageProvider,
  runId: string,
  mode: IngestMode,
  acc: IngestAcc,
) {
  const domain = detectDomain(runId);

  if (mode === 'learning') {
    const files = await readRunFiles(storage, runId, [
      'worker_learning.csv',
      'worker_decisions.csv',
      'worker_assist.csv',
    ]);
    const learningContent = files['worker_learning.csv'];
    const decisionContent = files['worker_decisions.csv'];
    const workerAssistContent = files['worker_assist.csv'];

    if (learningContent) {
      const learningRows = parseLearningCSV(learningContent, runId);
      acc.learning.push(...learningRows);
      acc.decisionLearning.push(...learningRows);
    }
    if (decisionContent) acc.decisions.push(...parseDecisionCSV(decisionContent, runId));
    else if (workerAssistContent) acc.decisions.push(...parseWorkerAssistAsDecisions(workerAssistContent, runId));
    if (workerAssistContent) acc.assistRecords.push(...parseWorkerAssistCSV(workerAssistContent, runId));
    return;
  }

  const files = await readRunFiles(storage, runId, [
    'policy_decisions.csv',
    'policy_learning_report.json',
    'policy_evaluation.csv',
    'generic_search_assist.csv',
    'portfolio_assist.csv',
    'worker_assist.csv',
  ]);

  if (files['policy_decisions.csv']) {
    acc.policyDecisions.push(...parsePolicyDecisionsCSV(files['policy_decisions.csv']!, runId));
  }
  if (files['policy_learning_report.json']) {
    const report = parsePolicyLearningReport(files['policy_learning_report.json']!, runId);
    if (report) acc.policyLearningReports.push(report);
  }
  if (files['policy_evaluation.csv']) {
    const evalLines = files['policy_evaluation.csv']!.trim().split('\n').length - 1;
    if (evalLines > 0) acc.policyEvalCount += evalLines;
  }
  if (files['generic_search_assist.csv']) {
    acc.assistRecords.push(...parseSearchAssistCSV(files['generic_search_assist.csv']!, runId, domain));
  }
  if (files['portfolio_assist.csv']) {
    acc.assistRecords.push(...parsePortfolioAssistCSV(files['portfolio_assist.csv']!, runId));
  }
  if (files['worker_assist.csv']) {
    acc.assistRecords.push(...parseWorkerAssistCSV(files['worker_assist.csv']!, runId));
  }
}

async function ingestBatches(
  store: StorageProvider,
  runIds: string[],
  mode: IngestMode,
  acc: IngestAcc,
) {
  for (let i = 0; i < runIds.length; i += RUN_BATCH_SIZE) {
    const batch = runIds.slice(i, i + RUN_BATCH_SIZE);
    await Promise.all(batch.map((runId) => ingestRun(store, runId, mode, acc)));
  }
}

function summarizeCounterfactual(rows: import('@/lib/types/intelligence').CounterfactualRow[]): CounterfactualSummary {
  const totalDecisions = rows.length;
  if (totalDecisions === 0) {
    return {
      totalDecisions: 0,
      meanRegret: 0,
      regretRate: 0,
      trainingOpportunities: 0,
      byDomain: [],
      topRegret: [],
    };
  }
  const totalRegret = rows.reduce((sum, r) => sum + r.regret, 0);
  const positiveRegret = rows.filter((r) => r.regret > 0);
  const byDomainMap = new Map<string, typeof rows>();
  for (const row of rows) {
    const list = byDomainMap.get(row.domain) || [];
    list.push(row);
    byDomainMap.set(row.domain, list);
  }
  return {
    totalDecisions,
    meanRegret: totalRegret / totalDecisions,
    regretRate: (positiveRegret.length / totalDecisions) * 100,
    trainingOpportunities: positiveRegret.length,
    byDomain: [...byDomainMap.entries()].map(([domain, domainRows]) => {
      const domainRegret = domainRows.reduce((s, r) => s + r.regret, 0);
      return {
        domain,
        decisions: domainRows.length,
        totalRegret: domainRegret,
        meanRegret: domainRegret / domainRows.length,
        regretRate: (domainRows.filter((r) => r.regret > 0).length / domainRows.length) * 100,
      };
    }),
    topRegret: [...rows].sort((a, b) => b.regret - a.regret).slice(0, 10),
  };
}

async function readRootFileFirst(
  store: StorageProvider,
  paths: string[],
): Promise<string | null> {
  for (const path of paths) {
    const content = await store.readRootFile(path);
    if (content) return content;
  }
  return null;
}

/** Load learning state — tries S3 paths, then synthesizes from policy_registry.json. */
async function loadContinuousLearningState(
  store: StorageProvider,
): Promise<ContinuousLearningState | null> {
  const content = await readRootFileFirst(store, [
    'policies/learning_state.json',
    'learning_state.json',
  ]);
  if (content) {
    try { return JSON.parse(content) as ContinuousLearningState; } catch { /* fall through */ }
  }

  const registryContent = await store.readRootFile('policy_registry.json');
  if (!registryContent) return null;

  try {
    const versions = (JSON.parse(registryContent).versions || []) as PolicyVersion[];
    if (versions.length === 0) return null;

    const active = versions.filter((v) => v.status === 'active');
    const candidate = versions.find((v) => v.status === 'shadow' || v.status === 'training');
    const totalSamples = versions.reduce((sum, v) => sum + (v.training_samples || 0), 0);
    const prod = active[0];

    return {
      new_samples_since_training: 0,
      total_samples: totalSamples,
      last_trained_at: prod?.training_date,
      last_training_accuracy: prod?.offline_accuracy ?? 0,
      production_version: prod?.version ?? 'rules',
      candidate_version: candidate?.version,
      candidate_accuracy: candidate?.offline_accuracy,
      recommendation: candidate ? 'wait' : active.length > 0 ? 'none' : 'retrain',
      recommend_reason: active.length > 0
        ? `${active.length} active policies in registry (${totalSamples.toLocaleString()} training samples)`
        : 'No active policies — run train_policies.py to populate registry',
    };
  } catch {
    return null;
  }
}

export type IntelligenceSection =
  | 'summary'
  | 'learning'
  | 'decisions'
  | 'model'
  | 'assist'
  | 'policies'
  | 'continuous-learning'
  | 'promotion'
  | 'counterfactual';

export interface IntelligenceSummary {
  totalRuns: number;
  runsScanned: number;
  si2RunIds: string[];
  registryVersionCount: number;
  policyEvalCount: number;
  hasModel: boolean;
}

const SECTION_CACHE_TTL: Partial<Record<IntelligenceSection, number>> = {
  summary: 120_000,
  promotion: 120_000,
  model: 120_000,
  'continuous-learning': 90_000,
  counterfactual: 90_000,
  policies: 60_000,
  assist: 60_000,
  learning: 45_000,
  decisions: 45_000,
};

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
      };
    }
    const acc: IngestAcc = {
      learning: [],
      decisions: [],
      decisionLearning: [],
      assistRecords: [],
      policyDecisions: [],
      policyLearningReports: [],
      policyEvalCount: 0,
    };
    const [state] = await Promise.all([
      loadContinuousLearningState(store),
      ingestBatches(store, policyRunIds, 'policy', acc),
    ]);
    return {
      continuousLearning: state,
      policyLearningReports: acc.policyLearningReports,
      runsScanned: policyRunIds.length,
      totalRuns: allRunIds.length,
    };
  }

  if (section === 'promotion') {
    const policyArtifact = await loadPolicyArtifact(store);
    if (policyArtifact) {
      return {
        policyVersions: policyArtifact.policyVersions,
        artifactGeneratedAt: policyArtifact.generatedAt,
      };
    }
    const content = await store.readRootFile('policy_registry.json');
    let policyVersions: PolicyVersion[] = [];
    if (content) {
      try { policyVersions = JSON.parse(content).versions || []; } catch { /* graceful */ }
    }
    return { policyVersions };
  }

  if (section === 'counterfactual') {
    const policyArtifact = await loadPolicyArtifact(store);
    if (policyArtifact?.counterfactual) {
      return {
        counterfactual: policyArtifact.counterfactual,
        runsScanned: policyArtifact.runsScanned,
        totalRuns: policyArtifact.totalRuns,
        artifactGeneratedAt: policyArtifact.generatedAt,
      };
    }
    const allRows: import('@/lib/types/intelligence').CounterfactualRow[] = [];
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
    };
  }

  const acc: IngestAcc = {
    learning: [],
    decisions: [],
    decisionLearning: [],
    assistRecords: [],
    policyDecisions: [],
    policyLearningReports: [],
    policyEvalCount: 0,
  };

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
      };
    }
    await ingestBatches(store, policyRunIds, 'policy', acc);
    return {
      assistRecords: acc.assistRecords,
      runsScanned: policyRunIds.length,
      totalRuns: allRunIds.length,
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
    };
  }

  const policyArtifact = await loadPolicyArtifact(store);
  if (policyArtifact?.model) {
    return {
      model: policyArtifact.model,
      totalRuns: policyArtifact.totalRuns,
      artifactGeneratedAt: policyArtifact.generatedAt,
    };
  }
  const modelContent = await store.readRootFile('worker_model.json');
  let model: WorkerModel | null = null;
  if (modelContent) {
    try { model = JSON.parse(modelContent); } catch { /* graceful */ }
  }
  return { model, totalRuns: allRunIds.length };
}

/** Scan storage and write precomputed intelligence artifacts to S3 root. */
export async function buildIntelligenceArtifacts(storage?: StorageProvider) {
  const store = storage ?? getStorageProvider();
  const { allRunIds, si2RunIds, learningRunIds, policyRunIds, counterfactualRunIds } =
    await resolveRunIdSets(store);

  const learningAcc: IngestAcc = {
    learning: [],
    decisions: [],
    decisionLearning: [],
    assistRecords: [],
    policyDecisions: [],
    policyLearningReports: [],
    policyEvalCount: 0,
  };
  const policyAcc: IngestAcc = {
    learning: [],
    decisions: [],
    decisionLearning: [],
    assistRecords: [],
    policyDecisions: [],
    policyLearningReports: [],
    policyEvalCount: 0,
  };

  await Promise.all([
    ingestBatches(store, learningRunIds, 'learning', learningAcc),
    ingestBatches(store, policyRunIds, 'policy', policyAcc),
  ]);

  const allRows: import('@/lib/types/intelligence').CounterfactualRow[] = [];
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

  const [continuousLearning, modelContent, registryContent] = await Promise.all([
    loadContinuousLearningState(store),
    store.readRootFile('worker_model.json'),
    store.readRootFile('policy_registry.json'),
  ]);

  let model: WorkerModel | null = null;
  if (modelContent) {
    try { model = JSON.parse(modelContent); } catch { /* graceful */ }
  }

  let policyVersions: PolicyVersion[] = [];
  let registryVersionCount = 0;
  if (registryContent) {
    try {
      const reg = JSON.parse(registryContent);
      policyVersions = reg.versions || [];
      registryVersionCount = policyVersions.length;
    } catch { /* graceful */ }
  }

  const summary: IntelligenceSummary = {
    totalRuns: allRunIds.length,
    runsScanned: new Set([...learningRunIds, ...policyRunIds]).size,
    si2RunIds,
    registryVersionCount,
    policyEvalCount: policyAcc.policyEvalCount,
    hasModel: Boolean(model),
  };

  const { paths } = await writeIntelligenceArtifacts(store, {
    summary,
    totalRuns: allRunIds.length,
    learningRunsScanned: learningRunIds.length,
    learning: learningAcc.learning,
    decisionLearning: learningAcc.decisionLearning,
    decisions: learningAcc.decisions,
    policyRunsScanned: policyRunIds.length,
    continuousLearning,
    policyVersions,
    policyLearningReports: policyAcc.policyLearningReports,
    policyDecisions: policyAcc.policyDecisions,
    policyEvalCount: policyAcc.policyEvalCount,
    registryVersionCount,
    assistRecords: policyAcc.assistRecords,
    counterfactual: summarizeCounterfactual(allRows),
    model,
  });

  invalidateAll();
  return { paths, generatedAt: new Date().toISOString() };
}
