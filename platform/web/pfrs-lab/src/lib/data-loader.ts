import { readFile } from 'fs/promises';
import { existsSync } from 'fs';
import path from 'path';
import { parseAuditCSV, parseTreeCSV, parsePlateauCSV, parseBranchCSV, parseWorkerLifecycleCSV, parseImprovementsCSV, parseDiversityCSV, parseDiscoveriesCSV } from './csv-parser';
import { RunMetadata, PreviousBest, RunSummary, WeekRecord, TreeNode, PlateauEvent, BranchEvent, WorkerLifecycle, ImprovementEvent, DiversityRecord, DiscoveryRecord } from './types';

const DATA_DIR = path.join(process.cwd(), 'data');

export async function loadRunMetadata(): Promise<RunMetadata | null> {
  const p = path.join(DATA_DIR, 'run.json');
  if (!existsSync(p)) return null;
  const content = await readFile(p, 'utf-8');
  return JSON.parse(content) as RunMetadata;
}

export async function loadPreviousBest(): Promise<PreviousBest | null> {
  const p = path.join(DATA_DIR, 'best.json');
  if (!existsSync(p)) return null;
  const content = await readFile(p, 'utf-8');
  return JSON.parse(content) as PreviousBest;
}

export async function loadWeeks(): Promise<WeekRecord[]> {
  const p = path.join(DATA_DIR, 'results.csv');
  if (!existsSync(p)) return [];
  const content = await readFile(p, 'utf-8');
  return parseAuditCSV(content);
}

export async function loadTree(): Promise<TreeNode[]> {
  const p = path.join(DATA_DIR, 'tree.csv');
  if (!existsSync(p)) return [];
  const content = await readFile(p, 'utf-8');
  return parseTreeCSV(content);
}

export async function loadPlateaus(): Promise<PlateauEvent[]> {
  const p = path.join(DATA_DIR, 'plateaus.csv');
  if (!existsSync(p)) return [];
  const content = await readFile(p, 'utf-8');
  return parsePlateauCSV(content);
}

export async function loadBranches(): Promise<BranchEvent[]> {
  const p = path.join(DATA_DIR, 'branches.csv');
  if (!existsSync(p)) return [];
  const content = await readFile(p, 'utf-8');
  return parseBranchCSV(content);
}

export async function loadWorkerLifecycles(): Promise<WorkerLifecycle[]> {
  const p = path.join(DATA_DIR, 'workers.csv');
  if (!existsSync(p)) return [];
  const content = await readFile(p, 'utf-8');
  return parseWorkerLifecycleCSV(content);
}

export async function loadImprovements(): Promise<ImprovementEvent[]> {
  const p = path.join(DATA_DIR, 'improvements.csv');
  if (!existsSync(p)) return [];
  const content = await readFile(p, 'utf-8');
  return parseImprovementsCSV(content);
}

export async function loadDiversity(): Promise<DiversityRecord[]> {
  const p = path.join(DATA_DIR, 'diversity.csv');
  if (!existsSync(p)) return [];
  const content = await readFile(p, 'utf-8');
  return parseDiversityCSV(content);
}

export async function loadDiscoveries(): Promise<DiscoveryRecord[]> {
  const p = path.join(DATA_DIR, 'discoveries.csv');
  if (!existsSync(p)) return [];
  const content = await readFile(p, 'utf-8');
  return parseDiscoveriesCSV(content);
}

export async function loadRunSummary(): Promise<RunSummary> {
  const [metadata, previousBest, weeks] = await Promise.all([
    loadRunMetadata(),
    loadPreviousBest(),
    loadWeeks(),
  ]);
  return computeSummary(weeks, metadata, previousBest);
}

function computeSummary(
  weeks: WeekRecord[],
  metadata: RunMetadata | null,
  previousBest: PreviousBest | null
): RunSummary {
  let totalPenalty = 0, totalCandidates = 0, totalAccepted = 0;
  let totalRejected = 0, totalSABetter = 0, totalSAWorse = 0;
  let totalSARejProb = 0, totalWorkers = 0, totalBranches = 0;
  let totalDurationMs = 0, maxWeekPenalty = 0, maxWeekNum = 0;
  const cumulativePenalties: number[] = [];
  let cum = 0;

  for (const w of weeks) {
    totalPenalty += w.finalPenalty;
    totalCandidates += w.candidates;
    totalAccepted += w.accepted;
    totalRejected += w.rejected;
    totalSABetter += w.saAcceptedBetter;
    totalSAWorse += w.saAcceptedWorse;
    totalSARejProb += w.saRejectedByProb;
    totalWorkers += w.workersStarted;
    totalBranches += w.branchesCreated;
    totalDurationMs += w.durationMs;
    if (w.finalPenalty > maxWeekPenalty) {
      maxWeekPenalty = w.finalPenalty;
      maxWeekNum = w.week;
    }
    cum += w.finalPenalty;
    cumulativePenalties.push(cum);
  }

  const hardRejectRate = (totalCandidates + totalRejected) > 0
    ? (totalRejected / (totalCandidates + totalRejected)) * 100 : 0;
  const acceptWorseRate = totalCandidates > 0
    ? (totalSAWorse / totalCandidates) * 100 : 0;

  return {
    metadata, previousBest, weeks, totalPenalty, totalCandidates,
    totalAccepted, totalRejected, totalSABetter, totalSAWorse,
    totalSARejProb, totalWorkers, totalBranches, totalDurationMs,
    hardRejectRate, acceptWorseRate, cumulativePenalties,
    numWeeks: weeks.length, maxWeekPenalty, maxWeekNum,
  };
}
