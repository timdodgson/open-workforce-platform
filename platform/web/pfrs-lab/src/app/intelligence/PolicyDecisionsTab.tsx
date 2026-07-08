'use client';

export type { PolicyDecisionRecord, PolicyLearningReport } from '@/lib/types/intelligence';

import { useMemo } from 'react';
import Card from '@/components/Card';
import TabSpinner from '@/components/TabSpinner';
import type { PolicyDecisionRecord, PolicyLearningReport } from '@/lib/types/intelligence';

interface Props {
  decisions: PolicyDecisionRecord[];
  learningReports: PolicyLearningReport[];
  evalCount: number;
  registryVersionCount: number;
  loading?: boolean;
}

export default function PolicyDecisionsTab({
  decisions,
  learningReports,
  evalCount,
  registryVersionCount,
  loading,
}: Props) {
  const stats = useMemo(() => {
    const learned = decisions.filter(d => d.policyUsed.includes('learned') || d.policyUsed === 'restart');
    const earlyStops = decisions.filter(d => d.action === 'early_stop');
    const restarts = decisions.filter(d => d.action === 'restart');
    const runs = new Set(decisions.map(d => d.runId)).size;
    return { total: decisions.length, learned: learned.length, earlyStops: earlyStops.length, restarts: restarts.length, runs };
  }, [decisions]);

  const byRun = useMemo(() => {
    const map = new Map<string, PolicyDecisionRecord[]>();
    for (const d of decisions) {
      const list = map.get(d.runId) || [];
      list.push(d);
      map.set(d.runId, list);
    }
    return [...map.entries()].sort((a, b) => b[1].length - a[1].length).slice(0, 20);
  }, [decisions]);

  if (loading) {
    return (
      <Card title="SI 2.0 Policies">
        <TabSpinner label="Loading policy telemetry…" />
      </Card>
    );
  }

  if (decisions.length === 0 && learningReports.length === 0 && registryVersionCount === 0) {
    return (
      <Card title="SI 2.0 Policies">
        <p className="text-xs text-gray-500 text-center py-12">
          No SI 2.0 policy telemetry yet. Run with{' '}
          <code className="text-blue-400">--policy-mode hybrid --run-label my-run</code>
        </p>
      </Card>
    );
  }

  return (
    <div className="space-y-4">
      <Card title="SI 2.0 Policy Telemetry">
        <div className="grid grid-cols-2 sm:grid-cols-4 lg:grid-cols-6 gap-2 mb-4">
          <Stat label="Runs" value={stats.runs} />
          <Stat label="Decisions" value={stats.total} />
          <Stat label="Learned" value={stats.learned} colour="emerald" />
          <Stat label="Early stops" value={stats.earlyStops} colour="amber" />
          <Stat label="Restarts" value={stats.restarts} colour="blue" />
          <Stat label="Eval rows" value={evalCount} />
        </div>
        <p className="text-[10px] text-gray-500">
          Registry: {registryVersionCount} active policy version(s) · {learningReports.length} learning report(s)
        </p>
      </Card>

      {byRun.length > 0 && (
        <Card title="Recent runs with policy_decisions.csv">
          <div className="overflow-x-auto">
            <table className="w-full text-[10px] text-left">
              <thead className="text-gray-500 border-b border-gray-700">
                <tr>
                  <th className="py-1 pr-2">Run</th>
                  <th className="py-1 pr-2">Checkpoints</th>
                  <th className="py-1 pr-2">Learned</th>
                  <th className="py-1 pr-2">Last action</th>
                  <th className="py-1">Last policy</th>
                </tr>
              </thead>
              <tbody>
                {byRun.map(([runId, rows]) => {
                  const last = rows[rows.length - 1];
                  const learnedCount = rows.filter(r => r.policyUsed.includes('learned') || r.policyUsed === 'restart').length;
                  return (
                    <tr key={runId} className="border-b border-gray-800 text-gray-300">
                      <td className="py-1 pr-2 font-mono">{runId}</td>
                      <td className="py-1 pr-2">{rows.length}</td>
                      <td className="py-1 pr-2 text-emerald-400">{learnedCount}</td>
                      <td className="py-1 pr-2">{last.action}</td>
                      <td className="py-1">{last.policyUsed}</td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        </Card>
      )}

      {learningReports.length > 0 && (
        <Card title="Post-run learning recommendations">
          <div className="space-y-1 text-[10px] text-gray-400">
            {learningReports.slice(0, 10).map(r => (
              <div key={r.runId} className="flex justify-between border-b border-gray-800 py-1">
                <span className="font-mono text-gray-300">{r.runId}</span>
                <span className="text-amber-400">{r.action}</span>
                <span className="text-gray-500 truncate max-w-[40%]">{r.reason}</span>
              </div>
            ))}
          </div>
        </Card>
      )}
    </div>
  );
}

function Stat({ label, value, colour }: { label: string; value: number; colour?: string }) {
  const text = colour === 'emerald' ? 'text-emerald-400' : colour === 'amber' ? 'text-amber-400' : colour === 'blue' ? 'text-blue-400' : 'text-gray-200';
  return (
    <div className="bg-gray-800 rounded p-2 text-center">
      <p className={`text-sm font-semibold ${text}`}>{value}</p>
      <p className="text-[9px] text-gray-500 uppercase">{label}</p>
    </div>
  );
}
