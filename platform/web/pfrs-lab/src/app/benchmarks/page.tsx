import { listRunsAsync, loadRunSummary, loadRunMetadata } from '@/lib/data-loader';
import Card from '@/components/Card';
import BenchmarkLadder from './BenchmarkLadder';
import { RunMetadata, RunSummary } from '@/lib/types';
import type { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Benchmarks',
  description: 'Algorithm leaderboard and benchmark results across CVRP, JSS, VRPTW, and NRP instances.',
};

export const revalidate = 60;

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

  // Load metadata and summaries in parallel (not sequentially).
  const enriched = await Promise.all(
    runs.map(async (run) => {
      const [metadata, summary] = await Promise.all([
        loadRunMetadata(run.id),
        loadRunSummary(run.id),
      ]);
      return { run, metadata, summary };
    })
  );

  for (const { run, metadata, summary } of enriched) {
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

    // Get the objective value. Prefer metadata fields over summary.
    let penalty = 0;
    if (meta.bestObjective && Number(meta.bestObjective) > 0) {
      penalty = Number(meta.bestObjective);
    } else if (meta.bestDistance && Number(meta.bestDistance) > 0) {
      penalty = Number(meta.bestDistance);
    } else if (meta.bestMakespan && Number(meta.bestMakespan) > 0) {
      penalty = Number(meta.bestMakespan);
    } else if (meta.totalPenalty && Number(meta.totalPenalty) > 0) {
      penalty = Number(meta.totalPenalty);
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
      runtimeMs: Number(meta.runtimeMs || summary.totalDurationMs || 0),
      iterations: Number(meta.iterations || meta.iterationsPerWorker || 0),
      seed: Number(meta.seed || 0),
      temperature: Number(meta.initialTemperature || meta.temperature || 0),
      customers: Number(meta.customers || meta.dimension || 0),
      vehicles: Number(meta.vehicles || meta.bestVehicles || 0),
      capacity: Number(meta.capacity || 0),
      timestamp: String(run.metadata ? (run.metadata as unknown as Record<string, unknown>).timestamp || '' : ''),
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
