import type { StorageProvider } from '@/lib/storage';
import {
  MAX_COUNTERFACTUAL_RUNS,
  MAX_LEARNING_RUNS,
  MAX_POLICY_RUNS,
} from './constants';

export async function resolveRunIdSets(store: StorageProvider) {
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
