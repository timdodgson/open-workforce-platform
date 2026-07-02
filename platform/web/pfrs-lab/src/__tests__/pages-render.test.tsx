/**
 * Render tests for all client page components.
 * Verifies each page renders without crashing given valid mock data.
 */
import React from 'react';
import { render } from '@testing-library/react';
import '@testing-library/jest-dom';
import { RunSummary, TreeNode, DiversityRecord, DiscoveryRecord, WorkerLifecycle, PlateauEvent, WeekRecord, ImprovementEvent } from '@/lib/types';

// --- Mock data factories ---
function mockWeek(week: number): WeekRecord {
  return {
    instance: 'n012w8', seed: 42, mode: 'sa', iterationsPerWorker: 500000,
    maxTotalWorkers: 32, maxConcurrent: 16, initialTemperature: 2.0,
    coolingRate: 0.001, coolingMode: 'adaptive', effectiveCoolingRate: 0.001,
    minTemperature: 0.0001, lateAcceptanceLen: 1000, week,
    startPenalty: 1000, finalPenalty: 400 + week * 50, improvement: 600 - week * 50,
    hardViolations: 0, softViolations: 10, candidates: 500000,
    accepted: 100, rejected: 400000, acceptanceRate: 0.02,
    bestIteration: 250000, bestWorkerID: 5, workersStarted: 10,
    branchesCreated: 8, branchesDropped: 2, maxQueueDepth: 4,
    maxConcurrentSeen: 8, durationMs: 1200, saFinalTemp: 0.001,
    saTempAtBest: 0.5, saAcceptedBetter: 50, saAcceptedWorse: 30,
    saRejectedByProb: 420000, lahcAcceptedByCurrent: 0, lahcAcceptedByLate: 0,
    lahcRejectedByLate: 0, branchesQueued: 8, branchesStarted: 8,
    branchesCompleted: 8, winningBranchDepth: 2, workersImproved: 5,
    workersProducedBest: 2, rejectedNoop: 380000, rejectedSkill: 10000,
    rejectedSuccession: 5000, rejectedHistory: 5000,
  };
}

function mockSummary(): RunSummary {
  const weeks = [1,2,3,4,5,6,7].map(mockWeek);
  return {
    metadata: { instance: 'n012w8', algorithm: 'pfrs', mode: 'sa', iterationsPerWorker: 500000, initialTemperature: 2.0, coolingMode: 'adaptive', effectiveCoolingRate: 0.001, beamWidth: 5, beamSeeds: [42,101,202], seed: 42, cpus: 16, maxTotalWorkers: 32 },
    previousBest: null, weeks, totalPenalty: 3430, totalCandidates: 3500000,
    totalAccepted: 700, totalRejected: 2800000, totalSABetter: 350,
    totalSAWorse: 210, totalSARejProb: 2940000, totalWorkers: 70,
    totalBranches: 56, totalDurationMs: 8400, hardRejectRate: 80,
    acceptWorseRate: 0.006, lahcAcceptByLateRate: 0, cumulativePenalties: [450,900,1350,1800,2250,2700,3430],
    numWeeks: 7, maxWeekPenalty: 750, maxWeekNum: 7,
  };
}

function mockTree(): TreeNode[] {
  return [
    { pathID: 1, parentID: -1, week: 1, seed: 42, weekPenalty: 385, cumulativePenalty: 385, workersStarted: 10, candidates: 500000, accepted: 100, rejected: 400000, saAcceptedBetter: 50, saAcceptedWorse: 30, saRejectedByProb: 420000, hardRejectRate: 80, durationMs: 1200, retained: true, retainedRank: 1, winning: true },
    { pathID: 2, parentID: 1, week: 2, seed: 42, weekPenalty: 245, cumulativePenalty: 630, workersStarted: 8, candidates: 400000, accepted: 80, rejected: 320000, saAcceptedBetter: 40, saAcceptedWorse: 25, saRejectedByProb: 335000, hardRejectRate: 80, durationMs: 900, retained: true, retainedRank: 1, winning: true },
    { pathID: 3, parentID: 1, week: 2, seed: 101, weekPenalty: 300, cumulativePenalty: 685, workersStarted: 8, candidates: 400000, accepted: 70, rejected: 330000, saAcceptedBetter: 35, saAcceptedWorse: 20, saRejectedByProb: 345000, hardRejectRate: 80, durationMs: 850, retained: true, retainedRank: 2, winning: false },
  ];
}

function mockDiscoveries(): DiscoveryRecord[] {
  return [
    { week: 1, workerID: 5, beamPath: 1, candidate: 250000, elapsedMs: 600, temperatureAtEvent: 0.5, currentPenalty: 500, previousBest: 1000, newBest: 500, improvement: 500, improvementPercent: 50, eventType: 'global_best', branchDepth: 0, seedUsed: 42, acceptedWorseCount: 30, hardRejectCount: 5000, softRejectCount: 380000, discoveryNumber: 1, candsSincePrevious: 250000, timeSincePreviousMs: 600, improvementPer10K: 20, improvementPerSecond: 833, postReheatImproved: false, postReheatBestDelta: 0, postReheatCandidatesToImprove: 0, postReheatSpawnedBranch: false, postReheatBeatGlobal: false, postReheatOnWinningLineage: false },
    { week: 1, workerID: 3, beamPath: 1, candidate: 100000, elapsedMs: 300, temperatureAtEvent: 1.0, currentPenalty: 700, previousBest: 1000, newBest: 700, improvement: 300, improvementPercent: 30, eventType: 'local_best', branchDepth: 0, seedUsed: 42, acceptedWorseCount: 15, hardRejectCount: 3000, softRejectCount: 200000, discoveryNumber: 1, candsSincePrevious: 100000, timeSincePreviousMs: 300, improvementPer10K: 30, improvementPerSecond: 1000, postReheatImproved: false, postReheatBestDelta: 0, postReheatCandidatesToImprove: 0, postReheatSpawnedBranch: false, postReheatBeatGlobal: false, postReheatOnWinningLineage: false },
  ];
}

function mockWorkers(): WorkerLifecycle[] {
  return [
    { workerID: 1, parentWorkerID: -1, week: 1, seed: 42, depth: 0, startTimeMs: 0, finishTimeMs: 1200, finishCandidate: 500000, initialTemperature: 2.0, finalTemperature: 0.001, temperatureAtBest: 0.5, bestCandidate: 250000, plateauCount: 2, branchCount: 1, producedGlobalBest: true, finalPenalty: 500, bestPenalty: 500, startPenalty: 1000 },
    { workerID: 2, parentWorkerID: 1, week: 1, seed: 42, depth: 1, startTimeMs: 200, finishTimeMs: 1000, finishCandidate: 400000, initialTemperature: 1.5, finalTemperature: 0.002, temperatureAtBest: 0.8, bestCandidate: 150000, plateauCount: 1, branchCount: 0, producedGlobalBest: false, finalPenalty: 700, bestPenalty: 650, startPenalty: 800 },
  ];
}

function mockPlateaus(): PlateauEvent[] {
  return [
    { week: 1, workerID: 1, parentWorkerID: -1, depth: 0, candidate: 100000, temperature: 1.0, currentPenalty: 600, localBest: 600, globalBest: 500, candsSinceImprove: 3000 },
  ];
}

function mockDiversity(): DiversityRecord[] {
  return [
    { week: 1, pathID: 1, fingerprint: 'abc', hammingToBest: 0, hammingToParent: 0, beamSpread: 5, nearDuplicate: false, retained: true, retainedRank: 1, winning: true, cumulativePenalty: 385, weekPenalty: 385 },
    { week: 1, pathID: 2, fingerprint: 'def', hammingToBest: 3, hammingToParent: 3, beamSpread: 5, nearDuplicate: false, retained: true, retainedRank: 2, winning: false, cumulativePenalty: 400, weekPenalty: 400 },
  ];
}

function mockImprovements(): ImprovementEvent[] {
  return [
    { week: 1, workerID: 5, candidate: 250000, temperatureAtEvent: 0.5, oldGlobalBest: 1000, newGlobalBest: 500, improvement: 500, elapsedMs: 600 },
  ];
}

// --- Import all client components ---
import PenaltyWaterfall from '@/app/runs/[id]/waterfall/PenaltyWaterfall';
import WorkerAnalysis from '@/app/runs/[id]/workers/WorkerAnalysis';
import FamilyEvolution from '@/app/runs/[id]/families/FamilyEvolution';
import PlateauAtlas from '@/app/runs/[id]/plateaus/PlateauAtlas';
import SearchMap from '@/app/runs/[id]/map/SearchMap';
import ReplayPlayer from '@/app/runs/[id]/replay/ReplayPlayer';

// --- Render tests ---
describe('PenaltyWaterfall', () => {
  it('renders without crashing', () => {
    const { container } = render(<PenaltyWaterfall weeks={mockSummary().weeks} totalPenalty={3430} />);
    expect(container).toBeTruthy();
  });
});

describe('WorkerAnalysis', () => {
  it('renders without crashing', () => {
    const { container } = render(<WorkerAnalysis workers={mockWorkers()} discoveries={mockDiscoveries()} />);
    expect(container).toBeTruthy();
  });
});

describe('FamilyEvolution', () => {
  it('renders without crashing', () => {
    const { container } = render(<FamilyEvolution tree={mockTree()} />);
    expect(container).toBeTruthy();
  });
});

describe('PlateauAtlas', () => {
  it('renders without crashing', () => {
    const { container } = render(<PlateauAtlas plateaus={mockPlateaus()} numWeeks={7} />);
    expect(container).toBeTruthy();
  });
});

describe('SearchMap', () => {
  it('renders without crashing', () => {
    const { container } = render(<SearchMap discoveries={mockDiscoveries()} />);
    expect(container).toBeTruthy();
  });
});

describe('ReplayPlayer', () => {
  it('renders without crashing', () => {
    const { container } = render(
      <ReplayPlayer discoveries={mockDiscoveries()} improvements={mockImprovements()} tree={mockTree()} workers={mockWorkers()} />
    );
    expect(container).toBeTruthy();
  });
});
