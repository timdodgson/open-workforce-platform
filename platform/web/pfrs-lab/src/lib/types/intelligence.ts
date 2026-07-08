/** Shared intelligence / SI types (lib layer — no app imports). */

export interface LearningRecord {
  runId: string;
  problemType: string;
  instance: string;
  algorithm: string;
  seed: number;
  week: number;
  depth: number;
  temperature: number;
  iterationsAlloc: number;
  globalBest: number;
  parentObjective: number;
  distanceFromBest: number;
  improved: boolean;
  producedGlobalBest: boolean;
  improvementAmount: number;
  finalObjective: number;
  runtimeMs: number;
  candidatesEval: number;
  accepted: number;
  rejected: number;
  roi: number;
  improvPerCPU: number;
  improvPer100K: number;
}

export interface DecisionRecord {
  runId: string;
  workerId: number;
  week: number;
  depth: number;
  algorithm: string;
  parentObjective: number;
  globalBest: number;
  distanceFromBest: number;
  beamRank: number;
  entropy: number;
  beamHealth: number;
  recentImprovRate: number;
  allocatedIters: number;
  recommendation: string;
  confidence: number;
  reasonCodes: string;
  suggestedAlgorithm: string;
  suggestedBudget: number;
  improved: boolean;
  producedGlobalBest: boolean;
  improvementAmount: number;
  finalObjective: number;
  runtimeMs: number;
  roi: number;
}

export interface ModelResult {
  target: string;
  type: 'classifier' | 'regressor';
  max_depth: number;
  n_train: number;
  n_test: number;
  metrics: Record<string, number>;
  feature_importance: Record<string, number>;
  confusion_matrix?: { labels: string[]; matrix: number[][]; tp?: number; fp?: number; fn?: number; tn?: number };
}

export interface WorkerModel {
  version: string;
  description: string;
  training_samples: number;
  test_samples: number;
  features: string[];
  models: Record<string, ModelResult>;
  data_summary: {
    total_records: number;
    improvement_rate: number;
    global_best_rate: number;
    mean_improvement: number;
    mean_roi: number;
  };
}

export type AssistArchitecture = 'worker' | 'search' | 'portfolio';

export interface WorkerAssistRecord {
  architecture: 'worker';
  domain: string;
  runId: string;
  workerId: number;
  algorithm: string;
  recommendation: string;
  confidence: number;
  reasonCodes: string;
  safetyTriggered: boolean;
  safetyRule: string;
  outcome: string;
  finalAction: string;
  improved: boolean;
  producedGlobalBest: boolean;
  improvementAmount: number;
  runtimeMs: number;
}

export interface SearchAssistRecord {
  architecture: 'search';
  domain: string;
  runId: string;
  algorithm: string;
  checkpoint: number;
  candidates: number;
  iterationsTotal: number;
  bestPenalty: number;
  plateauLength: number;
  recommendedAction: string;
  confidence: number;
  reasons: string;
  safetyTriggered: boolean;
  safetyRule: string;
  accepted: boolean;
  finalAction: string;
  finalBestPenalty: number;
  runtimeMs: number;
  isAdaptive?: boolean;
}

export interface PortfolioAssistRecord {
  architecture: 'portfolio';
  domain: string;
  runId: string;
  instance: string;
  strategy: string;
  originalBudget: number;
  recommendedBudget: number;
  finalBudget: number;
  recommendation: string;
  confidence: number;
  reasonCodes: string;
  accepted: boolean;
  safetyRejected: boolean;
  safetyRule: string;
  resultObjective: number;
  strategyWon: boolean;
  runtimeMs: number;
  usedLearned?: boolean;
  fallbackReason?: string;
}

export type UnifiedAssistRecord = WorkerAssistRecord | SearchAssistRecord | PortfolioAssistRecord;

export interface PolicyDecisionRecord {
  runId: string;
  checkpoint: number;
  candidates: number;
  policyMode: string;
  policyUsed: string;
  action: string;
  confidence: number;
  fallbackReason: string;
  safetyOverride: boolean;
}

export interface PolicyLearningReport {
  runId: string;
  action: string;
  reason: string;
  samplesAdded: number;
}

export interface ContinuousLearningState {
  new_samples_since_training: number;
  total_samples: number;
  last_trained_at?: string;
  last_training_accuracy: number;
  production_version: string;
  candidate_version?: string;
  candidate_accuracy?: number;
  recommendation: string;
  recommend_reason: string;
}

export interface PolicyVersion {
  id: string;
  version: string;
  domain: string;
  decision_type: string;
  status: string;
  offline_accuracy: number;
  shadow_accuracy: number;
  production_accuracy: number;
  production_runs: number;
  regret_vs_rules: number;
  drift_detected: boolean;
  promoted_at?: string;
  retired_at?: string;
  rolled_back_from?: string;
  rollback_reason?: string;
}

export interface CounterfactualRow {
  timestamp: string;
  runId: string;
  decisionType: string;
  domain: string;
  instance: string;
  algorithm: string;
  actualAction: string;
  confidence: number;
  policyId: string;
  policyVersion: string;
  policyType: string;
  expectedValue: number;
  actualOutcome: number;
  outcomeMetric: string;
  regret: number;
  bestAlternative: string;
}

export interface CounterfactualSummary {
  totalDecisions: number;
  meanRegret: number;
  regretRate: number;
  trainingOpportunities: number;
  byDomain: { domain: string; decisions: number; totalRegret: number; meanRegret: number; regretRate: number }[];
  topRegret: CounterfactualRow[];
}
