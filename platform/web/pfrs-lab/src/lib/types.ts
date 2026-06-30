// PFRS Research Lab — Core Types

export interface WeekRecord {
  instance: string;
  seed: number;
  mode: string;
  iterationsPerWorker: number;
  maxTotalWorkers: number;
  maxConcurrent: number;
  initialTemperature: number;
  coolingRate: number;
  coolingMode: string;
  effectiveCoolingRate: number;
  minTemperature: number;
  lateAcceptanceLen: number;
  week: number;
  startPenalty: number;
  finalPenalty: number;
  improvement: number;
  hardViolations: number;
  softViolations: number;
  candidates: number;
  accepted: number;
  rejected: number;
  acceptanceRate: number;
  bestIteration: number;
  bestWorkerID: number;
  workersStarted: number;
  branchesCreated: number;
  branchesDropped: number;
  maxQueueDepth: number;
  maxConcurrentSeen: number;
  durationMs: number;
  saFinalTemp: number;
  saTempAtBest: number;
  saAcceptedBetter: number;
  saAcceptedWorse: number;
  saRejectedByProb: number;
  lahcAcceptedByCurrent: number;
  lahcAcceptedByLate: number;
  lahcRejectedByLate: number;
  branchesQueued: number;
  branchesStarted: number;
  branchesCompleted: number;
  winningBranchDepth: number;
  workersImproved: number;
  workersProducedBest: number;
  rejectedNoop: number;
  rejectedSkill: number;
  rejectedSuccession: number;
  rejectedHistory: number;
}

export interface RunMetadata {
  instance: string;
  algorithm: string;
  mode: string;
  iterationsPerWorker: number;
  initialTemperature: number;
  coolingMode: string;
  effectiveCoolingRate: number;
  beamWidth: number;
  beamSeeds: number[];
  seed: number;
  cpus: number;
  maxTotalWorkers: number;
}

export interface PreviousBest {
  instance: string;
  bestPenalty: number;
  label: string;
}

export interface RunSummary {
  metadata: RunMetadata | null;
  previousBest: PreviousBest | null;
  weeks: WeekRecord[];
  totalPenalty: number;
  totalCandidates: number;
  totalAccepted: number;
  totalRejected: number;
  totalSABetter: number;
  totalSAWorse: number;
  totalSARejProb: number;
  totalWorkers: number;
  totalBranches: number;
  totalDurationMs: number;
  hardRejectRate: number;
  acceptWorseRate: number;
  cumulativePenalties: number[];
  numWeeks: number;
  maxWeekPenalty: number;
  maxWeekNum: number;
}

export interface TreeNode {
  pathID: number;
  parentID: number;
  week: number;
  seed: number;
  weekPenalty: number;
  cumulativePenalty: number;
  workersStarted: number;
  candidates: number;
  accepted: number;
  rejected: number;
  saAcceptedBetter: number;
  saAcceptedWorse: number;
  saRejectedByProb: number;
  hardRejectRate: number;
  durationMs: number;
  retained: boolean;
  retainedRank: number;
  winning: boolean;
}

export interface PlateauEvent {
  week: number;
  workerID: number;
  parentWorkerID: number;
  depth: number;
  candidate: number;
  temperature: number;
  currentPenalty: number;
  localBest: number;
  globalBest: number;
  candsSinceImprove: number;
}

export interface BranchEvent {
  week: number;
  workerID: number;
  candidate: number;
  oldPenalty: number;
  newPenalty: number;
  improvement: number;
  timestampMs: number;
}

export interface WorkerLifecycle {
  workerID: number;
  parentWorkerID: number;
  week: number;
  seed: number;
  depth: number;
  startTimeMs: number;
  finishTimeMs: number;
  finishCandidate: number;
  initialTemperature: number;
  finalTemperature: number;
  temperatureAtBest: number;
  bestCandidate: number;
  plateauCount: number;
  branchCount: number;
  producedGlobalBest: boolean;
  finalPenalty: number;
  bestPenalty: number;
  startPenalty: number;
}

export interface ImprovementEvent {
  week: number;
  workerID: number;
  candidate: number;
  temperatureAtEvent: number;
  oldGlobalBest: number;
  newGlobalBest: number;
  improvement: number;
  elapsedMs: number;
}
