'use client';

import Card from '@/components/Card';
import { WorkerPrediction } from './page';
import { FEATURE_LABELS } from './PredictionExplorer';

interface Props {
  worker: WorkerPrediction;
  onClose: () => void;
}

function featureLabel(name: string): string {
  return FEATURE_LABELS[name] || name;
}

export default function WorkerDetail({ worker, onClose }: Props) {
  const topContributions = Object.entries(worker.feature_contributions)
    .sort(([, a], [, b]) => b - a)
    .slice(0, 8);

  return (
    <Card title={`Worker #${worker.index} — Detail`}>
      <div className="flex justify-between items-start mb-4">
        <div className="text-xs text-gray-400">
          <span className="text-blue-400">{worker.instance}</span>
          {' · '}
          <span className="text-emerald-400">{worker.algorithm}</span>
          {' · '}
          W{worker.week}/D{worker.depth}
          {worker.run_id && <span className="text-gray-500"> · {worker.run_id}</span>}
        </div>
        <button
          onClick={onClose}
          className="text-gray-500 hover:text-gray-300 text-xs px-2 py-1 rounded border border-gray-700"
        >
          ✕ Close
        </button>
      </div>

      {/* Actual vs Predicted */}
      <div className="grid grid-cols-2 gap-4 mb-4">
        <div className="bg-gray-800 rounded-lg p-4">
          <h4 className="text-xs text-emerald-400 font-semibold uppercase mb-3">Actual Outcome</h4>
          <div className="space-y-2">
            <Row label="Improved" value={worker.actual.improved ? '✓ Yes' : '✗ No'} />
            <Row label="Global Best" value={worker.actual.produced_global_best ? '⭐ Yes' : '— No'} />
            <Row label="Improvement" value={worker.actual.improvement_amount.toLocaleString()} />
            <Row label="ROI" value={worker.actual.roi.toFixed(4)} />
          </div>
        </div>
        <div className="bg-gray-800 rounded-lg p-4">
          <h4 className="text-xs text-blue-400 font-semibold uppercase mb-3">Model Prediction</h4>
          <div className="space-y-2">
            <Row label="P(Improved)" value={`${(worker.predicted.p_improved * 100).toFixed(1)}%`} />
            <Row label="P(Global Best)" value={`${(worker.predicted.p_global_best * 100).toFixed(1)}%`} />
            <Row label="Expected Improvement" value={worker.predicted.expected_improvement.toFixed(0)} />
            <Row label="Expected ROI" value={worker.predicted.expected_roi.toFixed(4)} />
          </div>
        </div>
      </div>

      {/* Prediction Error */}
      <div className="grid grid-cols-2 gap-4 mb-4">
        <div className="bg-gray-800 rounded p-3">
          <div className="text-[9px] text-gray-500 uppercase">Improvement Error</div>
          <div className={`text-sm font-bold ${
            Math.abs(worker.error.improvement) > 500 ? 'text-red-400' : 'text-gray-300'
          }`}>
            {worker.error.improvement > 0 ? '+' : ''}{worker.error.improvement.toFixed(1)}
          </div>
        </div>
        <div className="bg-gray-800 rounded p-3">
          <div className="text-[9px] text-gray-500 uppercase">ROI Error</div>
          <div className={`text-sm font-bold ${
            Math.abs(worker.error.roi) > 1 ? 'text-red-400' : 'text-gray-300'
          }`}>
            {worker.error.roi > 0 ? '+' : ''}{worker.error.roi.toFixed(4)}
          </div>
        </div>
      </div>

      {/* What influenced this prediction? */}
      <div className="mb-4">
        <h4 className="text-xs text-blue-400 font-semibold uppercase mb-2">
          What influenced this prediction?
        </h4>
        <div className="bg-blue-900/20 border border-blue-800 rounded-lg p-4 mb-3">
          <p className="text-xs text-gray-200 whitespace-pre-line">{worker.explanation}</p>
        </div>

        {topContributions.length > 0 && (
          <table className="w-full text-xs">
            <thead>
              <tr className="text-gray-500 uppercase">
                <th className="text-left p-1.5">Feature</th>
                <th className="text-right p-1.5">Value</th>
                <th className="text-right p-1.5">Contribution</th>
                <th className="text-left p-1.5 w-1/3">Weight</th>
              </tr>
            </thead>
            <tbody>
              {topContributions.map(([feature, contribution]) => (
                <tr key={feature} className="border-t border-gray-800">
                  <td className="p-1.5 text-blue-400">{featureLabel(feature)}</td>
                  <td className="text-right p-1.5 text-gray-300 font-mono">
                    {worker.feature_values[feature]?.toFixed(2) ?? '—'}
                  </td>
                  <td className="text-right p-1.5 text-emerald-400">
                    {(contribution * 100).toFixed(0)}%
                  </td>
                  <td className="p-1.5">
                    <div className="w-full bg-gray-700 rounded-full h-1.5">
                      <div
                        className="h-1.5 rounded-full bg-emerald-500"
                        style={{ width: `${contribution * 100}%` }}
                      />
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {/* Decision Path */}
      {worker.decision_path.length > 0 && (
        <div>
          <h4 className="text-xs text-blue-400 font-semibold uppercase mb-2">
            Decision Path (Global Best Model)
          </h4>
          <div className="bg-gray-800 rounded-lg p-3 font-mono text-[10px] space-y-1">
            {worker.decision_path.map((step, i) => (
              <div key={i} className="flex gap-2" style={{ paddingLeft: `${i * 12}px` }}>
                <span className="text-gray-500">{i + 1}.</span>
                <span className="text-blue-400">{featureLabel(step.feature)}</span>
                <span className="text-gray-500">{step.condition}</span>
                <span className="text-amber-400">{step.threshold}</span>
                <span className="text-gray-600">|</span>
                <span className="text-gray-300">value={step.value}</span>
              </div>
            ))}
          </div>
        </div>
      )}
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
