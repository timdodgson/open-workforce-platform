import type { StorageProvider } from '@/lib/storage';
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
import type { IntelligenceSummary } from '@/lib/intelligence/types';

export const ARTIFACT_SUMMARY = 'intelligence_summary.json';
export const ARTIFACT_LEARNING = 'intelligence_learning.json';
export const ARTIFACT_POLICY = 'policy_dashboard.json';

export interface IntelligenceSummaryArtifact {
  generatedAt: string;
  summary: IntelligenceSummary;
  totalRuns: number;
}

export interface IntelligenceLearningArtifact {
  generatedAt: string;
  runsScanned: number;
  totalRuns: number;
  learning: LearningRecord[];
  decisionLearning: LearningRecord[];
  decisions: DecisionRecord[];
}

export interface PolicyDashboardArtifact {
  generatedAt: string;
  runsScanned: number;
  totalRuns: number;
  continuousLearning: ContinuousLearningState | null;
  policyVersions: PolicyVersion[];
  policyLearningReports: PolicyLearningReport[];
  policyDecisions: PolicyDecisionRecord[];
  policyEvalCount: number;
  registryVersionCount: number;
  assistRecords: UnifiedAssistRecord[];
  counterfactual: CounterfactualSummary | null;
  model: WorkerModel | null;
}

async function readJson<T>(store: StorageProvider, path: string): Promise<T | null> {
  const content = await store.readRootFile(path);
  if (!content) return null;
  try {
    return JSON.parse(content) as T;
  } catch {
    return null;
  }
}

export async function loadSummaryArtifact(
  store: StorageProvider,
): Promise<IntelligenceSummaryArtifact | null> {
  return readJson(store, ARTIFACT_SUMMARY);
}

export async function loadLearningArtifact(
  store: StorageProvider,
): Promise<IntelligenceLearningArtifact | null> {
  return readJson(store, ARTIFACT_LEARNING);
}

export async function loadPolicyArtifact(
  store: StorageProvider,
): Promise<PolicyDashboardArtifact | null> {
  return readJson(store, ARTIFACT_POLICY);
}

export interface BuildIntelligenceArtifactsInput {
  summary: IntelligenceSummary;
  totalRuns: number;
  learningRunsScanned: number;
  learning: LearningRecord[];
  decisionLearning: LearningRecord[];
  decisions: DecisionRecord[];
  policyRunsScanned: number;
  continuousLearning: ContinuousLearningState | null;
  policyVersions: PolicyVersion[];
  policyLearningReports: PolicyLearningReport[];
  policyDecisions: PolicyDecisionRecord[];
  policyEvalCount: number;
  registryVersionCount: number;
  assistRecords: UnifiedAssistRecord[];
  counterfactual: CounterfactualSummary | null;
  model: WorkerModel | null;
}

/** Write precomputed intelligence artifacts to S3 (run after uploads / retrain). */
export async function writeIntelligenceArtifacts(
  store: StorageProvider,
  data: BuildIntelligenceArtifactsInput,
): Promise<{ paths: string[] }> {
  const generatedAt = new Date().toISOString();

  const summaryArtifact: IntelligenceSummaryArtifact = {
    generatedAt,
    summary: data.summary,
    totalRuns: data.totalRuns,
  };

  const learningArtifact: IntelligenceLearningArtifact = {
    generatedAt,
    runsScanned: data.learningRunsScanned,
    totalRuns: data.totalRuns,
    learning: data.learning,
    decisionLearning: data.decisionLearning,
    decisions: data.decisions,
  };

  const policyArtifact: PolicyDashboardArtifact = {
    generatedAt,
    runsScanned: data.policyRunsScanned,
    totalRuns: data.totalRuns,
    continuousLearning: data.continuousLearning,
    policyVersions: data.policyVersions,
    policyLearningReports: data.policyLearningReports,
    policyDecisions: data.policyDecisions,
    policyEvalCount: data.policyEvalCount,
    registryVersionCount: data.registryVersionCount,
    assistRecords: data.assistRecords,
    counterfactual: data.counterfactual,
    model: data.model,
  };

  await Promise.all([
    store.writeRootFile(ARTIFACT_SUMMARY, JSON.stringify(summaryArtifact)),
    store.writeRootFile(ARTIFACT_LEARNING, JSON.stringify(learningArtifact)),
    store.writeRootFile(ARTIFACT_POLICY, JSON.stringify(policyArtifact)),
  ]);

  return { paths: [ARTIFACT_SUMMARY, ARTIFACT_LEARNING, ARTIFACT_POLICY] };
}
