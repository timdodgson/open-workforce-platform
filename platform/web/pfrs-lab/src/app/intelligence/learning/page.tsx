import { getStorageProvider } from '@/lib/storage';
import Card from '@/components/Card';
import type { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Continuous Learning',
  description: 'Training set growth, candidate quality, and promotion recommendations.',
};

export const dynamic = 'force-dynamic';

interface LearningState {
  new_samples_since_training: number;
  total_samples: number;
  last_trained_at?: string;
  last_training_accuracy: number;
  production_version: string;
  candidate_version?: string;
  candidate_accuracy?: number;
  recommendation: string;
  recommend_reason: string;
}

export default async function ContinuousLearningPage() {
  const storage = getStorageProvider();
  const content = await storage.readRootFile('policies/learning_state.json');

  let state: LearningState | null = null;
  if (content) {
    try { state = JSON.parse(content); } catch { /* */ }
  }

  if (!state) {
    return (
      <Card title="Continuous Learning">
        <div className="border-2 border-dashed border-slate-300 rounded-lg p-8 text-center text-slate-500">
          <p className="mb-2">No learning state available.</p>
          <p className="text-xs">Run experiments to accumulate training data.</p>
        </div>
      </Card>
    );
  }

  const recColors: Record<string, string> = {
    retrain: 'bg-amber-50 text-amber-700 border-amber-200',
    promote: 'bg-emerald-50 text-emerald-700 border-emerald-200',
    wait: 'bg-slate-50 text-slate-600 border-slate-200',
    none: 'bg-slate-50 text-slate-500 border-slate-200',
  };

  return (
    <div className="space-y-6">
      <Card title="Continuous Learning">
        <p className="text-sm text-slate-600 mb-4">
          Every completed run appends telemetry. When enough data accumulates,
          retraining is recommended. Promotion requires human approval.
        </p>

        <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
          <MetricBox label="Training Set Size" value={state.total_samples.toLocaleString()} />
          <MetricBox label="New Since Training" value={`${state.new_samples_since_training} / 50`} />
          <MetricBox label="Last Accuracy" value={state.last_training_accuracy > 0 ? `${(state.last_training_accuracy * 100).toFixed(1)}%` : '—'} />
          <MetricBox label="Production" value={state.production_version || 'rules'} />
        </div>

        {/* Recommendation */}
        <div className={`border rounded-lg p-4 ${recColors[state.recommendation] || recColors.none}`}>
          <div className="flex items-center gap-2 mb-1">
            <span className="text-sm font-semibold">Recommendation: {state.recommendation.toUpperCase()}</span>
          </div>
          <p className="text-xs opacity-80">{state.recommend_reason}</p>
        </div>
      </Card>

      {/* Candidate */}
      {state.candidate_version && (
        <Card title="Candidate Policy">
          <div className="border border-blue-200 bg-blue-50 rounded-lg p-4">
            <div className="flex items-baseline justify-between mb-2">
              <span className="text-sm font-semibold text-blue-800">v{state.candidate_version}</span>
              <span className="text-xs text-blue-600">{state.candidate_accuracy ? `${(state.candidate_accuracy * 100).toFixed(1)}% accuracy` : ''}</span>
            </div>
            <p className="text-xs text-blue-700">
              Trained but not yet promoted. Awaiting validation gate or manual approval.
            </p>
          </div>
        </Card>
      )}

      {/* Training History */}
      <Card title="Training Timeline">
        <div className="space-y-2">
          {state.last_trained_at && (
            <div className="flex items-center gap-3 text-xs">
              <span className="w-2 h-2 rounded-full bg-emerald-500" />
              <span className="text-slate-600">Last trained: {new Date(state.last_trained_at).toLocaleString()}</span>
              <span className="text-slate-400">accuracy: {(state.last_training_accuracy * 100).toFixed(1)}%</span>
            </div>
          )}
          <div className="flex items-center gap-3 text-xs">
            <span className="w-2 h-2 rounded-full bg-blue-500" />
            <span className="text-slate-600">{state.new_samples_since_training} new samples awaiting next training cycle</span>
          </div>
          <div className="flex items-center gap-3 text-xs">
            <span className="w-2 h-2 rounded-full bg-slate-300" />
            <span className="text-slate-500">Retrain threshold: 50 samples</span>
          </div>
        </div>
      </Card>
    </div>
  );
}

function MetricBox({ label, value }: { label: string; value: string }) {
  return (
    <div className="border border-slate-200 rounded-lg p-3">
      <p className="text-xs text-slate-500 uppercase tracking-wider">{label}</p>
      <p className="text-lg font-bold text-slate-800 mt-1">{value}</p>
    </div>
  );
}
