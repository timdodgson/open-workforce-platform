import { readFile } from 'fs/promises';
import { existsSync, readdirSync } from 'fs';
import path from 'path';
import { parseAuditCSV, parseTreeCSV, parsePlateauCSV, parseBranchCSV, parseWorkerLifecycleCSV, parseImprovementsCSV, parseDiversityCSV, parseDiscoveriesCSV } from './csv-parser';
import { RunMetadata, PreviousBest, RunSummary, WeekRecord, TreeNode, PlateauEvent, BranchEvent, WorkerLifecycle, ImprovementEvent, DiversityRecord, DiscoveryRecord } from './types';

const DATA_DIR = path.join(process.cwd(), 'data');
const RUNS_DIR = path.join(DATA_DIR, 'runs');

// Resolve the data directory for a given run ID. If null, uses root data/.
function resolveDataDir(runId?: string | null): string {
  if (runId) {
    const runDir = path.join(RUNS_DIR, runId);
    if (existsSync(runDir)) return runDir;
  }
  return DATA_DIR;
}

// List all available runs.
export function listRuns(): { id: string; metadata: RunMetadata | null }[] {
  if (!existsSync(RUNS_DIR)) return [];
  const entries = readdirSync(RUNS_DIR, { withFileTypes: true });
  const runs: { id: string; metadata: RunMetadata | null }[] = [];
  for (const entry of entries) {
    if (!entry.isDirectory()) continue;
    const runJsonPath = path.join(RUNS_DIR, entry.name, 'run.json');
    let metadata: RunMetadata | null = null;
    if (existsSync(runJsonPath)) {
      try {
        const content = require('fs').readFileSync(runJsonPath, 'utf-8');
        metadata = JSON.parse(content) as RunMetadata;
      } catch { /* ignore parse errors */ }
    }
    runs.push({ id: entry.name, metadata });
  }
  return runs;
}

export async function loadRunMetadata(runId?: string | null): Promise<RunMetadata | null> {
  const dir = resolveDataDir(runId);
  const p = path.join(dir, 'run.json');
  if (!existsSync(p)) return null;
  const content = await readFile(p, 'utf-8');
  return JSON.parse(content) as RunMetadata;
}

export async function loadPreviousBest(runId?: string | null): Promise<PreviousBest | null> {
  const dir = resolveDataDir(runId);
  const p = path.join(dir, 'best.json');
  if (!existsSync(p)) return null;
  const content = await readFile(p, 'utf-8');
  return JSON.parse(content) as PreviousBest;
}

export async function loadWeeks(runId?: string | null): Promise<WeekRecord[]> {
  const dir = resolveDataDir(runId);
  const p = path.join(dir, 'results.csv');
  if (!existsSync(p)) return [];
  const content = await readFile(p, 'utf-8');
  return parseAuditCSV(content);
}

export async function loadTree(runId?: string | null): Promise<TreeNode[]> {
  const dir = resolveDataDir(runId);
  const p = path.join(dir, 'tree.csv');
  if (!existsSync(p)) return [];
  const content = await readFile(p, 'utf-8');
  return parseTreeCSV(content);
}

export async function loadPlateaus(runId?: string | null): Promise<PlateauEvent[]> {
  const dir = resolveDataDir(runId);
  const p = path.join(dir, 'plateaus.csv');
  if (!existsSync(p)) return [];
  const content = await readFile(p, 'utf-8');
  return parsePlateauCSV(content);
}

export async function loadBranches(runId?: string | null): Promise<BranchEvent[]> {
  const dir = resolveDataDir(runId);
  const p = path.join(dir, 'branches.csv');
  if (!existsSync(p)) return [];
  const content = await readFile(p, 'utf-8');
  return parseBranchCSV(content);
}

export async function loadWorkerLifecycles(runId?: string | null): Promise<WorkerLifecycle[]> {
  const dir = resolveDataDir(runId);
  const p = path.join(dir, 'workers.csv');
  if (!existsSync(p)) return [];
  const content = await readFile(p, 'utf-8');
  return parseWorkerLifecycleCSV(content);
}

export async function loadImprovements(runId?: string | null): Promise<ImprovementEvent[]> {
  const dir = resolveDataDir(runId);
  const p = path.join(dir, 'improvements.csv');
  if (!existsSync(p)) return [];
  const content = await readFile(p, 'utf-8');
  return parseImprovementsCSV(content);
}

export async function loadDiversity(runId?: string | null): Promise<DiversityRecord[]> {
  const dir = resolveDataDir(runId);
  const p = path.join(dir, 'diversity.csv');
  if (!existsSync(p)) return [];
  const content = await readFile(p, 'utf-8');
  return parseDiversityCSV(content);
}

export async function loadDiscoveries(runId?: string | null): Promise<DiscoveryRecord[]> {
  const dir = resolveDataDir(runId);
  const p = path.join(dir, 'discoveries.csv');
  if (!existsSync(p)) return [];
  const content = await readFile(p, 'utf-8');
  return parseDiscoveriesCSV(content);
}

export interface RosterEntry {
  week: number;
  nurse: string;
  day: string;
  dayIndex: number;
  shiftType: string;
  skill: string;
  contract: string;
  nurseSkills: string[];
}

export async function loadRoster(runId?: string | null): Promise<RosterEntry[]> {
  const dir = resolveDataDir(runId);
  const p = path.join(dir, 'roster.json');
  if (!existsSync(p)) return [];
  const content = await readFile(p, 'utf-8');
  try {
    return JSON.parse(content) as RosterEntry[];
  } catch {
    return [];
  }
}

export async function loadRunSummary(runId?: string | null): Promise<RunSummary> {
  const [metadata, previousBest, weeks] = await Promise.all([
    loadRunMetadata(runId),
    loadPreviousBest(runId),
    loadWeeks(runId),
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

  let totalLAHCByLate = 0;
  for (const w of weeks) {
    totalLAHCByLate += w.lahcAcceptedByLate;
  }
  const lahcAcceptByLateRate = totalCandidates > 0
    ? (totalLAHCByLate / totalCandidates) * 100 : 0;

  return {
    metadata, previousBest, weeks, totalPenalty, totalCandidates,
    totalAccepted, totalRejected, totalSABetter, totalSAWorse,
    totalSARejProb, totalWorkers, totalBranches, totalDurationMs,
    hardRejectRate, acceptWorseRate, lahcAcceptByLateRate, cumulativePenalties,
    numWeeks: weeks.length, maxWeekPenalty, maxWeekNum,
  };
}
