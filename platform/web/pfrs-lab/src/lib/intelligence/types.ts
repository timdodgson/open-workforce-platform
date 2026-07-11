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
  dataSource?: 'artifact' | 'live';
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

export type IngestAcc = {
  learning: LearningRecord[];
  decisions: DecisionRecord[];
  decisionLearning: LearningRecord[];
  assistRecords: UnifiedAssistRecord[];
  policyDecisions: PolicyDecisionRecord[];
  policyLearningReports: PolicyLearningReport[];
  policyEvalCount: number;
};

export type IngestMode = 'learning' | 'policy';

export interface ArtifactFileStatus {
  exists: boolean;
  generatedAt?: string;
  totalRuns?: number;
  runsScanned?: number;
}

export interface ArtifactStatus {
  summary: ArtifactFileStatus;
  learning: ArtifactFileStatus;
  policy: ArtifactFileStatus;
  currentTotalRuns: number;
  stale: boolean;
  staleReason?: string;
}
