'use client';

import { useState } from 'react';
import Card from '@/components/Card';
import { RunInfo } from './page';

interface CompareMetrics {
  id: string;
  objective: number;
  runtime: number;
  candidates: number;
  improved: number;
  feasible: boolean;
}

export default function CompareShell({ runs }: { runs: RunInfo[] }) {
  const [runAId, setRunAId] = useState(runs[0]?.id || '');
  const [runBId, setRunBId] = useState(runs[1]?.id || '');

  const runA = runs.find(r => r.id === runAId);
  const runB = runs.find(r => r.id === runBId);

  if (!runA || !runB) {
    return (
      <Card title="Compare">
        <p className="text-xs text-gray-500">Select two runs to compare.</p>
      </Card>
    );
  }

  return (
    <div className="space-y-4">
      {/* Run selector */}
      <Card title="Head-to-Head Comparison">
        <div className="flex gap-4 items-center">
          <div className="flex-1">
            <label className="text-[10px] text-gray-500 uppercase block mb-1">Run A</label>
            <select value={runAId} onChange={e => setRunAId(e.target.value)}
              className="w-full bg-gray-800 border border-gray-700 rounded px-2 py-1 text-xs text-blue-400">
              {runs.map(r => (
                <option key={r.id} value={r.id}>
                  {r.id} ({r.problemType.toUpperCase()} {r.mode} — {r.objective.toLocaleString()})
                </option>
              ))}
            </select>
          </div>
          <span className="text-gray-600 text-lg">vs</span>
          <div className="flex-1">
            <label className="text-[10px] text-gray-500 uppercase block mb-1">Run B</label>
            <select value={runBId} onChange={e => setRunBId(e.target.value)}
              className="w-full bg-gray-800 border border-gray-700 rounded px-2 py-1 text-xs text-rose-400">
              {runs.map(r => (
                <option key={r.id} value={r.id}>
                  {r.id} ({r.problemType.toUpperCase()} {r.mode} — {r.objective.toLocaleString()})
                </option>
              ))}
            </select>
          </div>
        </div>
      </Card>

      {/* Comparison result */}
      <ComparisonResult runA={runA} runB={runB} />
    </div>
  );
}

function ComparisonResult({ runA, runB }: { runA: RunInfo; runB: RunInfo }) {
  const objLabel = runA.problemType === 'cvrp' || runA.problemType === 'vrptw' ? 'Distance' :
                   runA.problemType === 'jss' ? 'Makespan' : 'Penalty';

  const diff = runA.objective - runB.objective;
  const diffPct = runB.objective > 0 ? ((diff / runB.objective) * 100).toFixed(1) : '0';
  const winner = diff < 0 ? 'A' : diff > 0 ? 'B' : 'Tie';

  return (
    <>
      {/* Conclusion */}
      <Card title="Conclusion">
        {winner === 'Tie' ? (
          <p className="text-sm text-gray-300">Both runs achieved identical objectives ({runA.objective.toLocaleString()}).</p>
        ) : (
          <p className="text-sm text-gray-300">
            <span className={winner === 'A' ? 'text-blue-400' : 'text-rose-400'}>Run {winner}</span>{' '}
            is better by {Math.abs(diff).toLocaleString()} ({Math.abs(Number(diffPct))}%).
          </p>
        )}
      </Card>

      {/* Evidence */}
      <Card title="Comparison">
        <table className="w-full text-xs">
          <thead>
            <tr className="text-gray-500 uppercase">
              <th className="text-left p-2">Metric</th>
              <th className="text-right p-2 text-blue-400">Run A</th>
              <th className="text-right p-2 text-rose-400">Run B</th>
              <th className="text-right p-2">Difference</th>
            </tr>
          </thead>
          <tbody>
            <tr className="border-t border-gray-800">
              <td className="p-2">{objLabel}</td>
              <td className="text-right p-2 text-blue-400">{runA.objective.toLocaleString()}</td>
              <td className="text-right p-2 text-rose-400">{runB.objective.toLocaleString()}</td>
              <td className={`text-right p-2 ${diff < 0 ? 'text-emerald-400' : diff > 0 ? 'text-red-400' : 'text-gray-500'}`}>
                {diff === 0 ? '=' : `${diff > 0 ? '+' : ''}${diff.toLocaleString()} (${diffPct}%)`}
              </td>
            </tr>
            <tr className="border-t border-gray-800">
              <td className="p-2">Domain</td>
              <td className="text-right p-2 text-blue-400">{runA.problemType.toUpperCase()}</td>
              <td className="text-right p-2 text-rose-400">{runB.problemType.toUpperCase()}</td>
              <td className="text-right p-2 text-gray-500">
                {runA.problemType === runB.problemType ? 'Same' : 'Different ⚠️'}
              </td>
            </tr>
            <tr className="border-t border-gray-800">
              <td className="p-2">Algorithm</td>
              <td className="text-right p-2 text-blue-400">{runA.mode}</td>
              <td className="text-right p-2 text-rose-400">{runB.mode}</td>
              <td className="text-right p-2 text-gray-500">
                {runA.mode === runB.mode ? 'Same' : 'Different'}
              </td>
            </tr>
          </tbody>
        </table>
      </Card>

      {/* Warning for cross-domain comparison */}
      {runA.problemType !== runB.problemType && (
        <Card title="Warning">
          <p className="text-xs text-amber-400">
            Comparing runs from different domains ({runA.problemType.toUpperCase()} vs {runB.problemType.toUpperCase()}).
            Objectives are not directly comparable across domains.
          </p>
        </Card>
      )}
    </>
  );
}
