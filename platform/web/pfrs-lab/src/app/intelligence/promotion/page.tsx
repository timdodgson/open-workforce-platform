import { getStorageProvider } from '@/lib/storage';
import Card from '@/components/Card';
import type { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Policy Promotion',
  description: 'Automatic policy promotion pipeline with history and rollback tracking.',
};

export const revalidate = 60;

interface PolicyVersion {
  id: string;
  version: string;
  domain: string;
  decision_type: string;
  status: string;
  offline_accuracy: number;
  shadow_accuracy: number;
  production_accuracy: number;
  production_runs: number;
  regret_vs_rules: number;
  drift_detected: boolean;
  promoted_at?: string;
  retired_at?: string;
  rolled_back_from?: string;
  rollback_reason?: string;
}

export default async function PromotionPage() {
  const storage = getStorageProvider();
  const content = await storage.readRootFile('policy_registry.json');
  let versions: PolicyVersion[] = [];
  if (content) {
    try { versions = JSON.parse(content).versions || []; } catch { /* */ }
  }

  const production = versions.filter(v => v.status === 'active');
  const candidates = versions.filter(v => v.status === 'training' || v.status === 'shadow');
  const retired = versions.filter(v => v.status === 'retired');
  const rolledBack = versions.filter(v => v.rolled_back_from);

  return (
    <div className="space-y-6">
      <Card title="Policy Promotion Pipeline">
        <p className="text-sm text-slate-600 mb-4">
          Policies advance automatically: Candidate → Shadow → Hybrid → Production.
          Failed policies are never promoted. Rollback is always supported.
        </p>
        <div className="flex items-center gap-2 text-xs text-slate-500 mb-4">
          <Stage label="Candidate" color="amber" />
          <span>→</span>
          <Stage label="Shadow" color="blue" />
          <span>→</span>
          <Stage label="Active" color="emerald" />
          <span className="ml-4 text-slate-300">|</span>
          <Stage label="Retired" color="slate" />
        </div>
      </Card>

      {/* Current Production */}
      <Card title="Current Production Policies">
        {production.length === 0 ? (
          <p className="text-sm text-slate-500">No active policies. Using rule-based fallback.</p>
        ) : (
          <table className="w-full text-xs border-collapse">
            <thead>
              <tr className="border-b border-slate-200">
                <th className="text-left py-2 px-2 text-slate-500 uppercase">Policy</th>
                <th className="text-left py-2 px-2 text-slate-500 uppercase">Version</th>
                <th className="text-left py-2 px-2 text-slate-500 uppercase">Domain</th>
                <th className="text-left py-2 px-2 text-slate-500 uppercase">Accuracy</th>
                <th className="text-left py-2 px-2 text-slate-500 uppercase">Runs</th>
                <th className="text-left py-2 px-2 text-slate-500 uppercase">Regret</th>
              </tr>
            </thead>
            <tbody>
              {production.map((v, i) => (
                <tr key={i} className="border-b border-slate-100">
                  <td className="py-2 px-2 font-medium text-slate-700">{v.id}</td>
                  <td className="py-2 px-2 text-slate-600">{v.version}</td>
                  <td className="py-2 px-2 text-slate-600">{v.domain}</td>
                  <td className="py-2 px-2 text-slate-600">{v.production_accuracy > 0 ? `${(v.production_accuracy * 100).toFixed(0)}%` : '—'}</td>
                  <td className="py-2 px-2 text-slate-600">{v.production_runs}</td>
                  <td className={`py-2 px-2 ${v.regret_vs_rules < 0 ? 'text-emerald-600' : 'text-slate-600'}`}>{v.regret_vs_rules.toFixed(2)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </Card>

      {/* Candidates */}
      <Card title="Candidate Policies">
        {candidates.length === 0 ? (
          <p className="text-sm text-slate-500">No candidate policies in pipeline.</p>
        ) : (
          <table className="w-full text-xs border-collapse">
            <thead>
              <tr className="border-b border-slate-200">
                <th className="text-left py-2 px-2 text-slate-500 uppercase">Policy</th>
                <th className="text-left py-2 px-2 text-slate-500 uppercase">Version</th>
                <th className="text-left py-2 px-2 text-slate-500 uppercase">Stage</th>
                <th className="text-left py-2 px-2 text-slate-500 uppercase">Offline</th>
                <th className="text-left py-2 px-2 text-slate-500 uppercase">Shadow</th>
                <th className="text-left py-2 px-2 text-slate-500 uppercase">Next Gate</th>
              </tr>
            </thead>
            <tbody>
              {candidates.map((v, i) => {
                const nextGate = v.status === 'training'
                  ? `Offline ≥ 65% (current: ${(v.offline_accuracy * 100).toFixed(0)}%)`
                  : `Shadow ≥ 60% + 20 runs (current: ${v.shadow_accuracy > 0 ? (v.shadow_accuracy * 100).toFixed(0) + '%' : '—'}, ${v.production_runs} runs)`;
                return (
                  <tr key={i} className="border-b border-slate-100">
                    <td className="py-2 px-2 font-medium text-slate-700">{v.id}</td>
                    <td className="py-2 px-2 text-slate-600">{v.version}</td>
                    <td className="py-2 px-2"><StatusBadge status={v.status} /></td>
                    <td className="py-2 px-2 text-slate-600">{(v.offline_accuracy * 100).toFixed(0)}%</td>
                    <td className="py-2 px-2 text-slate-600">{v.shadow_accuracy > 0 ? `${(v.shadow_accuracy * 100).toFixed(0)}%` : '—'}</td>
                    <td className="py-2 px-2 text-slate-500">{nextGate}</td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        )}
      </Card>

      {/* Rollback History */}
      {rolledBack.length > 0 && (
        <Card title="Rollback History">
          <table className="w-full text-xs border-collapse">
            <thead>
              <tr className="border-b border-slate-200">
                <th className="text-left py-2 px-2 text-slate-500 uppercase">Policy</th>
                <th className="text-left py-2 px-2 text-slate-500 uppercase">Rolled Back From</th>
                <th className="text-left py-2 px-2 text-slate-500 uppercase">Reason</th>
              </tr>
            </thead>
            <tbody>
              {rolledBack.map((v, i) => (
                <tr key={i} className="border-b border-slate-100">
                  <td className="py-2 px-2 font-medium text-slate-700">{v.id} v{v.version}</td>
                  <td className="py-2 px-2 text-slate-600">v{v.rolled_back_from}</td>
                  <td className="py-2 px-2 text-red-600">{v.rollback_reason}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </Card>
      )}

      {/* Promotion Rules */}
      <Card title="Promotion Rules">
        <table className="w-full text-xs border-collapse">
          <thead>
            <tr className="border-b border-slate-200">
              <th className="text-left py-2 px-2 text-slate-500 uppercase">Gate</th>
              <th className="text-left py-2 px-2 text-slate-500 uppercase">Threshold</th>
            </tr>
          </thead>
          <tbody>
            <tr className="border-b border-slate-100"><td className="py-1.5 px-2 text-slate-700">Candidate → Shadow</td><td className="py-1.5 px-2 text-slate-600">Offline accuracy ≥ 65%</td></tr>
            <tr className="border-b border-slate-100"><td className="py-1.5 px-2 text-slate-700">Shadow → Active</td><td className="py-1.5 px-2 text-slate-600">Shadow accuracy ≥ 60%, 20+ runs, regret ≤ 0</td></tr>
            <tr className="border-b border-slate-100"><td className="py-1.5 px-2 text-slate-700">Block on drift</td><td className="py-1.5 px-2 text-slate-600">Yes</td></tr>
            <tr className="border-b border-slate-100"><td className="py-1.5 px-2 text-slate-700">Max safety override rate</td><td className="py-1.5 px-2 text-slate-600">5%</td></tr>
          </tbody>
        </table>
      </Card>
    </div>
  );
}

function Stage({ label, color }: { label: string; color: string }) {
  const styles: Record<string, string> = {
    amber: 'bg-amber-50 text-amber-700 border-amber-200',
    blue: 'bg-blue-50 text-blue-700 border-blue-200',
    emerald: 'bg-emerald-50 text-emerald-700 border-emerald-200',
    slate: 'bg-slate-50 text-slate-500 border-slate-200',
  };
  return <span className={`text-[10px] font-semibold px-2 py-0.5 rounded border ${styles[color]}`}>{label}</span>;
}

function StatusBadge({ status }: { status: string }) {
  const styles: Record<string, string> = {
    active: 'bg-emerald-50 text-emerald-700 border-emerald-200',
    shadow: 'bg-blue-50 text-blue-700 border-blue-200',
    training: 'bg-amber-50 text-amber-700 border-amber-200',
    retired: 'bg-slate-100 text-slate-500 border-slate-200',
  };
  return <span className={`text-[10px] font-semibold px-1.5 py-0.5 rounded border ${styles[status] || styles.retired}`}>{status}</span>;
}
