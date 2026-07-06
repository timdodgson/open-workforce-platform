'use client';

import Card from '@/components/Card';

interface GroupStats {
  count: number;
  successRate: number;
  avgImprovement: number;
  avgRuntime: number;
}

interface Props {
  run: GroupStats;
  skip: GroupStats;
}

export default function ROIComparison({ run, skip }: Props) {
  const runROI = run.avgRuntime > 0 ? (run.avgImprovement / run.avgRuntime) * 1000 : 0;
  const skipROI = skip.avgRuntime > 0 ? (skip.avgImprovement / skip.avgRuntime) * 1000 : 0;

  return (
    <Card title="Run vs Skip — ROI Comparison">
      <p className="text-xs text-gray-500 mb-4">
        Compares actual outcomes of workers the engine recommended running vs skipping.
        If &quot;skip&quot; workers have low success rate and low improvement, the engine is working well.
      </p>
      <div className="grid grid-cols-2 gap-4">
        {/* Run group */}
        <div className="border border-emerald-800 rounded-lg p-4">
          <h4 className="text-xs text-emerald-400 font-semibold uppercase mb-3">Recommended: Run</h4>
          <div className="space-y-2">
            <Row label="Workers" value={run.count.toString()} />
            <Row label="Success Rate" value={`${run.successRate.toFixed(1)}%`} />
            <Row label="Avg Improvement" value={run.avgImprovement.toFixed(0)} />
            <Row label="Avg Runtime" value={`${run.avgRuntime.toFixed(0)}ms`} />
            <Row label="ROI (Δ/s)" value={runROI.toFixed(2)} />
          </div>
        </div>

        {/* Skip group */}
        <div className="border border-amber-800 rounded-lg p-4">
          <h4 className="text-xs text-amber-400 font-semibold uppercase mb-3">Recommended: Skip</h4>
          <div className="space-y-2">
            <Row label="Workers" value={skip.count.toString()} />
            <Row label="Success Rate" value={`${skip.successRate.toFixed(1)}%`} />
            <Row label="Avg Improvement" value={skip.avgImprovement.toFixed(0)} />
            <Row label="Avg Runtime" value={`${skip.avgRuntime.toFixed(0)}ms`} />
            <Row label="ROI (Δ/s)" value={skipROI.toFixed(2)} />
          </div>
        </div>
      </div>

      {/* Interpretation */}
      <div className="mt-4 border-t border-gray-800 pt-3">
        <p className="text-xs text-gray-500 uppercase mb-1">Interpretation</p>
        {skip.count === 0 ? (
          <p className="text-sm text-gray-400">No skip recommendations — engine is conservative (all workers run).</p>
        ) : run.successRate > skip.successRate ? (
          <p className="text-sm text-emerald-400">
            ✓ Run-recommended workers improve more often ({run.successRate.toFixed(0)}% vs {skip.successRate.toFixed(0)}%) — engine discriminates correctly.
          </p>
        ) : (
          <p className="text-sm text-red-400">
            ✗ Skip-recommended workers improve more often ({skip.successRate.toFixed(0)}% vs {run.successRate.toFixed(0)}%) — engine is inverted.
          </p>
        )}
      </div>
    </Card>
  );
}

function Row({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex justify-between text-xs">
      <span className="text-gray-500">{label}</span>
      <span className="text-gray-200 font-medium">{value}</span>
    </div>
  );
}
