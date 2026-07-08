import { listRunsAsync } from '@/lib/data-loader';
import Card from '@/components/Card';
import TrendAnalysis from './TrendAnalysis';
import type { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Trends',
  description: 'Research history and performance trends over time across all experiments.',
};

export const dynamic = 'force-dynamic';

export interface TrendPoint {
  id: string;
  index: number;
  objective: number;
  runtime: number;
  candidates: number;
  domain: string;
  instance: string;
  mode: string;
  seed: number;
}

// Best-known / optimal references for gap calculation.
const BEST_KNOWN: Record<string, number> = {
  'A-n32-k5': 784,
  'A-n45-k6': 944,
  'A-n60-k9': 1354,
  'A-n80-k10': 1763,
  'C101': 828,
  'la01': 666,
  'ft06': 55,
  'ft10': 930,
};

export default async function TrendsPage() {
  const runs = await listRunsAsync();

  if (runs.length < 3) {
    return (
      <Card title="Research History">
        <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500">
          <p className="mb-2">Need at least 3 runs to show trends.</p>
          <p className="text-xs">Run experiments over time to track platform improvement.</p>
        </div>
      </Card>
    );
  }

  // Extract lightweight data from run.json metadata.
  const points: TrendPoint[] = [];
  for (let i = 0; i < runs.length; i++) {
    const meta = runs[i].metadata as unknown as Record<string, unknown>;
    if (!meta) continue;

    const objective = Number(meta.bestObjective || meta.bestDistance || meta.bestMakespan || meta.totalPenalty || 0);
    if (objective <= 0) continue;

    const domain = String(meta.problemType || 'nrp');
    const instance = String(meta.instance || '');
    const mode = String(meta.mode || 'unknown');
    const runtime = Number(meta.runtimeMs || 0);
    const candidates = Number(meta.iterations || 0);
    const seed = Number(meta.seed || 0);

    points.push({
      id: runs[i].id,
      index: i,
      objective,
      runtime,
      candidates,
      domain,
      instance,
      mode,
      seed,
    });
  }

  if (points.length < 3) {
    return (
      <Card title="Research History">
        <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500">
          <p>Not enough runs with valid objectives for trend analysis.</p>
        </div>
      </Card>
    );
  }

  return <TrendAnalysis points={points} bestKnown={BEST_KNOWN} />;
}
