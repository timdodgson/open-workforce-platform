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
    const instance = String(meta.instance || metadata.instance || 'unknown');
    const problemType = String(meta.problemType || 'nrp');
    const mode = String(meta.mode || metadata.mode || 'unknown');

    // Get the objective value.
    let penalty = summary.totalPenalty;
    // For CVRP with bestDistance in metadata, prefer that.
    if (meta.bestDistance && Number(meta.bestDistance) > 0) {
      penalty = Number(meta.bestDistance);
    }
    // For ILP, use the objective from metadata.
    if (mode === 'ilp' && meta.objective && Number(meta.objective) > 0) {
      penalty = Number(meta.objective);
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
