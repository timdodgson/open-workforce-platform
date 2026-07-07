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

export interface LearningRecord {
  runId: string;
  problemType: string;
  instance: string;
  algorithm: string;
  seed: number;
  week: number;
  depth: number;
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
  roi: number;
}
