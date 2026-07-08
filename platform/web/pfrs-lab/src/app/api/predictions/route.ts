import { NextRequest, NextResponse } from 'next/server';
import { getStorageProvider } from '@/lib/storage';
import type { WorkerPrediction } from '@/app/predictions/page.types';

export const dynamic = 'force-dynamic';

interface PredictionIndex {
  version: string;
  total_predictions: number;
  runs: { runId: string; count: number }[];
}

const DEFAULT_LIMIT = 500;
const MAX_LIMIT = 2000;

async function readIndex(): Promise<PredictionIndex | null> {
  const storage = getStorageProvider();
  const content = await storage.readRootFile('worker_predictions_index.json');
  if (!content) return null;
  try {
    return JSON.parse(content) as PredictionIndex;
  } catch {
    return null;
  }
}

async function loadRunPredictions(runId: string): Promise<WorkerPrediction[]> {
  const storage = getStorageProvider();
  const content = await storage.readFile(runId, 'worker_predictions.json');
  if (!content) return [];
  try {
    const parsed = JSON.parse(content);
    return Array.isArray(parsed) ? parsed as WorkerPrediction[] : [];
  } catch {
    return [];
  }
}

/** Paginated predictions API — reads per-run shards, never the monolithic root file. */
export async function GET(request: NextRequest) {
  const index = await readIndex();
  if (!index?.runs?.length) {
    return NextResponse.json({
      version: '0.0.0',
      total_predictions: 0,
      predictions: [],
      offset: 0,
      limit: 0,
      has_more: false,
    });
  }

  const params = request.nextUrl.searchParams;
  const runId = params.get('runId') || '';
  const limit = Math.min(
    Math.max(parseInt(params.get('limit') || String(DEFAULT_LIMIT), 10) || DEFAULT_LIMIT, 1),
    MAX_LIMIT,
  );
  const offset = Math.max(parseInt(params.get('offset') || '0', 10) || 0, 0);

  const runs = runId
    ? index.runs.filter((r) => r.runId === runId)
    : [...index.runs].sort((a, b) => b.runId.localeCompare(a.runId));

  const collected: WorkerPrediction[] = [];
  let skipped = 0;

  for (const entry of runs) {
    const rows = await loadRunPredictions(entry.runId);
    for (const row of rows) {
      if (skipped < offset) {
        skipped++;
        continue;
      }
      collected.push(row);
      if (collected.length >= limit) {
        return NextResponse.json({
          version: index.version,
          total_predictions: index.total_predictions,
          predictions: collected,
          offset,
          limit,
          has_more: offset + collected.length < index.total_predictions,
        });
      }
    }
  }

  return NextResponse.json({
    version: index.version,
    total_predictions: index.total_predictions,
    predictions: collected,
    offset,
    limit,
    has_more: false,
  });
}
