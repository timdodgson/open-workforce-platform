'use client';
import { useState } from 'react';
import Card from '@/components/Card';
import { TreeNode } from '@/lib/types';

export default function TreeView({ nodes }: { nodes: TreeNode[] }) {
  const [selected, setSelected] = useState<TreeNode | null>(null);

  // Group by week
  const byWeek = new Map<number, TreeNode[]>();
  for (const n of nodes) {
    const list = byWeek.get(n.week) || [];
    list.push(n);
    byWeek.set(n.week, list);
  }
  const weeks = Array.from(byWeek.keys()).sort((a, b) => a - b);

  return (
    <>
      <Card title="Search Tree">
        <div className="flex gap-4 overflow-x-auto pb-4">
          {weeks.map(w => {
            const weekNodes = byWeek.get(w)!;
            return (
              <div key={w} className="flex flex-col items-center min-w-[80px]">
                <div className="text-[10px] text-gray-500 mb-2">W{w}</div>
                {weekNodes.map(n => {
                  const color = n.winning ? 'bg-emerald-500'
                    : n.retained ? 'bg-blue-500'
                    : 'bg-gray-600 opacity-50';
                  const border = selected?.pathID === n.pathID
                    ? 'ring-2 ring-blue-400' : '';
                  return (
                    <button key={n.pathID}
                      onClick={() => setSelected(n)}
                      className={`w-5 h-5 rounded-full ${color} ${border} mb-1 hover:ring-1 hover:ring-white transition-all`}
                      title={`Path ${n.pathID} — penalty ${n.weekPenalty} (cum ${n.cumulativePenalty})`}
                    />
                  );
                })}
              </div>
            );
          })}
        </div>
        <div className="flex gap-4 text-[10px] text-gray-500 mt-2">
          <span><span className="inline-block w-3 h-3 rounded-full bg-emerald-500 mr-1"></span>Winning</span>
          <span><span className="inline-block w-3 h-3 rounded-full bg-blue-500 mr-1"></span>Retained</span>
          <span><span className="inline-block w-3 h-3 rounded-full bg-gray-600 opacity-50 mr-1"></span>Discarded</span>
        </div>
      </Card>

      {/* Node detail panel */}
      {selected && (
        <Card title={`Path ${selected.pathID} — Week ${selected.week}`}>
          <div className="grid grid-cols-2 sm:grid-cols-3 gap-2 text-xs">
            <Detail label="Week" value={String(selected.week)} />
            <Detail label="Path ID" value={String(selected.pathID)} />
            <Detail label="Parent ID" value={selected.parentID <= 0 ? 'root' : String(selected.parentID)} />
            <Detail label="Seed" value={String(selected.seed)} />
            <Detail label="Week Penalty" value={selected.weekPenalty.toLocaleString()} />
            <Detail label="Cumulative" value={selected.cumulativePenalty.toLocaleString()} />
            <Detail label="Retained Rank" value={selected.retainedRank > 0 ? String(selected.retainedRank) : '—'} />
            <Detail label="Workers" value={String(selected.workersStarted)} />
            <Detail label="Candidates" value={selected.candidates.toLocaleString()} />
            <Detail label="Accepted Worse" value={selected.saAcceptedWorse.toLocaleString()} />
            <Detail label="Hard Reject %" value={`${selected.hardRejectRate.toFixed(1)}%`} />
            <Detail label="Winning" value={selected.winning ? 'Yes' : 'No'} />
          </div>
        </Card>
      )}
    </>
  );
}

function Detail({ label, value }: { label: string; value: string }) {
  return (
    <div className="bg-gray-800 rounded px-2 py-1">
      <div className="text-[9px] text-gray-500 uppercase">{label}</div>
      <div className="text-gray-200">{value}</div>
    </div>
  );
}
