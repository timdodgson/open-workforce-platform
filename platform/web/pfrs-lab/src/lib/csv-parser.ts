import { WeekRecord, TreeNode, PlateauEvent, BranchEvent, WorkerLifecycle, ImprovementEvent, DiversityRecord } from './types';

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
    if (f.length < 22) continue;

    // Skip 8 run context columns. Event fields start at index 8.
    events.push({
      week: parseInt(f[8]) || 0,
      workerID: parseInt(f[9]) || 0,
      parentWorkerID: parseInt(f[10]) || 0,
      depth: parseInt(f[11]) || 0,
      candidate: parseInt(f[12]) || 0,
      temperature: parseFloat(f[13]) || 0,
      currentPenalty: parseInt(f[14]) || 0,
      localBest: parseInt(f[15]) || 0,
      globalBest: parseInt(f[16]) || 0,
      candsSinceImprove: parseInt(f[17]) || 0,
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
    if (f.length < 20) continue;

    // Skip 8 run context columns, then read event fields.
    events.push({
      week: parseInt(f[8]) || 0,
      workerID: parseInt(f[9]) || 0,
      candidate: parseInt(f[14]) || 0,
      oldPenalty: parseInt(f[16]) || 0,
      newPenalty: parseInt(f[17]) || 0,
      improvement: parseInt(f[18]) || 0,
      timestampMs: parseInt(f[19]) || 0,
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
    if (f.length < 35) continue;

    // Skip 8 run context columns. Worker fields start at index 8.
    workers.push({
      workerID: parseInt(f[8]) || 0,
      parentWorkerID: parseInt(f[9]) || 0,
      week: parseInt(f[10]) || 0,
      seed: parseInt(f[11]) || 0,
      depth: parseInt(f[12]) || 0,
      startTimeMs: parseInt(f[13]) || 0,
      finishTimeMs: parseInt(f[14]) || 0,
      finishCandidate: parseInt(f[15]) || 0,
      initialTemperature: parseFloat(f[16]) || 0,
      finalTemperature: parseFloat(f[17]) || 0,
      temperatureAtBest: parseFloat(f[18]) || 0,
      bestCandidate: parseInt(f[19]) || 0,
      plateauCount: parseInt(f[20]) || 0,
      branchCount: parseInt(f[21]) || 0,
      producedGlobalBest: f[22] === '1',
      finalPenalty: parseInt(f[23]) || 0,
      bestPenalty: parseInt(f[24]) || 0,
      startPenalty: parseInt(f[25]) || 0,
    });
  }
  return workers;
}

export function parseImprovementsCSV(content: string): ImprovementEvent[] {
  const lines = content.trim().split('\n');
  if (lines.length < 2) return [];

  const events: ImprovementEvent[] = [];
  for (let i = 1; i < lines.length; i++) {
    const line = lines[i].trim();
    if (!line) continue;
    const f = line.split(',');
    if (f.length < 16) continue;

    // Skip run context fields (0-7), read event fields (8-15).
    events.push({
      week: parseInt(f[8]) || 0,
      workerID: parseInt(f[9]) || 0,
      candidate: parseInt(f[10]) || 0,
      temperatureAtEvent: parseFloat(f[11]) || 0,
      oldGlobalBest: parseInt(f[12]) || 0,
      newGlobalBest: parseInt(f[13]) || 0,
      improvement: parseInt(f[14]) || 0,
      elapsedMs: parseInt(f[15]) || 0,
    });
  }
  return events;
}

export function parseDiversityCSV(content: string): DiversityRecord[] {
  const lines = content.trim().split('\n');
  if (lines.length < 2) return [];

  const records: DiversityRecord[] = [];
  for (let i = 1; i < lines.length; i++) {
    const line = lines[i].trim();
    if (!line) continue;
    const f = line.split(',');
    if (f.length < 20) continue;

    // Skip 8 run context columns. Diversity fields start at index 8.
    records.push({
      week: parseInt(f[8]) || 0,
      pathID: parseInt(f[9]) || 0,
      fingerprint: f[10] || '',
      hammingToBest: parseFloat(f[11]) || 0,
      hammingToParent: parseFloat(f[12]) || 0,
      beamSpread: parseInt(f[13]) || 0,
      nearDuplicate: f[14] === '1',
      retained: f[15] === '1',
      retainedRank: parseInt(f[16]) || 0,
      winning: f[17] === '1',
      cumulativePenalty: parseInt(f[18]) || 0,
      weekPenalty: parseInt(f[19]) || 0,
    });
  }
  return records;
}
