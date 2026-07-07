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
  isAdaptive: boolean;
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
  usedLearned: boolean;
  fallbackReason: string;
}

export type UnifiedAssistRecord = WorkerAssistRecord | SearchAssistRecord | PortfolioAssistRecord;
