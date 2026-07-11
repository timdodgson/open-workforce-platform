import {
  loadLearningArtifact,
  loadPolicyArtifact,
  loadSummaryArtifact,
} from '@/lib/intelligence-artifacts';
import { getStorageProvider, type StorageProvider } from '@/lib/storage';
import type { ArtifactFileStatus, ArtifactStatus } from './types';

export { formatArtifactAge } from './format-utils';

function fileStatus(
  exists: boolean,
  generatedAt?: string,
  totalRuns?: number,
  runsScanned?: number,
): ArtifactFileStatus {
  return { exists, generatedAt, totalRuns, runsScanned };
}

function staleReason(
  currentTotalRuns: number,
  summary: ArtifactFileStatus,
  learning: ArtifactFileStatus,
  policy: ArtifactFileStatus,
): string | undefined {
  if (!summary.exists || !learning.exists || !policy.exists) {
    const missing = [
      !summary.exists && 'summary',
      !learning.exists && 'learning',
      !policy.exists && 'policy',
    ].filter(Boolean);
    return `Missing artifacts: ${missing.join(', ')}`;
  }

  const runCounts = [summary.totalRuns, learning.totalRuns, policy.totalRuns].filter(
    (n) => n !== undefined,
  );
  if (runCounts.some((n) => n !== currentTotalRuns)) {
    return `Run count changed (${runCounts.join('/')} vs ${currentTotalRuns} current)`;
  }

  const timestamps = [summary.generatedAt, learning.generatedAt, policy.generatedAt].filter(Boolean);
  const unique = new Set(timestamps);
  if (unique.size > 1) {
    return 'Artifacts rebuilt at different times — run rebuild to sync';
  }

  return undefined;
}

/** Check whether precomputed intelligence artifacts are present and up to date. */
export async function getArtifactStatus(storage?: StorageProvider): Promise<ArtifactStatus> {
  const store = storage ?? getStorageProvider();
  const [summaryArtifact, learningArtifact, policyArtifact, currentTotalRuns] = await Promise.all([
    loadSummaryArtifact(store),
    loadLearningArtifact(store),
    loadPolicyArtifact(store),
    store.listRuns().then((ids) => ids.length),
  ]);

  const summary = fileStatus(
    Boolean(summaryArtifact),
    summaryArtifact?.generatedAt,
    summaryArtifact?.totalRuns,
  );
  const learning = fileStatus(
    Boolean(learningArtifact),
    learningArtifact?.generatedAt,
    learningArtifact?.totalRuns,
    learningArtifact?.runsScanned,
  );
  const policy = fileStatus(
    Boolean(policyArtifact),
    policyArtifact?.generatedAt,
    policyArtifact?.totalRuns,
    policyArtifact?.runsScanned,
  );

  const reason = staleReason(currentTotalRuns, summary, learning, policy);

  return {
    summary,
    learning,
    policy,
    currentTotalRuns,
    stale: Boolean(reason),
    staleReason: reason,
  };
}
