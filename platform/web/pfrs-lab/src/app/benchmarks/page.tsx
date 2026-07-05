import { listRunsAsync, loadRunSummary, loadRunMetadata } from '@/lib/data-loader';
import Card from '@/components/Card';
import BenchmarkLadder from './BenchmarkLadder';
import { RunMetadata, RunSummary } from '@/lib/types';

export const dynamic = 'force-dynamic';

export interface BenchmarkRun {
  id: string;
  instance: string;
  problemType: string;
  mode: string;
  penalty: number; // objective value (penalty for NRP, distance for CVRP)
  runtimeMs: number;
}

export default async function BenchmarksPage() {
  const runs = await listRunsAsync();

  const benchmarkRuns: BenchmarkRun[] = [];

  for (const run of runs) {
    const [metadata, summary] = await Promise.all([
      loadRunMetadata(run.id),
      loadRunSummary(run.id),
    ]);

    if (!metadata) continue;

    const meta = metadata as unknown as Record<string, unknown>;
    let instance = String(meta.instance || metadata.instance || 'unknown');
    // Clean up instance paths (e.g. "internal/.../ft06.txt" → "ft06").
    if (instance.includes('/')) {
      const parts = instance.split('/');
      instance = parts[parts.length - 1].replace(/\.\w+$/, '');
    }
    const problemType = String(meta.problemType || 'nrp');
    const mode = String(meta.mode || metadata.mode || 'unknown');

    // Get the objective value. Prefer metadata fields over summary for non-NRP runs.
    let penalty = 0;
    if (meta.bestDistance && Number(meta.bestDistance) > 0) {
      penalty = Number(meta.bestDistance);
    } else if (meta.bestMakespan && Number(meta.bestMakespan) > 0) {
      penalty = Number(meta.bestMakespan);
    } else if (mode === 'ilp' && meta.objective && Number(meta.objective) > 0) {
      penalty = Number(meta.objective);
    } else {
      penalty = summary.totalPenalty;
    }

    if (penalty <= 0) continue; // skip runs without valid results

    benchmarkRuns.push({
      id: run.id,
      instance,
      problemType,
      mode,
      penalty,
      runtimeMs: summary.totalDurationMs,
    });
  }

  if (benchmarkRuns.length === 0) {
    return (
      <Card title="Benchmark Ladder">
        <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500">
          <p className="mb-2">No benchmark data available.</p>
          <p className="text-xs">Run experiments with --run-label to populate the ladder.</p>
        </div>
      </Card>
    );
  }

  return <BenchmarkLadder runs={benchmarkRuns} />;
}
