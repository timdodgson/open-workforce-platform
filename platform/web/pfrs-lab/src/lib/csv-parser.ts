import { WeekRecord, TreeNode, PlateauEvent, BranchEvent, WorkerLifecycle } from './types';

export function parseAuditCSV(content: string): WeekRecord[] {
  const lines = content.trim().split('\n');
  if (lines.length < 2) return [];

  const records: WeekRecord[] = [];
  for (let i = 1; i < lines.length; i++) {
    const line = lines[i].trim();
    if (!line) continue;
    const f = line.split(',');
    if (f.length < 20) continue;

    let idx = 0;
    const r: WeekRecord = {
      instance: f[idx++],
      seed: parseInt(f[idx++]) || 0,
      mode: f[idx++],
      iterationsPerWorker: parseInt(f[idx++]) || 0,
      maxTotalWorkers: parseInt(f[idx++]) || 0,
      maxConcurrent: parseInt(f[idx++]) || 0,
      initialTemperature: parseFloat(f[idx++]) || 0,
      coolingRate: parseFloat(f[idx++]) || 0,
      coolingMode: f[idx++] || '',
      effectiveCoolingRate: parseFloat(f[idx++]) || 0,
      minTemperature: parseFloat(f[idx++]) || 0,
      lateAcceptanceLen: parseInt(f[idx++]) || 0,
      week: parseInt(f[idx++]) || 0,
      startPenalty: parseInt(f[idx++]) || 0,
      finalPenalty: parseInt(f[idx++]) || 0,
      improvement: parseInt(f[idx++]) || 0,
      hardViolations: parseInt(f[idx++]) || 0,
      softViolations: parseInt(f[idx++]) || 0,
      candidates: parseInt(f[idx++]) || 0,
      accepted: parseInt(f[idx++]) || 0,
      rejected: parseInt(f[idx++]) || 0,
      acceptanceRate: parseFloat(f[idx++]) || 0,
      bestIteration: parseInt(f[idx++]) || 0,
      bestWorkerID: parseInt(f[idx++]) || 0,
      workersStarted: parseInt(f[idx++]) || 0,
      branchesCreated: parseInt(f[idx++]) || 0,
      branchesDropped: parseInt(f[idx++]) || 0,
      maxQueueDepth: parseInt(f[idx++]) || 0,
      maxConcurrentSeen: parseInt(f[idx++]) || 0,
      durationMs: parseInt(f[idx++]) || 0,
      saFinalTemp: parseFloat(f[idx++]) || 0,
      saTempAtBest: parseFloat(f[idx++]) || 0,
      saAcceptedBetter: parseInt(f[idx++]) || 0,
      saAcceptedWorse: parseInt(f[idx++]) || 0,
      saRejectedByProb: parseInt(f[idx++]) || 0,
      lahcAcceptedByCurrent: parseInt(f[idx++]) || 0,
      lahcAcceptedByLate: parseInt(f[idx++]) || 0,
      lahcRejectedByLate: parseInt(f[idx++]) || 0,
      branchesQueued: parseInt(f[idx++]) || 0,
      branchesStarted: parseInt(f[idx++]) || 0,
      branchesCompleted: parseInt(f[idx++]) || 0,
      winningBranchDepth: parseInt(f[idx++]) || 0,
      workersImproved: parseInt(f[idx++]) || 0,
      workersProducedBest: parseInt(f[idx++]) || 0,
      rejectedNoop: parseInt(f[idx++]) || 0,
      rejectedSkill: parseInt(f[idx++]) || 0,
      rejectedSuccession: parseInt(f[idx++]) || 0,
      rejectedHistory: parseInt(f[idx++]) || 0,
    };
    records.push(r);
  }
  return records;
}

export function parseTreeCSV(content: string): TreeNode[] {
  const lines = content.trim().split('\n');
  if (lines.length < 2) return [];

  const nodes: TreeNode[] = [];
  for (let i = 1; i < lines.length; i++) {
    const line = lines[i].trim();
    if (!line) continue;
    const f = line.split(',');
    if (f.length < 18) continue;

    let idx = 0;
    const n: TreeNode = {
      pathID: parseInt(f[idx++]) || 0,
      parentID: parseInt(f[idx++]) || 0,
      week: parseInt(f[idx++]) || 0,
      seed: parseInt(f[idx++]) || 0,
      weekPenalty: parseInt(f[idx++]) || 0,
      cumulativePenalty: parseInt(f[idx++]) || 0,
      workersStarted: parseInt(f[idx++]) || 0,
      candidates: parseInt(f[idx++]) || 0,
      accepted: parseInt(f[idx++]) || 0,
      rejected: parseInt(f[idx++]) || 0,
      saAcceptedBetter: parseInt(f[idx++]) || 0,
      saAcceptedWorse: parseInt(f[idx++]) || 0,
      saRejectedByProb: parseInt(f[idx++]) || 0,
      hardRejectRate: parseFloat(f[idx++]) || 0,
      durationMs: parseInt(f[idx++]) || 0,
      retained: f[idx++] === '1',
      retainedRank: parseInt(f[idx++]) || 0,
      winning: f[idx++] === '1',
    };
    nodes.push(n);
  }
  return nodes;
}

export function parsePlateauCSV(content: string): PlateauEvent[] {
  const lines = content.trim().split('\n');
  if (lines.length < 2) return [];

  const events: PlateauEvent[] = [];
  for (let i = 1; i < lines.length; i++) {
    const line = lines[i].trim();
    if (!line) continue;
    const f = line.split(',');
    if (f.length < 10) continue;

    let idx = 0;
    events.push({
      week: parseInt(f[idx++]) || 0,
      workerID: parseInt(f[idx++]) || 0,
      parentWorkerID: parseInt(f[idx++]) || 0,
      depth: parseInt(f[idx++]) || 0,
      candidate: parseInt(f[idx++]) || 0,
      temperature: parseFloat(f[idx++]) || 0,
      currentPenalty: parseInt(f[idx++]) || 0,
      localBest: parseInt(f[idx++]) || 0,
      globalBest: parseInt(f[idx++]) || 0,
      candsSinceImprove: parseInt(f[idx++]) || 0,
    });
  }
  return events;
}

export function parseBranchCSV(content: string): BranchEvent[] {
  const lines = content.trim().split('\n');
  if (lines.length < 2) return [];

  const events: BranchEvent[] = [];
  for (let i = 1; i < lines.length; i++) {
    const line = lines[i].trim();
    if (!line) continue;
    const f = line.split(',');
    if (f.length < 7) continue;

    let idx = 0;
    events.push({
      week: parseInt(f[idx++]) || 0,
      workerID: parseInt(f[idx++]) || 0,
      candidate: parseInt(f[idx++]) || 0,
      oldPenalty: parseInt(f[idx++]) || 0,
      newPenalty: parseInt(f[idx++]) || 0,
      improvement: parseInt(f[idx++]) || 0,
      timestampMs: parseInt(f[idx++]) || 0,
    });
  }
  return events;
}

export function parseWorkerLifecycleCSV(content: string): WorkerLifecycle[] {
  const lines = content.trim().split('\n');
  if (lines.length < 2) return [];

  const workers: WorkerLifecycle[] = [];
  for (let i = 1; i < lines.length; i++) {
    const line = lines[i].trim();
    if (!line) continue;
    const f = line.split(',');
    if (f.length < 18) continue;

    let idx = 0;
    workers.push({
      workerID: parseInt(f[idx++]) || 0,
      parentWorkerID: parseInt(f[idx++]) || 0,
      week: parseInt(f[idx++]) || 0,
      seed: parseInt(f[idx++]) || 0,
      depth: parseInt(f[idx++]) || 0,
      startTimeMs: parseInt(f[idx++]) || 0,
      finishTimeMs: parseInt(f[idx++]) || 0,
      finishCandidate: parseInt(f[idx++]) || 0,
      initialTemperature: parseFloat(f[idx++]) || 0,
      finalTemperature: parseFloat(f[idx++]) || 0,
      temperatureAtBest: parseFloat(f[idx++]) || 0,
      bestCandidate: parseInt(f[idx++]) || 0,
      plateauCount: parseInt(f[idx++]) || 0,
      branchCount: parseInt(f[idx++]) || 0,
      producedGlobalBest: f[idx++] === '1',
      finalPenalty: parseInt(f[idx++]) || 0,
      bestPenalty: parseInt(f[idx++]) || 0,
      startPenalty: parseInt(f[idx++]) || 0,
    });
  }
  return workers;
}
