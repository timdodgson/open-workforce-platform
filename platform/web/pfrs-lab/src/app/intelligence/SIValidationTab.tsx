'use client';

import Card from '@/components/Card';

interface Props {
  completed: number;
  totalExpected: number;
  si2RunIds: string[];
}

export default function SIValidationTab({ completed, totalExpected, si2RunIds }: Props) {
  const pending = totalExpected - completed;
  const status = completed === 0 ? 'pending' : completed >= totalExpected ? 'completed' : 'running';
  const pct = totalExpected > 0 ? Math.round((completed / totalExpected) * 100) : 0;

  return (
    <div className="space-y-4">
      <Card title="Policy layer validation (240-run sweep)">
        <p className="text-xs text-gray-500 mb-4">
          Compares <code className="text-blue-400">rules</code>, <code className="text-blue-400">hybrid</code>, and{' '}
          <code className="text-blue-400">learned</code> across CVRP, JSS, VRPTW (portfolio + single-search configs).
          Run from <code className="text-gray-400">platform/go/scripts/validate-si2.ps1</code>.
        </p>

        <div className="grid grid-cols-3 gap-3 mb-4">
          <div className="bg-gray-800 rounded p-3 text-center">
            <p className="text-lg font-semibold text-emerald-400">{completed}</p>
            <p className="text-[9px] text-gray-500 uppercase">Completed</p>
          </div>
          <div className="bg-gray-800 rounded p-3 text-center">
            <p className="text-lg font-semibold text-amber-400">{pending}</p>
            <p className="text-[9px] text-gray-500 uppercase">Pending</p>
          </div>
          <div className="bg-gray-800 rounded p-3 text-center">
            <p className="text-lg font-semibold text-blue-400">{pct}%</p>
            <p className="text-[9px] text-gray-500 uppercase">Progress</p>
          </div>
        </div>

        <p className="text-[10px] text-gray-500 mb-2">
          Status: <span className={status === 'completed' ? 'text-emerald-400' : status === 'running' ? 'text-amber-400' : 'text-gray-400'}>{status}</span>
        </p>

        {si2RunIds.length > 0 ? (
          <div className="max-h-48 overflow-y-auto text-[10px] font-mono text-gray-500 space-y-0.5">
            {si2RunIds.slice(0, 30).map(id => (
              <div key={id}>{id}</div>
            ))}
            {si2RunIds.length > 30 && <div>…and {si2RunIds.length - 30} more</div>}
          </div>
        ) : (
          <p className="text-[10px] text-gray-600">No si2-* or val-* labelled runs found yet.</p>
        )}
      </Card>
    </div>
  );
}
