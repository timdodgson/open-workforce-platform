import { readFile } from 'fs/promises';
import { existsSync, readdirSync } from 'fs';
import path from 'path';
import { parseAuditCSV, parseTreeCSV, parsePlateauCSV, parseBranchCSV, parseWorkerLifecycleCSV, parseImprovementsCSV, parseDiversityCSV, parseDiscoveriesCSV } from './csv-parser';
import { RunMetadata, PreviousBest, RunSummary, WeekRecord, TreeNode, PlateauEvent, BranchEvent, WorkerLifecycle, ImprovementEvent, DiversityRecord, DiscoveryRecord } from './types';
import { getStorageProvider } from './storage';
import { cached } from './cache';

const DATA_DIR = path.join(process.cwd(), 'data');
const RUNS_DIR = path.join(DATA_DIR, 'runs');

interface ManifestRunEntry {
  runId: string;
  label: string;
  algorithm: string;
  timestamp: string;
  totalPenalty: number;
}

export interface RunListEntry {
  id: string;
  metadata: RunMetadata | null;
  timestamp?: string;
  manifestPenalty?: number;
}

async function readManifestIndex(storage: ReturnType<typeof getStorageProvider>): Promise<Map<string, ManifestRunEntry>> {
  return cached('manifestIndex', async () => {
    const map = new Map<string, ManifestRunEntry>();
    const content = await storage.readRootFile('manifest.json');
    if (!content) return map;
    try {
      const manifest = JSON.parse(content) as { runs?: ManifestRunEntry[] };
      for (const entry of manifest.runs || []) {
        if (entry.runId) map.set(entry.runId, entry);
      }
    } catch { /* ignore */ }
    return map;
  });
}

/** Best objective value from run.json metadata, or 0 when unavailable. */
export function objectiveFromMetadata(
  metadata: RunMetadata | Record<string, unknown> | null | undefined,
  mode?: string
): number {
  if (!metadata) return 0;
  const meta = metadata as Record<string, unknown>;
  const modeStr = mode ?? String(meta.mode || '');
  if (meta.bestObjective && Number(meta.bestObjective) > 0) return Number(meta.bestObjective);
  if (meta.bestDistance && Number(meta.bestDistance) > 0) return Number(meta.bestDistance);
  if (meta.bestMakespan && Number(meta.bestMakespan) > 0) return Number(meta.bestMakespan);
  if (meta.totalPenalty && Number(meta.totalPenalty) > 0) return Number(meta.totalPenalty);
  if (modeStr === 'ilp' && meta.objective && Number(meta.objective) > 0) return Number(meta.objective);
  return 0;
}

/** Empty summary for pages that only need metadata-derived objectives. */
export function emptyRunSummary(metadata: RunMetadata | null = null): RunSummary {
  return {
    metadata,
    previousBest: null,
    weeks: [],
    totalPenalty: 0,
    totalCandidates: 0,
    totalAccepted: 0,
    totalRejected: 0,
    totalSABetter: 0,
    totalSAWorse: 0,
    totalSARejProb: 0,
    totalWorkers: 0,
    totalBranches: 0,
    totalDurationMs: 0,
    hardRejectRate: 0,
    acceptWorseRate: 0,
    lahcAcceptByLateRate: 0,
    cumulativePenalties: [],
    numWeeks: 0,
    maxWeekPenalty: 0,
    maxWeekNum: 0,
  };
}

// Resolve the data directory for a given run ID. If null, uses root data/.
function resolveDataDir(runId?: string | null): string {
  if (runId) {
    const runDir = path.join(RUNS_DIR, runId);
    if (existsSync(runDir)) return runDir;
  }
  return DATA_DIR;
}

// List all available runs.
export async function listRunsAsync(): Promise<RunListEntry[]> {
  return cached('listRuns', async () => {
    const storage = getStorageProvider();
    const [runIds, manifestIndex] = await Promise.all([
      storage.listRuns(),
      readManifestIndex(storage),
    ]);
    return Promise.all(
      runIds.map(async (id): Promise<RunListEntry> => {
        const manifestEntry = manifestIndex.get(id);
        const content = await storage.readFile(id, 'run.json');
        let metadata: RunMetadata | null = null;
        if (content) {
          try { metadata = JSON.parse(content) as RunMetadata; } catch { /* ignore */ }
        }
        return { id, metadata, timestamp: manifestEntry?.timestamp, manifestPenalty: manifestEntry?.totalPenalty };
      })
    );
  });
}

// Synchronous version for backwards compatibility (local only).
export function listRuns(): { id: string; metadata: RunMetadata | null }[] {
  // For listRuns we use the local provider synchronously for now.
  // Full async migration happens when S3 becomes primary.
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
  if (runId) {
    return cached(`runMeta:${runId}`, async () => {
      const content = await getStorageProvider().readFile(runId, 'run.json');
      if (!content) return null;
      return JSON.parse(content) as RunMetadata;
    });
  }
  // Fallback to root data/ for backwards compatibility.
  return cached('runMeta:root', async () => {
    const dir = resolveDataDir(runId);
    const p = path.join(dir, 'run.json');
    if (!existsSync(p)) return null;
    const content = await readFile(p, 'utf-8');
    return JSON.parse(content) as RunMetadata;
  });
}

export async function loadPreviousBest(runId?: string | null): Promise<PreviousBest | null> {
  if (runId) {
    const content = await getStorageProvider().readFile(runId, 'best.json');
    if (!content) return null;
    try { return JSON.parse(content) as PreviousBest; } catch { return null; }
  }
  const dir = resolveDataDir(runId);
  const p = path.join(dir, 'best.json');
  if (!existsSync(p)) return null;
  const content = await readFile(p, 'utf-8');
  return JSON.parse(content) as PreviousBest;
}

export async function loadWeeks(runId?: string | null): Promise<WeekRecord[]> {
  const cacheKey = runId ? `weeks:${runId}` : 'weeks:root';
  return cached(cacheKey, async () => {
    if (runId) {
      const content = await getStorageProvider().readFile(runId, 'results.csv');
      if (!content) return [];
      return parseAuditCSV(content);
    }
    const dir = resolveDataDir(runId);
    const p = path.join(dir, 'results.csv');
    if (!existsSync(p)) return [];
    const content = await readFile(p, 'utf-8');
    return parseAuditCSV(content);
  });
}

export async function loadTree(runId?: string | null): Promise<TreeNode[]> {
  if (runId) {
    const content = await getStorageProvider().readFile(runId, 'tree.csv');
    if (!content) return [];
    return parseTreeCSV(content);
  }
  const dir = resolveDataDir(runId);
  const p = path.join(dir, 'tree.csv');
  if (!existsSync(p)) return [];
  const content = await readFile(p, 'utf-8');
  return parseTreeCSV(content);
}

export async function loadPlateaus(runId?: string | null): Promise<PlateauEvent[]> {
  if (runId) {
    const content = await getStorageProvider().readFile(runId, 'plateaus.csv');
    if (!content) return [];
    return parsePlateauCSV(content);
  }
  const dir = resolveDataDir(runId);
  const p = path.join(dir, 'plateaus.csv');
  if (!existsSync(p)) return [];
  const content = await readFile(p, 'utf-8');
  return parsePlateauCSV(content);
}

export async function loadBranches(runId?: string | null): Promise<BranchEvent[]> {
  if (runId) {
    const content = await getStorageProvider().readFile(runId, 'branches.csv');
    if (!content) return [];
    return parseBranchCSV(content);
  }
  const dir = resolveDataDir(runId);
  const p = path.join(dir, 'branches.csv');
  if (!existsSync(p)) return [];
  const content = await readFile(p, 'utf-8');
  return parseBranchCSV(content);
}

export async function loadWorkerLifecycles(runId?: string | null): Promise<WorkerLifecycle[]> {
  if (runId) {
    const content = await getStorageProvider().readFile(runId, 'workers.csv');
    if (!content) return [];
    return parseWorkerLifecycleCSV(content);
  }
  const dir = resolveDataDir(runId);
  const p = path.join(dir, 'workers.csv');
  if (!existsSync(p)) return [];
  const content = await readFile(p, 'utf-8');
  return parseWorkerLifecycleCSV(content);
}

export async function loadImprovements(runId?: string | null): Promise<ImprovementEvent[]> {
  if (runId) {
    const content = await getStorageProvider().readFile(runId, 'improvements.csv');
    if (!content) return [];
    return parseImprovementsCSV(content);
  }
  const dir = resolveDataDir(runId);
  const p = path.join(dir, 'improvements.csv');
  if (!existsSync(p)) return [];
  const content = await readFile(p, 'utf-8');
  return parseImprovementsCSV(content);
}

export async function loadDiversity(runId?: string | null): Promise<DiversityRecord[]> {
  if (runId) {
    const content = await getStorageProvider().readFile(runId, 'diversity.csv');
    if (!content) return [];
    return parseDiversityCSV(content);
  }
  const dir = resolveDataDir(runId);
  const p = path.join(dir, 'diversity.csv');
  if (!existsSync(p)) return [];
  const content = await readFile(p, 'utf-8');
  return parseDiversityCSV(content);
}

export async function loadDiscoveries(runId?: string | null): Promise<DiscoveryRecord[]> {
  if (runId) {
    const content = await getStorageProvider().readFile(runId, 'discoveries.csv');
    if (!content) return [];
    return parseDiscoveriesCSV(content);
  }
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
  if (runId) {
    const content = await getStorageProvider().readFile(runId, 'roster.json');
    if (!content) return [];
    try { return JSON.parse(content) as RosterEntry[]; } catch { return []; }
  }
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
