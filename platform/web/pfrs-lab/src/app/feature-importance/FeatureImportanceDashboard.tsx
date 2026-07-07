'use client';

import { useMemo } from 'react';
import Card from '@/components/Card';
import { WorkerModel, ModelResult } from './types';
import FeatureBarChart from './FeatureBarChart';

interface Props {
  model: WorkerModel;
}

// Human-readable labels for feature names.
const FEATURE_LABELS: Record<string, string> = {
  distance_from_best: 'Parent Gap',
  parent_objective: 'Parent Objective',
  global_best: 'Global Best',
  depth: 'Depth',
  week: 'Week',
  beam_rank: 'Beam Rank',
  beam_score: 'Beam Score',
  entropy: 'Entropy',
  diversity: 'Diversity',
  beam_health: 'Beam Health',
  temperature: 'Temperature',
  iterations_alloc: 'Iteration Budget',
  plateau_length: 'Plateau Length',
  recent_improv_rate: 'Recent Improvement Rate',
  worker_count: 'Worker Count',
  active_families: 'Active Families',
};

function featureLabel(name: string): string {
  return FEATURE_LABELS[name] || name;
}

export default function FeatureImportanceDashboard({ model }: Props) {
  // Aggregate importance across all models (weighted by model utility).
  const aggregatedImportance = useMemo(() => {
    const allModels = Object.values(model.models);
    const featureScores = new Map<string, number>();

    for (const m of allModels) {
      for (const [feature, importance] of Object.entries(m.feature_importance)) {
        featureScores.set(feature, (featureScores.get(feature) || 0) + importance);
      }
    }

    // Normalise so total = 1.
    const total = Array.from(featureScores.values()).reduce((a, b) => a + b, 0);
    const normalised = new Map<string, number>();
    for (const [k, v] of featureScores) {
      normalised.set(k, total > 0 ? v / total : 0);
    }

    return Array.from(normalised.entries())
      .map(([feature, importance]) => ({ feature, importance, label: featureLabel(feature) }))
      .sort((a, b) => b.importance - a.importance);
  }, [model]);

  // Per-model importance (for comparison).
  const perModelImportance = useMemo(() => {
    return Object.entries(model.models).map(([name, m]) => ({
      name,
      features: Object.entries(m.feature_importance)
        .map(([feature, importance]) => ({ feature, importance, label: featureLabel(feature) }))
        .sort((a, b) => b.importance - a.importance)
        .filter(f => f.importance > 0),
    }));
  }, [model]);

  // Auto-generated explanation.
  const explanation = useMemo(() => {
    const top3 = aggregatedImportance.slice(0, 3).filter(f => f.importance > 0.05);
    if (top3.length === 0) return 'Insufficient data to determine feature importance.';

    const parts = top3.map(f => `**${f.label}** (${(f.importance * 100).toFixed(0)}%)`);

    let sentence: string;
    if (parts.length === 1) {
      sentence = `The model relies almost entirely on ${parts[0]}.`;
    } else if (parts.length === 2) {
      sentence = `The model relies heavily on ${parts[0]} and ${parts[1]}.`;
    } else {
      sentence = `The model relies heavily on ${parts[0]}, ${parts[1]}, and ${parts[2]}.`;
    }

    // Add context about what these features mean.
    const contextParts: string[] = [];
    for (const f of top3) {
      switch (f.feature) {
        case 'distance_from_best':
          contextParts.push('Workers closer to the global best are more likely to find improvements.');
          break;
        case 'parent_objective':
          contextParts.push('The parent solution quality strongly predicts how much improvement is possible.');
          break;
        case 'iterations_alloc':
          contextParts.push('Iteration budget determines how much search effort each worker gets.');
          break;
        case 'beam_health':
          contextParts.push('Beam health reflects search diversity — healthier beams discover more.');
          break;
        case 'entropy':
          contextParts.push('Low entropy suggests the search is converging — different strategies may help.');
          break;
        case 'depth':
          contextParts.push('Deeper workers have had more ancestry to build on, but diminishing returns appear.');
          break;
        case 'temperature':
          contextParts.push('Temperature controls exploration vs exploitation in simulated annealing.');
          break;
        case 'week':
          contextParts.push('Later weeks have tighter constraints, making improvement harder.');
          break;
      }
    }

    return sentence + (contextParts.length > 0 ? ' ' + contextParts[0] : '');
  }, [aggregatedImportance]);

  // Feature correlations: which features are important for which predictions.
  const correlationMatrix = useMemo(() => {
    return model.features.map(feature => {
      const improved = model.models.improved.feature_importance[feature] || 0;
      const produced_global_best = model.models.produced_global_best.feature_importance[feature] || 0;
      const improvement_amount = model.models.improvement_amount.feature_importance[feature] || 0;
      const roi = model.models.roi.feature_importance[feature] || 0;
      const total = improved + produced_global_best + improvement_amount + roi;
      return { feature, label: featureLabel(feature), improved, produced_global_best, improvement_amount, roi, total };
    }).sort((a, b) => b.total - a.total);
  }, [model]);

  return (
    <div className="space-y-4">
      {/* Header */}
      <Card title="Feature Importance — Worker Value Model">
        <p className="text-xs text-gray-500 mb-4">
          Which features the model uses to predict worker value. Higher importance means the feature
          has more influence on predictions. Features are measured at spawn time — before the worker runs.
        </p>
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 mb-4">
          <Stat label="Model Version" value={model.version} colour="blue" />
          <Stat label="Training Samples" value={model.training_samples.toLocaleString()} colour="emerald" />
          <Stat label="Features Used" value={model.features.length} colour="blue" />
          <Stat label="Models Trained" value={Object.keys(model.models).length} colour="emerald" />
        </div>

        {/* Auto-explanation */}
        <div className="bg-blue-900/20 border border-blue-800 rounded-lg p-4 mt-3">
          <p className="text-xs text-blue-300 uppercase font-semibold mb-1">Model Insight</p>
          <p className="text-sm text-gray-200">{explanation.replace(/\*\*/g, '')}</p>
        </div>
      </Card>

      {/* Aggregate Importance Bar Chart */}
      <Card title="Overall Feature Importance (Aggregated)">
        <p className="text-xs text-gray-500 mb-4">
          Combined importance across all four prediction targets. Shows which features matter most overall.
        </p>
        <FeatureBarChart
          data={aggregatedImportance.filter(f => f.importance > 0)}
          colour="#34d399"
        />
      </Card>

      {/* Importance Ranking Table */}
      <Card title="Feature Importance Ranking">
        <div className="overflow-x-auto">
          <table className="w-full text-xs">
            <thead>
              <tr className="text-gray-500 uppercase">
                <th className="text-left p-2">#</th>
                <th className="text-left p-2">Feature</th>
                <th className="text-right p-2">Importance</th>
                <th className="text-left p-2 w-1/2">Distribution</th>
              </tr>
            </thead>
            <tbody>
              {aggregatedImportance.map((f, i) => (
                <tr key={f.feature} className={`border-t border-gray-800 ${i < 3 ? 'bg-emerald-900/10' : ''}`}>
                  <td className="p-2 text-gray-500">{i + 1}</td>
                  <td className="p-2">
                    <span className="text-blue-400 font-medium">{f.label}</span>
                    <span className="text-gray-600 ml-2 text-[10px] font-mono">{f.feature}</span>
                  </td>
                  <td className="text-right p-2 text-emerald-400 font-semibold">
                    {(f.importance * 100).toFixed(1)}%
                  </td>
                  <td className="p-2">
                    <div className="w-full bg-gray-800 rounded-full h-2.5">
                      <div
                        className={`h-2.5 rounded-full ${i < 3 ? 'bg-emerald-500' : 'bg-gray-600'}`}
                        style={{ width: `${Math.max(f.importance * 100, 0.5)}%` }}
                      />
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </Card>

      {/* Per-Model Importance */}
      <Card title="Feature Importance by Prediction Target">
        <p className="text-xs text-gray-500 mb-4">
          Different features matter for different predictions. A feature important for &quot;improvement amount&quot;
          may not matter for &quot;global best&quot; prediction.
        </p>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {perModelImportance.map(m => (
            <div key={m.name} className="bg-gray-800 rounded-lg p-4">
              <h4 className="text-xs text-blue-400 font-semibold uppercase mb-3">
                {formatModelName(m.name)}
              </h4>
              {m.features.length === 0 ? (
                <p className="text-xs text-gray-500 italic">No distinguishing features (trivial prediction)</p>
              ) : (
                <div className="space-y-1.5">
                  {m.features.slice(0, 5).map((f, i) => (
                    <div key={f.feature} className="flex items-center gap-2">
                      <span className="text-[10px] text-gray-500 w-4">{i + 1}.</span>
                      <span className="text-[11px] text-gray-300 w-28 truncate">{f.label}</span>
                      <div className="flex-1 bg-gray-700 rounded-full h-1.5">
                        <div
                          className="h-1.5 rounded-full bg-amber-500"
                          style={{ width: `${f.importance * 100}%` }}
                        />
                      </div>
                      <span className="text-[10px] text-gray-500 w-10 text-right">
                        {(f.importance * 100).toFixed(0)}%
                      </span>
                    </div>
                  ))}
                </div>
              )}
            </div>
          ))}
        </div>
      </Card>

      {/* Correlation Matrix */}
      <Card title="Feature × Target Correlation">
        <p className="text-xs text-gray-500 mb-4">
          Heatmap showing how much each feature contributes to each prediction target.
          Darker cells indicate higher importance for that specific target.
        </p>
        <div className="overflow-x-auto">
          <table className="w-full text-[10px]">
            <thead>
              <tr className="text-gray-500 uppercase">
                <th className="text-left p-1.5">Feature</th>
                <th className="text-center p-1.5">Improved</th>
                <th className="text-center p-1.5">Global Best</th>
                <th className="text-center p-1.5">Improvement Δ</th>
                <th className="text-center p-1.5">ROI</th>
              </tr>
            </thead>
            <tbody>
              {correlationMatrix.slice(0, 12).map(row => (
                <tr key={row.feature} className="border-t border-gray-800">
                  <td className="p-1.5 text-blue-400">{row.label}</td>
                  <HeatCell value={row.improved} />
                  <HeatCell value={row.produced_global_best} />
                  <HeatCell value={row.improvement_amount} />
                  <HeatCell value={row.roi} />
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </Card>

      {/* Model Performance Summary */}
      <Card title="Model Performance">
        <p className="text-xs text-gray-500 mb-4">
          How well each prediction target performs. Strong metrics mean the model has learned
          reliable patterns from the training data.
        </p>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {Object.entries(model.models).map(([name, m]) => (
            <ModelCard key={name} name={name} model={m} />
          ))}
        </div>
      </Card>

      {/* Data Summary */}
      <Card title="Training Data Summary">
        <div className="grid grid-cols-2 sm:grid-cols-5 gap-3">
          <Stat label="Total Records" value={model.data_summary.total_records.toLocaleString()} colour="blue" />
          <Stat label="Improvement Rate" value={`${(model.data_summary.improvement_rate * 100).toFixed(0)}%`} colour="emerald" />
          <Stat label="Global Best Rate" value={`${(model.data_summary.global_best_rate * 100).toFixed(1)}%`} colour="amber" />
          <Stat label="Mean Improvement" value={model.data_summary.mean_improvement.toFixed(0)} colour="emerald" />
          <Stat label="Mean ROI" value={model.data_summary.mean_roi.toFixed(2)} colour="blue" />
        </div>
      </Card>
    </div>
  );
}

function HeatCell({ value }: { value: number }) {
  const intensity = Math.min(value * 100, 100);
  const bg = intensity > 50
    ? `rgba(52, 211, 153, ${intensity / 100})`  // emerald
    : intensity > 10
    ? `rgba(251, 191, 36, ${intensity / 100})`   // amber
    : 'transparent';

  return (
    <td className="p-1.5 text-center" style={{ backgroundColor: bg }}>
      <span className={intensity > 30 ? 'text-gray-900 font-semibold' : 'text-gray-500'}>
        {value > 0 ? `${(value * 100).toFixed(0)}%` : '—'}
      </span>
    </td>
  );
}

function ModelCard({ name, model }: { name: string; model: ModelResult }) {
  const isClassifier = model.type === 'classifier';

  return (
    <div className="bg-gray-800 rounded-lg p-4">
      <h4 className="text-xs text-blue-400 font-semibold uppercase mb-2">
        {formatModelName(name)}
      </h4>
      <p className="text-[10px] text-gray-500 mb-2">
        {isClassifier ? 'Classification' : 'Regression'} • max_depth={model.max_depth} • n={model.n_train + model.n_test}
      </p>
      <div className="space-y-1">
        {Object.entries(model.metrics).map(([metric, value]) => (
          <div key={metric} className="flex justify-between text-xs">
            <span className="text-gray-500">{formatMetricName(metric)}</span>
            <span className={metricColour(metric, value)}>
              {formatMetricValue(metric, value)}
            </span>
          </div>
        ))}
      </div>
      {model.confusion_matrix && (
        <div className="mt-3 pt-2 border-t border-gray-700">
          <p className="text-[9px] text-gray-500 uppercase mb-1">Confusion Matrix</p>
          <div className="grid grid-cols-2 gap-1 text-[10px] text-center">
            <div className="bg-emerald-900/30 rounded p-1">TP: {model.confusion_matrix.tp}</div>
            <div className="bg-red-900/30 rounded p-1">FP: {model.confusion_matrix.fp}</div>
            <div className="bg-amber-900/30 rounded p-1">FN: {model.confusion_matrix.fn}</div>
            <div className="bg-emerald-900/30 rounded p-1">TN: {model.confusion_matrix.tn}</div>
          </div>
        </div>
      )}
    </div>
  );
}

function Stat({ label, value, colour }: { label: string; value: string | number; colour: string }) {
  const colourClass = colour === 'blue' ? 'text-blue-400' : colour === 'emerald' ? 'text-emerald-400' : 'text-amber-400';
  return (
    <div className="bg-gray-800 rounded p-3">
      <div className="text-[9px] text-gray-500 uppercase">{label}</div>
      <div className={`text-lg font-bold ${colourClass}`}>{value}</div>
    </div>
  );
}

function formatModelName(name: string): string {
  const names: Record<string, string> = {
    improved: 'Will Improve?',
    produced_global_best: 'Will Find Global Best?',
    improvement_amount: 'Expected Improvement',
    roi: 'Expected ROI',
  };
  return names[name] || name;
}

function formatMetricName(metric: string): string {
  const names: Record<string, string> = {
    accuracy: 'Accuracy',
    precision: 'Precision',
    recall: 'Recall',
    f1: 'F1 Score',
    roc_auc: 'ROC-AUC',
    mae: 'MAE',
    mse: 'MSE',
    rmse: 'RMSE',
    r2: 'R²',
  };
  return names[metric] || metric;
}

function formatMetricValue(metric: string, value: number): string {
  if (metric === 'mae' || metric === 'mse' || metric === 'rmse') {
    return value.toFixed(2);
  }
  if (metric === 'r2') {
    return value.toFixed(3);
  }
  return `${(value * 100).toFixed(1)}%`;
}

function metricColour(metric: string, value: number): string {
  if (metric === 'r2') {
    return value >= 0.9 ? 'text-emerald-400' : value >= 0.7 ? 'text-amber-400' : 'text-red-400';
  }
  if (['accuracy', 'precision', 'recall', 'f1', 'roc_auc'].includes(metric)) {
    return value >= 0.8 ? 'text-emerald-400' : value >= 0.5 ? 'text-amber-400' : 'text-red-400';
  }
  return 'text-gray-300';
}
