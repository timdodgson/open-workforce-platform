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
