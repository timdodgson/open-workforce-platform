'use client';
import { useState } from 'react';
import Card from '@/components/Card';
import { WorkerLifecycle } from '@/lib/types';

type SortKey = 'workerID' | 'finishTimeMs' | 'finalPenalty' | 'branchCount' | 'plateauCount' | 'bestCandidate' | 'producedGlobalBest';

export default function WorkerTable({ workers }: { workers: WorkerLifecycle[] }) {
  const [sortKey, setSortKey] = useState<SortKey>('workerID');
  const [sortAsc, setSortAsc] = useState(true);

  if (workers.length === 0) return null;

  const sorted = [...workers].sort((a, b) => {
    const av = a[sortKey] as number | boolean;
    const bv = b[sortKey] as number | boolean;
    const cmp = av < bv ? -1 : av > bv ? 1 : 0;
    return sortAsc ? cmp : -cmp;
  });

  function header(label: string, key: SortKey) {
    return (
      <th className="p-2 cursor-pointer hover:text-blue-400 select-none"
        onClick={() => { if (sortKey === key) setSortAsc(!sortAsc); else { setSortKey(key); setSortAsc(true); } }}>
        {label} {sortKey === key ? (sortAsc ? '▲' : '▼') : ''}
      </th>
    );
  }

  return (
    <Card title="Worker Table (sortable)">
      <div className="overflow-x-auto max-h-[400px] overflow-y-auto">
        <table className="w-full text-xs">
          <thead className="sticky top-0 bg-gray-900">
            <tr className="text-gray-500 uppercase text-right">
              {header('Worker', 'workerID')}
              {header('Runtime', 'finishTimeMs')}
              {header('Final Pen', 'finalPenalty')}
              {header('Branches', 'branchCount')}
              {header('Plateaus', 'plateauCount')}
              {header('Best Cand', 'bestCandidate')}
              {header('Global?', 'producedGlobalBest')}
            </tr>
          </thead>
          <tbody>
            {sorted.slice(0, 100).map(w => (
              <tr key={`${w.week}-${w.workerID}`} className={`border-t border-gray-800 ${w.producedGlobalBest ? 'bg-emerald-900/20' : ''}`}>
                <td className="p-2">W{w.workerID}</td>
                <td className="p-2 text-right">{(w.finishTimeMs / 1000).toFixed(1)}s</td>
                <td className="p-2 text-right">{w.finalPenalty}</td>
                <td className="p-2 text-right">{w.branchCount}</td>
                <td className="p-2 text-right">{w.plateauCount}</td>
                <td className="p-2 text-right">{(w.bestCandidate / 1000).toFixed(0)}K</td>
                <td className="p-2 text-center">{w.producedGlobalBest ? '✓' : ''}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      {workers.length > 100 && <p className="text-xs text-gray-500 mt-2">Showing first 100 of {workers.length} workers.</p>}
    </Card>
  );
}
