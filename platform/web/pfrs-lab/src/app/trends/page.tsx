import { listRunsAsync, loadRunSummary, loadTree, loadDiversity } from '@/lib/data-loader';
import Card from '@/components/Card';
import TrendAnalysis from './TrendAnalysis';
import { RunMetadata } from '@/lib/types';

export const dynamic = 'force-dynamic';

export interface TrendPoint {
  id: string;
  index: number;
  penalty: number;
  workers: number;
  candidates: number;
  durationMs: number;
  beamWidth: number;
  mode: string;
  instance: string;
  entropy: number;
  retainedPaths: number;
  nearDuplicates: number;
}

export default async function TrendsPage() {
  const runs = await listRunsAsync();

  if (runs.length < 2) {
    return (
      <Card title="Trend Analysis">
        <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500">
          <p className="mb-2">Need at least 2 runs for trend analysis.</p>
          <p className="text-xs">Run more experiments to see trends over time.</p>
        </div>
      </Card>
    );
  }

  const points: TrendPoint[] = [];
  for (let i = 0; i < runs.length; i++) {
    const run = runs[i];
    const [summary, tree, diversity] = await Promise.all([
      loadRunSummary(run.id),
      loadTree(run.id),
      loadDiversity(run.id),
    ]);
    const retained = tree.filter(t => t.retained);
    const nearDups = diversity.filter(d => d.nearDuplicate).length;

    // Compute entropy from final week retained paths.
    const maxWeek = Math.max(...tree.map(t => t.week), 0);
    const finalRetained = retained.filter(t => t.week === maxWeek);
    let entropy = 0;
    if (finalRetained.length > 1) {
      const parentMap = new Map<number, number>();
      for (const t of tree) parentMap.set(t.pathID, t.parentID);
      const families = new Map<number, number>();
      for (const t of finalRetained) {
        let cur = t.pathID; let iter = 0;
        while (parentMap.has(cur) && parentMap.get(cur)! >= 0 && iter < 100) { cur = parentMap.get(cur)!; iter++; }
        families.set(cur, (families.get(cur) || 0) + 1);
      }
      for (const count of families.values()) {
        const p = count / finalRetained.length;
        if (p > 0) entropy -= p * Math.log2(p);
      }
    }

    points.push({
      id: run.id,
      index: i,
      penalty: summary.totalPenalty,
      workers: summary.totalWorkers,
      candidates: summary.totalCandidates,
      durationMs: summary.totalDurationMs,
      beamWidth: summary.metadata?.beamWidth || 1,
      mode: summary.metadata?.mode || 'unknown',
      instance: summary.metadata?.instance || 'unknown',
      entropy,
      retainedPaths: retained.length,
      nearDuplicates: nearDups,
    });
  }

  return <TrendAnalysis points={points} />;
}
