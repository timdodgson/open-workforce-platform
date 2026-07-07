'use client';

import { useMemo } from 'react';
import Card from '@/components/Card';
import { LearningRecord } from './page';

export default function WorkerLearningDashboard({ records }: { records: LearningRecord[] }) {
  const stats = useMemo(() => {
    const improved = records.filter(r => r.improved);
    const successRate = records.length > 0 ? (improved.length / records.length) * 100 : 0;
    const avgImprovement = improved.length > 0
      ? improved.reduce((s, r) => s + r.improvementAmount, 0) / improved.length : 0;
    const avgImprovPer100K = improved.length > 0
      ? improved.reduce((s, r) => s + r.improvPer100K, 0) / improved.length : 0;
    const avgROI = improved.length > 0
      ? improved.reduce((s, r) => s + r.improvPerCPU, 0) / improved.length : 0;
    const avgRuntime = records.length > 0
      ? records.reduce((s, r) => s + r.runtimeMs, 0) / records.length : 0;
    const avgCandidates = records.length > 0
      ? records.reduce((s, r) => s + r.candidatesEval, 0) / records.length : 0;

    return { successRate, avgImprovement, avgImprovPer100K, avgROI, avgRuntime, avgCandidates };
  }, [records]);

  // Group by algorithm.
  const byAlgorithm = useMemo(() => {
    const map = new Map<string, LearningRecord[]>();
    for (const r of records) {
      const existing = map.get(r.algorithm) || [];
      existing.push(r);
      map.set(r.algorithm, existing);
    }
    return Array.from(map.entries()).map(([algo, recs]) => {
      const improved = recs.filter(r => r.improved);
      return {
        algorithm: algo,
        count: recs.length,
        successRate: recs.length > 0 ? (improved.length / recs.length) * 100 : 0,
        avgImprovement: improved.length > 0
          ? improved.reduce((s, r) => s + r.improvementAmount, 0) / improved.length : 0,
        avgImprovPer100K: improved.length > 0
          ? improved.reduce((s, r) => s + r.improvPer100K, 0) / improved.length : 0,
      };
    }).sort((a, b) => b.avgImprovPer100K - a.avgImprovPer100K);
  }, [records]);

  // Group by domain.
  const byDomain = useMemo(() => {
    const map = new Map<string, LearningRecord[]>();
    for (const r of records) {
      const existing = map.get(r.problemType) || [];
      existing.push(r);
      map.set(r.problemType, existing);
    }
    return Array.from(map.entries()).map(([domain, recs]) => {
      const improved = recs.filter(r => r.improved);
      return {
        domain,
        count: recs.length,
        successRate: recs.length > 0 ? (improved.length / recs.length) * 100 : 0,
        avgImprovement: improved.length > 0
          ? improved.reduce((s, r) => s + r.improvementAmount, 0) / improved.length : 0,
      };
    }).sort((a, b) => b.avgImprovement - a.avgImprovement);
  }, [records]);

  // Top and worst performers.
  const sorted = useMemo(() =>
    [...records].sort((a, b) => b.improvPer100K - a.improvPer100K), [records]);
  const top5 = sorted.slice(0, 5);
  const worst5 = sorted.slice(-5).reverse();

  // Observations.
  const observations = useMemo(() => {
    const obs: string[] = [];

    if (records.length < 10) {
      obs.push('Current dataset is sparse; stronger conclusions require more runs across domains.');
    }

    if (byAlgorithm.length > 1) {
      const best = byAlgorithm[0];
      obs.push(`${best.algorithm.toUpperCase()} produced the highest mean improvement per 100K evaluations (${best.avgImprovPer100K.toFixed(1)}).`);

      const lowSuccess = byAlgorithm.filter(a => a.successRate < 50);
      for (const a of lowSuccess) {
        obs.push(`${a.algorithm.toUpperCase()} has low success rate (${a.successRate.toFixed(0)}%).`);
      }
    }

    if (byDomain.length > 1) {
      const bestDomain = byDomain[0];
      obs.push(`${bestDomain.domain.toUpperCase()} shows the highest average improvement (${bestDomain.avgImprovement.toFixed(0)}).`);
    }

    const allImproved = records.every(r => r.improved);
    if (allImproved && records.length > 3) {
      obs.push('All workers improved — constructive baselines may be weak or iteration budget is generous.');
    }

    return obs;
  }, [records, byAlgorithm, byDomain]);

  return (
    <div className="space-y-4">
      {/* Overview Stats */}
      <Card title="Worker Learning Telemetry">
        <p className="text-xs text-gray-500 mb-4">
          Training dataset for future ML-based worker selection. One row per completed worker/run.
        </p>
        <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3">
          <Stat label="Total Records" value={records.length} colour="blue" />
          <Stat label="Success Rate" value={`${stats.successRate.toFixed(0)}%`} colour="emerald" />
          <Stat label="Avg Improvement" value={stats.avgImprovement.toFixed(0)} colour="emerald" />
          <Stat label="Improv / 100K" value={stats.avgImprovPer100K.toFixed(1)} colour="amber" />
          <Stat label="Improv / CPU ms" value={stats.avgROI.toFixed(2)} colour="amber" />
          <Stat label="Avg Runtime" value={`${stats.avgRuntime.toFixed(0)}ms`} colour="blue" />
        </div>
      </Card>

      {/* Algorithm Comparison */}
      <Card title="By Algorithm">
        <table className="w-full text-xs">
          <thead>
            <tr className="text-gray-500 uppercase">
              <th className="text-left p-2">Algorithm</th>
              <th className="text-right p-2">Records</th>
              <th className="text-right p-2">Success %</th>
              <th className="text-right p-2">Avg Improvement</th>
              <th className="text-right p-2">Improv / 100K</th>
            </tr>
          </thead>
          <tbody>
            {byAlgorithm.map((a, i) => (
              <tr key={a.algorithm} className={`border-t border-gray-800 ${i === 0 ? 'bg-emerald-900/10' : ''}`}>
                <td className="p-2 text-blue-400 font-semibold">{a.algorithm.toUpperCase()}</td>
                <td className="text-right p-2">{a.count}</td>
                <td className="text-right p-2">{a.successRate.toFixed(0)}%</td>
                <td className="text-right p-2 text-emerald-400">{a.avgImprovement.toFixed(0)}</td>
                <td className="text-right p-2 text-amber-400">{a.avgImprovPer100K.toFixed(1)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </Card>

      {/* Domain Comparison */}
      <Card title="By Domain">
        <table className="w-full text-xs">
          <thead>
            <tr className="text-gray-500 uppercase">
              <th className="text-left p-2">Domain</th>
              <th className="text-right p-2">Records</th>
              <th className="text-right p-2">Success %</th>
              <th className="text-right p-2">Avg Improvement</th>
            </tr>
          </thead>
          <tbody>
            {byDomain.map(d => (
              <tr key={d.domain} className="border-t border-gray-800">
                <td className="p-2 text-blue-400">{d.domain.toUpperCase()}</td>
                <td className="text-right p-2">{d.count}</td>
                <td className="text-right p-2">{d.successRate.toFixed(0)}%</td>
                <td className="text-right p-2 text-emerald-400">{d.avgImprovement.toFixed(0)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </Card>

      {/* Top Performers */}
      <Card title="Top Performing Workers (by Improvement / 100K)">
        <RecordTable records={top5} />
      </Card>

      {/* Worst Performers */}
      {worst5.length > 0 && worst5[0].improvPer100K < top5[top5.length - 1]?.improvPer100K && (
        <Card title="Lowest Performing Workers">
          <RecordTable records={worst5} />
        </Card>
      )}

      {/* Observations */}
      <Card title="Observations">
        {observations.length === 0 ? (
          <p className="text-gray-500 text-sm italic">Insufficient data for observations.</p>
        ) : (
          <ul className="space-y-2">
            {observations.map((obs, i) => (
              <li key={i} className="text-sm text-gray-300 flex gap-2">
                <span className="text-emerald-400">•</span>
                {obs}
              </li>
            ))}
          </ul>
        )}
      </Card>
    </div>
  );
}

function RecordTable({ records }: { records: LearningRecord[] }) {
  return (
    <div className="overflow-x-auto">
      <table className="w-full text-[10px]">
        <thead>
          <tr className="text-gray-500 uppercase">
            <th className="text-left p-1.5">Run</th>
            <th className="text-left p-1.5">Domain</th>
            <th className="text-left p-1.5">Instance</th>
            <th className="text-left p-1.5">Algo</th>
            <th className="text-right p-1.5">Seed</th>
            <th className="text-right p-1.5">Initial</th>
            <th className="text-right p-1.5">Final</th>
            <th className="text-right p-1.5">Δ</th>
            <th className="text-right p-1.5">Runtime</th>
            <th className="text-right p-1.5">Cands</th>
            <th className="text-right p-1.5">Imp/100K</th>
          </tr>
        </thead>
        <tbody>
          {records.map((r, i) => (
            <tr key={i} className="border-t border-gray-800">
              <td className="p-1.5 text-blue-400 font-mono">{r.runId.slice(0, 20)}</td>
              <td className="p-1.5">{r.problemType}</td>
              <td className="p-1.5">{r.instance}</td>
              <td className="p-1.5 text-emerald-400">{r.algorithm}</td>
              <td className="text-right p-1.5">{r.seed}</td>
              <td className="text-right p-1.5">{r.parentObjective.toLocaleString()}</td>
              <td className="text-right p-1.5 text-emerald-400">{r.finalObjective.toLocaleString()}</td>
              <td className="text-right p-1.5 text-amber-400">{r.improvementAmount.toLocaleString()}</td>
              <td className="text-right p-1.5">{r.runtimeMs}ms</td>
              <td className="text-right p-1.5">{(r.candidatesEval / 1000).toFixed(0)}K</td>
              <td className="text-right p-1.5 font-semibold text-amber-400">{r.improvPer100K.toFixed(1)}</td>
            </tr>
          ))}
        </tbody>
      </table>
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
