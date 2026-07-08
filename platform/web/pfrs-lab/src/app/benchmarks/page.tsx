import { listRunsAsync, loadRunSummary, objectiveFromMetadata } from '@/lib/data-loader';
import Card from '@/components/Card';
import BenchmarkLadder from './BenchmarkLadder';
import type { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Benchmarks',
  description: 'Algorithm leaderboard and benchmark results across CVRP, JSS, VRPTW, and NRP instances.',
};

export const dynamic = 'force-dynamic';

export interface BenchmarkRun {
  id: string;
  instance: string;
  problemType: string;
  mode: string;
  penalty: number; // objective value (penalty for NRP, distance for CVRP)
  runtimeMs: number;
  // Run parameters (from metadata).
  iterations: number;
  seed: number;
  temperature: number;
  customers: number;
  vehicles: number;
  capacity: number;
  timestamp: string;
}

export default async function BenchmarksPage() {
  const runs = await listRunsAsync();

  const benchmarkRuns: BenchmarkRun[] = [];

  const enriched = await Promise.all(
    runs.map(async (run) => {
      const metadata = run.metadata;
      if (!metadata && !run.manifestPenalty) return null;

      const meta = (metadata ?? {}) as unknown as Record<string, unknown>;
      const mode = String(meta.mode || 'unknown');
      let penalty = objectiveFromMetadata(meta, mode);
      let runtimeMs = Number(meta.runtimeMs || 0);

      if (penalty <= 0 && run.manifestPenalty && run.manifestPenalty > 0) {
        penalty = run.manifestPenalty;
      }

      if (penalty <= 0) {
        const summary = await loadRunSummary(run.id);
        penalty = summary.totalPenalty;
        runtimeMs = summary.totalDurationMs;
      }

      if (penalty <= 0) return null;

      let instance = String(meta.instance || 'unknown');
      if (instance.includes('/')) {
        const parts = instance.split('/');
        instance = parts[parts.length - 1].replace(/\.\w+$/, '');
      }

      return {
        id: run.id,
        instance,
        problemType: String(meta.problemType || 'nrp'),
        mode,
        penalty,
        runtimeMs,
        iterations: Number(meta.iterations || meta.iterationsPerWorker || 0),
        seed: Number(meta.seed || 0),
        temperature: Number(meta.initialTemperature || meta.temperature || 0),
        customers: Number(meta.customers || meta.dimension || 0),
        vehicles: Number(meta.vehicles || meta.bestVehicles || 0),
        capacity: Number(meta.capacity || 0),
        timestamp: run.timestamp || String(meta.timestamp || ''),
      } satisfies BenchmarkRun;
    })
  );

  for (const entry of enriched) {
    if (entry) benchmarkRuns.push(entry);
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
