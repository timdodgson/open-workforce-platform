import { listBenchmarkRunsAsync, objectiveFromMetadata } from '@/lib/data-loader';
import Card from '@/components/Card';
import SIComparison from './SIComparison';
import DomainBenchmarkCard from './DomainBenchmarkCard';
import { BENCHMARK_SUITES } from '@/lib/benchmark-suites';
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
  policyMode?: string;
}

// --- Known-optimal / best-known reference values (shared with DomainBenchmarkCard) ---
const KNOWN_OPTIMAL: Record<string, { value: number; source: string }> = {
  // CVRP (CVRPLIB best-known solutions)
  'A-n10-k2': { value: 204, source: 'CVRPLIB optimal' },
  'A-n32-k5': { value: 784, source: 'CVRPLIB optimal' },
  'A-n33-k5': { value: 661, source: 'CVRPLIB optimal' },
  'A-n45-k6': { value: 944, source: 'CVRPLIB optimal' },
  'A-n60-k9': { value: 1354, source: 'CVRPLIB optimal' },
  'A-n80-k10': { value: 1763, source: 'CVRPLIB optimal' },
  // NRP (ILP baseline)
  'n012w8': { value: 3020, source: 'ILP (HiGHS, 5hr)' },
  'n005w4': { value: 385, source: 'ILP baseline' },
  // JSS (Taillard/OR-Library optimal solutions)
  'ft06': { value: 55, source: 'Optimal (Fisher & Thompson)' },
  'ft10': { value: 930, source: 'Optimal (Fisher & Thompson)' },
  'la01': { value: 666, source: 'Optimal (Lawrence)' },
  // VRPTW (Solomon best-known solutions — distance only, ignoring vehicle count)
  'C101': { value: 828, source: 'Solomon BKS' },
  'R101': { value: 1645, source: 'Solomon BKS' },
  'RC101': { value: 1696, source: 'Solomon BKS' },
};

export default async function BenchmarksPage() {
  const runs = await listBenchmarkRunsAsync();

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
        policyMode: String(meta.policyMode || ''),
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

  return (
    <div className="space-y-6">
      <SIComparison runs={benchmarkRuns} />
      {BENCHMARK_SUITES.map((suite) => (
        <DomainBenchmarkCard key={suite.id} suite={suite} runs={benchmarkRuns} knownOptimal={KNOWN_OPTIMAL} />
      ))}
    </div>
  );
}
