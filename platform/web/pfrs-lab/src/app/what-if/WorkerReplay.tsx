'use client';

import { WhatIfPrediction } from './types';

interface SimulatedWorker {
  prediction: WhatIfPrediction;
  action: 'run' | 'skip' | 'reduce' | 'increase';
  budgetMultiplier: number;
  reason: string;
}

interface Props {
  worker: SimulatedWorker;
}

export default function WorkerReplay({ worker }: Props) {
  const p = worker.prediction;
  const wasCorrect = (worker.action === 'skip' && !p.actual.improved) ||
    (worker.action !== 'skip' && p.actual.improved);

  return (
    <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
      <div className="flex items-start justify-between mb-3">
        <div>
          <h4 className="text-xs text-blue-400 font-semibold">
            Worker #{p.index} — {p.instance}
          </h4>
          <p className="text-[10px] text-gray-500">
            {p.algorithm} · W{p.week}/D{p.depth} · seed {p.seed}
          </p>
        </div>
        <div className={`text-xs px-2 py-1 rounded ${wasCorrect ? 'bg-emerald-900/50 text-emerald-300' : 'bg-red-900/50 text-red-300'}`}>
          {wasCorrect ? '✓ Correct' : '✗ Incorrect'}
        </div>
      </div>

      <div className="grid grid-cols-3 gap-3 text-xs">
        {/* Original */}
        <div className="bg-gray-900 rounded p-3">
          <p className="text-[9px] text-gray-500 uppercase mb-2">Original Decision</p>
          <Row label="Action" value="Run (all workers ran)" />
          <Row label="Budget" value="100%" />
        </div>

        {/* Model Recommendation */}
        <div className="bg-gray-900 rounded p-3">
          <p className="text-[9px] text-blue-400 uppercase mb-2">Model Recommendation</p>
          <Row label="Action" value={worker.action} />
          <Row label="Budget" value={`${(worker.budgetMultiplier * 100).toFixed(0)}%`} />
          <Row label="Reason" value={worker.reason} />
          <Row label="P(Improved)" value={`${(p.predicted.p_improved * 100).toFixed(0)}%`} />
          <Row label="P(Global Best)" value={`${(p.predicted.p_global_best * 100).toFixed(0)}%`} />
        </div>

        {/* Actual Outcome */}
        <div className="bg-gray-900 rounded p-3">
          <p className="text-[9px] text-emerald-400 uppercase mb-2">Actual Outcome</p>
          <Row label="Improved" value={p.actual.improved ? '✓ Yes' : '✗ No'} />
          <Row label="Global Best" value={p.actual.produced_global_best ? '⭐ Yes' : '— No'} />
          <Row label="Improvement" value={p.actual.improvement_amount.toLocaleString()} />
          <Row label="ROI" value={p.actual.roi.toFixed(4)} />
        </div>
      </div>

      {/* Simulated outcome narrative */}
      <div className="mt-3 p-3 rounded bg-blue-900/20 border border-blue-800">
        <p className="text-[10px] text-gray-300">
          {worker.action === 'skip' && p.actual.improved && (
            <span className="text-red-400">
              ⚠ This worker would have been skipped but actually improved by {p.actual.improvement_amount.toLocaleString()}.
              {p.actual.produced_global_best && ' It also found the global best — critical miss.'}
            </span>
          )}
          {worker.action === 'skip' && !p.actual.improved && (
            <span className="text-emerald-400">
              ✓ This worker would have been correctly skipped — it did not improve.
              CPU saved with no loss of quality.
            </span>
          )}
          {worker.action === 'run' && (
            <span className="text-gray-400">
              Worker runs as normal. No change from baseline.
            </span>
          )}
          {worker.action === 'reduce' && (
            <span className="text-amber-400">
              Budget reduced to {(worker.budgetMultiplier * 100).toFixed(0)}%.
              {p.actual.improved
                ? ` Worker improved by ${p.actual.improvement_amount.toLocaleString()} — may have found less with reduced budget.`
                : ' Worker did not improve — reduced budget is appropriate.'}
            </span>
          )}
          {worker.action === 'increase' && (
            <span className="text-emerald-400">
              Budget increased to {(worker.budgetMultiplier * 100).toFixed(0)}%.
              {p.actual.produced_global_best
                ? ' Excellent — this worker found the global best. Extra budget was warranted.'
                : p.actual.improved
                ? ` Worker improved by ${p.actual.improvement_amount.toLocaleString()}.`
                : ' Worker did not improve — extra budget was wasted.'}
            </span>
          )}
        </p>
      </div>
    </div>
  );
}

function Row({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex justify-between text-[10px] mb-0.5">
      <span className="text-gray-500">{label}</span>
      <span className="text-gray-300 text-right max-w-[120px] truncate">{value}</span>
    </div>
  );
}
