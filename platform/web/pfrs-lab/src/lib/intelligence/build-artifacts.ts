import { invalidateAll } from '@/lib/cache';
import { writeIntelligenceArtifacts } from '@/lib/intelligence-artifacts';
import { getStorageProvider, type StorageProvider } from '@/lib/storage';
import { parseCounterfactualCSV } from '@/lib/parsers/intelligence';
import type { CounterfactualRow, PolicyVersion, WorkerModel } from '@/lib/types/intelligence';
import { RUN_BATCH_SIZE } from './constants';
import { emptyIngestAcc, ingestBatches } from './ingest';
import { loadContinuousLearningState, summarizeCounterfactual } from './helpers';
import { resolveRunIdSets } from './resolve-runs';
import type { IntelligenceSummary } from './types';

/** Scan storage and write precomputed intelligence artifacts to S3 root. */
export async function buildIntelligenceArtifacts(storage?: StorageProvider) {
  const store = storage ?? getStorageProvider();
  const { allRunIds, si2RunIds, learningRunIds, policyRunIds, counterfactualRunIds } =
    await resolveRunIdSets(store);

  const learningAcc = emptyIngestAcc();
  const policyAcc = emptyIngestAcc();

  await Promise.all([
    ingestBatches(store, learningRunIds, 'learning', learningAcc),
    ingestBatches(store, policyRunIds, 'policy', policyAcc),
  ]);

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
