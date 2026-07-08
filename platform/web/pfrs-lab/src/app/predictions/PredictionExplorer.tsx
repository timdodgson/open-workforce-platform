'use client';

import { useMemo, useState } from 'react';
import Card from '@/components/Card';
import type { PredictionsData, WorkerPrediction } from './page.types';
import WorkerDetail from './WorkerDetail';
import PredictionCharts from './PredictionCharts';

interface Props {
  data: PredictionsData;
}

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
  recent_improv_rate: 'Recent Improv Rate',
  worker_count: 'Worker Count',
  active_families: 'Active Families',
};

export { FEATURE_LABELS };

export default function PredictionExplorer({ data }: Props) {
  const [selectedWorker, setSelectedWorker] = useState<WorkerPrediction | null>(null);
  const [filters, setFilters] = useState({
    algorithm: '',
    instance: '',
    problemType: '',
    week: '',
    run: '',
  });

  const predictions = data.predictions;

  // Unique filter values.
  const filterOptions = useMemo(() => ({
    algorithms: [...new Set(predictions.map(p => p.algorithm))].sort(),
    instances: [...new Set(predictions.map(p => p.instance))].sort(),
    problemTypes: [...new Set(predictions.map(p => p.problem_type))].sort(),
    weeks: [...new Set(predictions.map(p => p.week))].sort((a, b) => a - b),
    runs: [...new Set(predictions.map(p => p.run_id))].filter(Boolean).sort(),
  }), [predictions]);

  // Filtered predictions.
  const filtered = useMemo(() => {
    return predictions.filter(p => {
      if (filters.algorithm && p.algorithm !== filters.algorithm) return false;
      if (filters.instance && p.instance !== filters.instance) return false;
      if (filters.problemType && p.problem_type !== filters.problemType) return false;
      if (filters.week && p.week !== parseInt(filters.week)) return false;
      if (filters.run && p.run_id !== filters.run) return false;
      return true;
    });
  }, [predictions, filters]);

  // Summary stats.
  const stats = useMemo(() => {
    const total = filtered.length;
    const threshold = 0.5;

    // Improvement prediction accuracy.
    const improvedCorrect = filtered.filter(p =>
      (p.predicted.p_improved >= threshold) === p.actual.improved
    ).length;

    // Global best prediction accuracy.
    const gbCorrect = filtered.filter(p =>
      (p.predicted.p_global_best >= threshold) === p.actual.produced_global_best
    ).length;

    const avgImprovError = total > 0
      ? filtered.reduce((s, p) => s + Math.abs(p.error.improvement), 0) / total : 0;
    const avgRoiError = total > 0
      ? filtered.reduce((s, p) => s + Math.abs(p.error.roi), 0) / total : 0;

    return {
      total,
      improvedAccuracy: total > 0 ? (improvedCorrect / total) * 100 : 0,
      gbAccuracy: total > 0 ? (gbCorrect / total) * 100 : 0,
      avgImprovError,
      avgRoiError,
    };
  }, [filtered]);

  return (
    <div className="space-y-4">
      {/* Header */}
      <Card title="Worker Prediction Explorer">
        <p className="text-xs text-gray-500 mb-4">
          Inspect what the trained model predicts for real workers. Compare predicted vs actual
          outcomes to understand whether the model is reasoning sensibly.
        </p>
        <div className="grid grid-cols-2 sm:grid-cols-5 gap-3">
          <Stat label="Total Predictions" value={stats.total} colour="blue" />
          <Stat label="Improved Accuracy" value={`${stats.improvedAccuracy.toFixed(1)}%`} colour="emerald" />
          <Stat label="Global Best Accuracy" value={`${stats.gbAccuracy.toFixed(1)}%`} colour="emerald" />
          <Stat label="Avg Improvement Error" value={stats.avgImprovError.toFixed(1)} colour="amber" />
          <Stat label="Avg ROI Error" value={stats.avgRoiError.toFixed(4)} colour="amber" />
        </div>
      </Card>

      {/* Filters */}
      <Card title="Filters">
        <div className="grid grid-cols-2 sm:grid-cols-5 gap-3">
          <FilterSelect
            label="Algorithm"
            value={filters.algorithm}
            options={filterOptions.algorithms}
            onChange={v => setFilters(f => ({ ...f, algorithm: v }))}
          />
          <FilterSelect
            label="Problem Type"
            value={filters.problemType}
            options={filterOptions.problemTypes}
            onChange={v => setFilters(f => ({ ...f, problemType: v }))}
          />
          <FilterSelect
            label="Instance"
            value={filters.instance}
            options={filterOptions.instances}
            onChange={v => setFilters(f => ({ ...f, instance: v }))}
          />
          <FilterSelect
            label="Week"
            value={filters.week}
            options={filterOptions.weeks.map(String)}
            onChange={v => setFilters(f => ({ ...f, week: v }))}
          />
          <FilterSelect
            label="Run"
            value={filters.run}
            options={filterOptions.runs}
            onChange={v => setFilters(f => ({ ...f, run: v }))}
          />
        </div>
        <p className="text-[10px] text-gray-500 mt-2">
          Showing {filtered.length} of {predictions.length} workers
        </p>
      </Card>

      {/* Charts */}
      <PredictionCharts predictions={filtered} />

      {/* Worker Table */}
      <Card title="Worker Predictions Table">
        <div className="overflow-x-auto max-h-[500px] overflow-y-auto">
          <table className="w-full text-[10px]">
            <thead className="sticky top-0 bg-gray-850 z-10">
              <tr className="text-gray-500 uppercase">
                <th className="text-left p-1.5">#</th>
                <th className="text-left p-1.5">Instance</th>
                <th className="text-left p-1.5">Algo</th>
                <th className="text-right p-1.5">W/D</th>
                <th className="text-center p-1.5">Improved</th>
                <th className="text-right p-1.5">P(Impr)</th>
                <th className="text-center p-1.5">GB</th>
                <th className="text-right p-1.5">P(GB)</th>
                <th className="text-right p-1.5">Actual Δ</th>
                <th className="text-right p-1.5">Pred Δ</th>
                <th className="text-right p-1.5">Error</th>
              </tr>
            </thead>
            <tbody>
              {filtered.slice(0, 200).map(p => (
                <tr
                  key={p.index}
                  className={`border-t border-gray-800 cursor-pointer hover:bg-gray-800/50 ${
                    selectedWorker?.index === p.index ? 'bg-blue-900/20' : ''
                  }`}
                  onClick={() => setSelectedWorker(p)}
                >
                  <td className="p-1.5 text-gray-500">{p.index}</td>
                  <td className="p-1.5 text-blue-400 truncate max-w-[100px]">{p.instance}</td>
                  <td className="p-1.5 text-emerald-400">{p.algorithm}</td>
                  <td className="text-right p-1.5">W{p.week}/D{p.depth}</td>
                  <td className="text-center p-1.5">
                    <span className={p.actual.improved ? 'text-emerald-400' : 'text-gray-600'}>
                      {p.actual.improved ? '✓' : '✗'}
                    </span>
                  </td>
                  <td className="text-right p-1.5">
                    <span className={p.predicted.p_improved >= 0.5 ? 'text-emerald-400' : 'text-gray-500'}>
                      {(p.predicted.p_improved * 100).toFixed(0)}%
                    </span>
                  </td>
                  <td className="text-center p-1.5">
                    <span className={p.actual.produced_global_best ? 'text-amber-400' : 'text-gray-600'}>
                      {p.actual.produced_global_best ? '⭐' : '—'}
                    </span>
                  </td>
                  <td className="text-right p-1.5">
                    <span className={p.predicted.p_global_best >= 0.1 ? 'text-amber-400' : 'text-gray-500'}>
                      {(p.predicted.p_global_best * 100).toFixed(0)}%
                    </span>
                  </td>
                  <td className="text-right p-1.5">{p.actual.improvement_amount.toLocaleString()}</td>
                  <td className="text-right p-1.5">{p.predicted.expected_improvement.toFixed(0)}</td>
                  <td className={`text-right p-1.5 ${
                    Math.abs(p.error.improvement) > 500 ? 'text-red-400' : 'text-gray-500'
                  }`}>
                    {p.error.improvement > 0 ? '+' : ''}{p.error.improvement.toFixed(0)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          {filtered.length > 200 && (
            <p className="text-[10px] text-gray-500 p-2 text-center">
              Showing first 200 of {filtered.length}. Use filters to narrow.
            </p>
          )}
        </div>
      </Card>

      {/* Worker Detail Panel */}
      {selectedWorker && (
        <WorkerDetail worker={selectedWorker} onClose={() => setSelectedWorker(null)} />
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

function FilterSelect({ label, value, options, onChange }: {
  label: string; value: string; options: string[]; onChange: (v: string) => void;
}) {
  return (
    <div>
      <label className="text-[9px] text-gray-500 uppercase block mb-1">{label}</label>
      <select
        className="w-full bg-gray-800 border border-gray-700 rounded text-xs p-1.5 text-gray-200"
        value={value}
        onChange={e => onChange(e.target.value)}
      >
        <option value="">All</option>
        {options.map(o => <option key={o} value={o}>{o}</option>)}
      </select>
    </div>
  );
}
