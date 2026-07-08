'use client';

import Card from '@/components/Card';
import type { ContinuousLearningState } from '@/lib/types/intelligence';

export default function ContinuousLearningTab({ state }: { state: ContinuousLearningState | null }) {
  if (!state) {
    return (
      <Card title="Continuous Learning">
        <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500 text-xs">
          No learning state available. Run experiments to accumulate training data.
        </div>
      </Card>
    );
  }

  const recColors: Record<string, string> = {
    retrain: 'border-amber-500/40 bg-amber-950/30 text-amber-300',
    promote: 'border-emerald-500/40 bg-emerald-950/30 text-emerald-300',
    wait: 'border-gray-600 bg-gray-800/50 text-gray-400',
    none: 'border-gray-700 bg-gray-800/30 text-gray-500',
  };

  return (
    <div className="space-y-4">
      <Card title="Continuous Learning">
        <p className="text-xs text-gray-400 mb-4">
          Every completed run appends telemetry. When enough data accumulates, retraining is recommended.
          Promotion requires human approval.
        </p>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-3 mb-4">
          <Metric label="Training Set" value={state.total_samples.toLocaleString()} />
          <Metric label="New Since Training" value={`${state.new_samples_since_training} / 50`} />
          <Metric label="Last Accuracy" value={state.last_training_accuracy > 0 ? `${(state.last_training_accuracy * 100).toFixed(1)}%` : '—'} />
          <Metric label="Production" value={state.production_version || 'rules'} />
        </div>
        <div className={`border rounded-lg p-4 ${recColors[state.recommendation] || recColors.none}`}>
          <p className="text-sm font-semibold mb-1">Recommendation: {state.recommendation.toUpperCase()}</p>
          <p className="text-xs opacity-80">{state.recommend_reason}</p>
        </div>
      </Card>
      {state.candidate_version && (
        <Card title="Candidate Policy">
          <div className="border border-blue-500/30 bg-blue-950/20 rounded-lg p-4">
            <div className="flex justify-between mb-2">
              <span className="text-sm font-semibold text-blue-300">v{state.candidate_version}</span>
              {state.candidate_accuracy != null && state.candidate_accuracy > 0 && (
                <span className="text-xs text-blue-400">{(state.candidate_accuracy * 100).toFixed(1)}% accuracy</span>
              )}
            </div>
            {state.last_trained_at && (
              <p className="text-[10px] text-gray-500">Last trained: {state.last_trained_at}</p>
            )}
          </div>
        </Card>
      )}
    </div>
  );
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div className="bg-gray-800 rounded p-3">
      <p className="text-[9px] text-gray-500 uppercase">{label}</p>
      <p className="text-lg font-bold text-gray-200 mt-1">{value}</p>
    </div>
  );
}
