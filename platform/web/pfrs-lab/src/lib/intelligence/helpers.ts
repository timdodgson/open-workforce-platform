import type { StorageProvider } from '@/lib/storage';
import type { CounterfactualRow, CounterfactualSummary, PolicyVersion } from '@/lib/types/intelligence';
import type { ContinuousLearningState } from './types';

export function paginateSlice<T>(items: T[], offset: number, limit: number) {
  const slice = items.slice(offset, offset + limit);
  return {
    items: slice,
    totalRows: items.length,
    offset,
    limit,
    hasMore: offset + limit < items.length,
  };
}

export function summarizeCounterfactual(rows: CounterfactualRow[]): CounterfactualSummary {
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

export async function readRootFileFirst(
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
export async function loadContinuousLearningState(
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
