import { parseAuditCSV, parseTreeCSV, parsePlateauCSV, parseBranchCSV, parseWorkerLifecycleCSV, parseImprovementsCSV, parseDiversityCSV, parseDiscoveriesCSV } from './csv-parser';
import { RunMetadata, PreviousBest, RunSummary, WeekRecord, TreeNode, PlateauEvent, BranchEvent, WorkerLifecycle, ImprovementEvent, DiversityRecord, DiscoveryRecord } from './types';
import { getStorageProvider } from './storage';
import { cached } from './cache';
import { resolveProblemType } from './resolve-problem-type';
import { isRosterSolution } from './roster-types';

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
  }, 120_000);
}

function parseRunIdFields(runId: string, entry: ManifestRunEntry): {
  instance: string;
  mode: string;
  problemType: string;
} {
  const lower = runId.toLowerCase();
  const parts = runId.split('-');

  const problemType = resolveProblemType(runId, { problemType: inferProblemTypeFromId(runId) });

  let mode = entry.algorithm?.toLowerCase() || 'unknown';
  const modeMatch = runId.match(/-(sa|lahc|tabu|portfolio|ilp|hybrid|beam)-/i);
  if (modeMatch) mode = modeMatch[1].toLowerCase();
  if (lower.includes('-ilp-') || lower.startsWith('ilp-')) mode = 'ilp';

  let instance = '—';
  if (problemType === 'cvrp' || problemType === 'vrptw') {
    instance = parts.find((p) => /^[a-z]\d/i.test(p)) || parts[1] || '—';
  } else if (problemType === 'jss') {
    instance = parts.find((p) => /^(la|ft)\d/i.test(p)) || parts[1] || '—';
  } else {
    instance = parts.find((p) => /^n\d|^w\d/i.test(p)) || parts[0] || '—';
  }

  return { instance, mode, problemType };
}

/** Manifest-only hint before run.json is loaded. */
function inferProblemTypeFromId(runId: string): string | undefined {
  const lower = runId.toLowerCase();
  if (lower.includes('-vrptw-') || lower.startsWith('vrptw')) return 'vrptw';
  if (lower.includes('-cvrp-') || lower.startsWith('cvrp')) return 'cvrp';
  if (lower.includes('-jss-') || lower.startsWith('jss') || lower.includes('jobshop')) return 'jss';
  if (lower.includes('-nrp-') || lower.startsWith('nrp') || lower.includes('bench-nrp')) return 'nrp';
  return undefined;
}

/** Synthesize list-view metadata from manifest — avoids per-run S3 run.json reads. */
function metadataFromManifest(entry: ManifestRunEntry): RunMetadata {
  const { instance, mode, problemType } = parseRunIdFields(entry.runId, entry);
  const meta = {
    instance,
    algorithm: entry.algorithm || mode,
    mode,
    problemType,
    runLabel: entry.label,
    iterationsPerWorker: 0,
    initialTemperature: 0,
    coolingMode: '',
    effectiveCoolingRate: 0,
    beamWidth: 0,
    beamSeeds: [] as number[],
    seed: 0,
    cpus: 0,
    maxTotalWorkers: 0,
    totalPenalty: entry.totalPenalty,
    bestObjective: entry.totalPenalty,
  };
  return meta as unknown as RunMetadata;
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

// List all available runs (manifest-only — one S3 read, not 600+ run.json fetches).
export async function listRunsAsync(): Promise<RunListEntry[]> {
  return cached('listRunsManifest', async () => {
    const storage = getStorageProvider();
    const manifestIndex = await readManifestIndex(storage);
    return [...manifestIndex.values()]
      .sort((a, b) => (b.timestamp || b.runId).localeCompare(a.timestamp || a.runId))
      .map((entry): RunListEntry => ({
        id: entry.runId,
        metadata: metadataFromManifest(entry),
        timestamp: entry.timestamp,
        manifestPenalty: entry.totalPenalty,
      }));
  }, 120_000);
}

/** Newest run ID from manifest (for /api/runs/latest/*). */
export async function getLatestRunId(): Promise<string | null> {
  const runs = await listRunsAsync();
  return runs[0]?.id ?? null;
}

const BENCHMARK_MAX_RUNS = 300;
const BENCHMARK_BATCH_SIZE = 15;

/** Manifest-first run list for /benchmarks — avoids 600+ S3 run.json reads per request. */
export async function listBenchmarkRunsAsync(): Promise<RunListEntry[]> {
  return cached('listBenchmarkRuns', async () => {
    const storage = getStorageProvider();
    const manifestIndex = await readManifestIndex(storage);
    const all = [...manifestIndex.values()].filter((e) => e.totalPenalty > 0);
    const siRuns = all.filter((e) => e.runId.startsWith('val-') || e.runId.startsWith('si2-'));
    const otherRuns = all
      .filter((e) => !e.runId.startsWith('val-') && !e.runId.startsWith('si2-'))
      .sort((a, b) => (b.timestamp || '').localeCompare(a.timestamp || ''));
    const cap = Math.max(0, BENCHMARK_MAX_RUNS - siRuns.length);
    const candidates = [...siRuns, ...otherRuns.slice(0, cap)];

    const runs: RunListEntry[] = [];
    for (let i = 0; i < candidates.length; i += BENCHMARK_BATCH_SIZE) {
      const batch = candidates.slice(i, i + BENCHMARK_BATCH_SIZE);
      const batchRuns = await Promise.all(
        batch.map(async (entry): Promise<RunListEntry> => {
          const content = await storage.readFile(entry.runId, 'run.json');
          let metadata: RunMetadata | null = null;
          if (content) {
            try { metadata = JSON.parse(content) as RunMetadata; } catch { /* ignore */ }
          }
          return {
            id: entry.runId,
            metadata,
            timestamp: entry.timestamp,
            manifestPenalty: entry.totalPenalty,
          };
        }),
      );
      runs.push(...batchRuns);
    }
    return runs;
  });
}

export async function loadRunMetadata(runId: string): Promise<RunMetadata | null> {
  return cached(`runMeta:${runId}`, async () => {
    const content = await getStorageProvider().readFile(runId, 'run.json');
    if (!content) return null;
    return JSON.parse(content) as RunMetadata;
  });
}

export async function loadPreviousBest(runId: string): Promise<PreviousBest | null> {
  const content = await getStorageProvider().readFile(runId, 'best.json');
  if (!content) return null;
  try { return JSON.parse(content) as PreviousBest; } catch { return null; }
}

export async function loadWeeks(runId: string): Promise<WeekRecord[]> {
  return cached(`weeks:${runId}`, async () => {
    const content = await getStorageProvider().readFile(runId, 'results.csv');
    if (!content) return [];
    return parseAuditCSV(content);
  });
}

export async function loadTree(runId: string): Promise<TreeNode[]> {
  const content = await getStorageProvider().readFile(runId, 'tree.csv');
  if (!content) return [];
  return parseTreeCSV(content);
}

export async function loadPlateaus(runId: string): Promise<PlateauEvent[]> {
  const content = await getStorageProvider().readFile(runId, 'plateaus.csv');
  if (!content) return [];
  return parsePlateauCSV(content);
}

export async function loadBranches(runId: string): Promise<BranchEvent[]> {
  const content = await getStorageProvider().readFile(runId, 'branches.csv');
  if (!content) return [];
  return parseBranchCSV(content);
}

export async function loadWorkerLifecycles(runId: string): Promise<WorkerLifecycle[]> {
  const content = await getStorageProvider().readFile(runId, 'workers.csv');
  if (!content) return [];
  return parseWorkerLifecycleCSV(content);
}

export async function loadImprovements(runId: string): Promise<ImprovementEvent[]> {
  const content = await getStorageProvider().readFile(runId, 'improvements.csv');
  if (!content) return [];
  return parseImprovementsCSV(content);
}

export async function loadDiversity(runId: string): Promise<DiversityRecord[]> {
  const content = await getStorageProvider().readFile(runId, 'diversity.csv');
  if (!content) return [];
  return parseDiversityCSV(content);
}

export async function loadDiscoveries(runId: string): Promise<DiscoveryRecord[]> {
  const content = await getStorageProvider().readFile(runId, 'discoveries.csv');
  if (!content) return [];
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

export async function loadRoster(runId: string): Promise<RosterEntry[]> {
  const storage = getStorageProvider();
  const rosterContent = await storage.readFile(runId, 'roster.json');
  if (rosterContent) {
    try {
      const parsed = JSON.parse(rosterContent);
      if (isRosterSolution(parsed)) return parsed;
    } catch { /* fall through */ }
  }

  // Single-week NRP (owp solve nrp) stores roster rows in solution.json.
  const solutionContent = await storage.readFile(runId, 'solution.json');
  if (solutionContent) {
    try {
      const parsed = JSON.parse(solutionContent);
      if (isRosterSolution(parsed)) return parsed;
    } catch { /* graceful */ }
  }

  return [];
}

/** Whether a run has nurse schedule grid data (roster.json or NRP solution.json). */
export async function hasRosterData(runId: string): Promise<boolean> {
  const roster = await loadRoster(runId);
  return roster.length > 0;
}

export async function loadRunSummary(runId: string): Promise<RunSummary> {
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
