/**
 * Render tests part 2 — DNA, Efficiency, Archetypes, Genealogy, Causality, Explain.
 */
import React from 'react';
import { render } from '@testing-library/react';
import '@testing-library/jest-dom';
import { RunSummary, TreeNode, DiversityRecord, DiscoveryRecord, WorkerLifecycle, PlateauEvent, ImprovementEvent } from '@/lib/types';

// Reuse mock factories.
function mockWeek(week: number) {
  return { instance: 'n012w8', seed: 42, mode: 'sa', iterationsPerWorker: 500000, maxTotalWorkers: 32, maxConcurrent: 16, initialTemperature: 2.0, coolingRate: 0.001, coolingMode: 'adaptive', effectiveCoolingRate: 0.001, minTemperature: 0.0001, lateAcceptanceLen: 1000, week, startPenalty: 1000, finalPenalty: 450, improvement: 550, hardViolations: 0, softViolations: 10, candidates: 500000, accepted: 100, rejected: 400000, acceptanceRate: 0.02, bestIteration: 250000, bestWorkerID: 5, workersStarted: 10, branchesCreated: 8, branchesDropped: 2, maxQueueDepth: 4, maxConcurrentSeen: 8, durationMs: 1200, saFinalTemp: 0.001, saTempAtBest: 0.5, saAcceptedBetter: 50, saAcceptedWorse: 30, saRejectedByProb: 420000, lahcAcceptedByCurrent: 0, lahcAcceptedByLate: 0, lahcRejectedByLate: 0, branchesQueued: 8, branchesStarted: 8, branchesCompleted: 8, winningBranchDepth: 2, workersImproved: 5, workersProducedBest: 2, rejectedNoop: 380000, rejectedSkill: 10000, rejectedSuccession: 5000, rejectedHistory: 5000 };
}
function mockSummary(): RunSummary {
  const weeks = [1,2,3].map(mockWeek);
  return { metadata: { instance: 'n012w8', algorithm: 'pfrs', mode: 'sa', iterationsPerWorker: 500000, initialTemperature: 2.0, coolingMode: 'adaptive', effectiveCoolingRate: 0.001, beamWidth: 5, beamSeeds: [42], seed: 42, cpus: 16, maxTotalWorkers: 32 }, previousBest: null, weeks, totalPenalty: 1350, totalCandidates: 1500000, totalAccepted: 300, totalRejected: 1200000, totalSABetter: 150, totalSAWorse: 90, totalSARejProb: 1260000, totalWorkers: 30, totalBranches: 24, totalDurationMs: 3600, hardRejectRate: 80, acceptWorseRate: 0.006, lahcAcceptByLateRate: 0, cumulativePenalties: [450,900,1350], numWeeks: 3, maxWeekPenalty: 450, maxWeekNum: 1 };
}
function mockTree(): TreeNode[] {
  return [
    { pathID: 1, parentID: -1, week: 1, seed: 42, weekPenalty: 450, cumulativePenalty: 450, workersStarted: 10, candidates: 500000, accepted: 100, rejected: 400000, saAcceptedBetter: 50, saAcceptedWorse: 30, saRejectedByProb: 420000, hardRejectRate: 80, durationMs: 1200, retained: true, retainedRank: 1, winning: true },
    { pathID: 2, parentID: 1, week: 2, seed: 42, weekPenalty: 450, cumulativePenalty: 900, workersStarted: 10, candidates: 500000, accepted: 100, rejected: 400000, saAcceptedBetter: 50, saAcceptedWorse: 30, saRejectedByProb: 420000, hardRejectRate: 80, durationMs: 1200, retained: true, retainedRank: 1, winning: true },
  ];
}
function mockDiscoveries(): DiscoveryRecord[] {
  return [{ week: 1, workerID: 5, beamPath: 1, candidate: 250000, elapsedMs: 600, temperatureAtEvent: 0.5, currentPenalty: 500, previousBest: 1000, newBest: 500, improvement: 500, improvementPercent: 50, eventType: 'global_best', branchDepth: 0, seedUsed: 42, acceptedWorseCount: 30, hardRejectCount: 5000, softRejectCount: 380000, discoveryNumber: 1, candsSincePrevious: 250000, timeSincePreviousMs: 600, improvementPer10K: 20, improvementPerSecond: 833, postReheatImproved: false, postReheatBestDelta: 0, postReheatCandidatesToImprove: 0, postReheatSpawnedBranch: false, postReheatBeatGlobal: false, postReheatOnWinningLineage: false }];
}
function mockWorkers(): WorkerLifecycle[] {
  return [{ workerID: 1, parentWorkerID: -1, week: 1, seed: 42, depth: 0, startTimeMs: 0, finishTimeMs: 1200, finishCandidate: 500000, initialTemperature: 2.0, finalTemperature: 0.001, temperatureAtBest: 0.5, bestCandidate: 250000, plateauCount: 2, branchCount: 1, producedGlobalBest: true, finalPenalty: 500, bestPenalty: 500, startPenalty: 1000 }];
}
function mockDiversity(): DiversityRecord[] {
  return [{ week: 1, pathID: 1, fingerprint: 'abc', hammingToBest: 0, hammingToParent: 0, beamSpread: 5, nearDuplicate: false, retained: true, retainedRank: 1, winning: true, cumulativePenalty: 450, weekPenalty: 450 }];
}
function mockPlateaus(): PlateauEvent[] {
  return [{ week: 1, workerID: 1, parentWorkerID: -1, depth: 0, candidate: 100000, temperature: 1.0, currentPenalty: 600, localBest: 600, globalBest: 500, candsSinceImprove: 3000 }];
}
function mockImprovements(): ImprovementEvent[] {
  return [{ week: 1, workerID: 5, candidate: 250000, temperatureAtEvent: 0.5, oldGlobalBest: 1000, newGlobalBest: 500, improvement: 500, elapsedMs: 600 }];
}

import SearchDNA from '@/app/runs/[id]/dna/SearchDNA';
import EfficiencyDashboard from '@/app/runs/[id]/efficiency/EfficiencyDashboard';
import WorkerArchetypes from '@/app/runs/[id]/archetypes/WorkerArchetypes';
import BranchGenealogy from '@/app/runs/[id]/genealogy/BranchGenealogy';

describe('SearchDNA', () => {
  it('renders without crashing', () => {
    const { container } = render(<SearchDNA summary={mockSummary()} discoveries={mockDiscoveries()} workers={mockWorkers()} tree={mockTree()} diversity={mockDiversity()} />);
    expect(container).toBeTruthy();
  });
});

describe('EfficiencyDashboard', () => {
  it('renders without crashing', () => {
    const { container } = render(<EfficiencyDashboard workers={mockWorkers()} discoveries={mockDiscoveries()} summary={mockSummary()} />);
    expect(container).toBeTruthy();
  });
});

describe('WorkerArchetypes', () => {
  it('renders without crashing', () => {
    const { container } = render(<WorkerArchetypes workers={mockWorkers()} discoveries={mockDiscoveries()} />);
    expect(container).toBeTruthy();
  });
});

describe('BranchGenealogy', () => {
  it('renders without crashing', () => {
    const { container } = render(<BranchGenealogy tree={mockTree()} workers={mockWorkers()} />);
    expect(container).toBeTruthy();
  });
});
