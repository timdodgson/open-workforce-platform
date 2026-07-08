'use client';

import Card from '@/components/Card';
import type { PolicyVersion } from '@/lib/types/intelligence';

export default function PromotionTab({ versions }: { versions: PolicyVersion[] }) {
  if (versions.length === 0) {
    return (
      <Card title="Policy Promotion">
        <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500 text-xs">
          No policy registry found. Train policies to populate promotion pipeline.
        </div>
      </Card>
    );
  }

  const production = versions.filter((v) => v.status === 'active');
  const candidates = versions.filter((v) => v.status === 'training' || v.status === 'shadow');
  const retired = versions.filter((v) => v.status === 'retired');

  return (
    <div className="space-y-4">
      <Card title="Policy Promotion Pipeline">
        <p className="text-xs text-gray-400 mb-4">
          Policies advance: Candidate → Shadow → Active. Failed policies are never promoted.
        </p>
        <div className="grid grid-cols-3 gap-3 mb-4">
          <Metric label="Active" value={production.length} colour="emerald" />
          <Metric label="Candidates" value={candidates.length} colour="amber" />
          <Metric label="Retired" value={retired.length} colour="gray" />
        </div>
      </Card>

      {production.length > 0 && (
        <Card title="Production Policies">
          <VersionTable versions={production} />
        </Card>
      )}
      {candidates.length > 0 && (
        <Card title="Candidates / Shadow">
          <VersionTable versions={candidates} />
        </Card>
      )}
    </div>
  );
}

function VersionTable({ versions }: { versions: PolicyVersion[] }) {
  return (
    <div className="overflow-x-auto">
      <table className="w-full text-[10px]">
        <thead>
          <tr className="text-gray-500 uppercase border-b border-gray-700">
            <th className="text-left p-1.5">Version</th>
            <th className="text-left p-1.5">Domain</th>
            <th className="text-left p-1.5">Type</th>
            <th className="text-right p-1.5">Offline</th>
            <th className="text-right p-1.5">Prod</th>
            <th className="text-right p-1.5">Runs</th>
          </tr>
        </thead>
        <tbody>
          {versions.map((v) => (
            <tr key={v.id} className="border-b border-gray-800">
              <td className="p-1.5 text-blue-400">v{v.version}</td>
              <td className="p-1.5">{v.domain}</td>
              <td className="p-1.5">{v.decision_type}</td>
              <td className="p-1.5 text-right">{(v.offline_accuracy * 100).toFixed(0)}%</td>
              <td className="p-1.5 text-right">{(v.production_accuracy * 100).toFixed(0)}%</td>
              <td className="p-1.5 text-right">{v.production_runs}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function Metric({ label, value, colour }: { label: string; value: number; colour: string }) {
  const c = colour === 'emerald' ? 'text-emerald-400' : colour === 'amber' ? 'text-amber-400' : 'text-gray-400';
  return (
    <div className="bg-gray-800 rounded p-3 text-center">
      <p className="text-[9px] text-gray-500 uppercase">{label}</p>
      <p className={`text-xl font-bold ${c}`}>{value}</p>
    </div>
  );
}
