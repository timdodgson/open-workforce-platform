import { getStorageProvider } from '@/lib/storage';
import Card from '@/components/Card';
import type { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Policy Validation',
  description: 'Search Intelligence 2.0 validation status and results.',
};

export const dynamic = 'force-dynamic';

export default async function ValidationPage() {
  const storage = getStorageProvider();

  // Check if validation results exist.
  const runIds = await storage.listRuns();
  const validationRuns = runIds.filter(id => id.startsWith('si2-'));
  const totalExpected = 240; // 8 configs × 3 modes × 10 seeds
  const completed = validationRuns.length;
  const pending = totalExpected - completed;

  const status = completed === 0 ? 'pending' : completed >= totalExpected ? 'completed' : 'running';

  return (
    <div className="space-y-6">
      <Card title="SI 2.0 Validation">
        <p className="text-sm text-slate-600 mb-4">
          Validation compares Rules, Hybrid, and Learned policies across all 4 domains
          with 10 seeds per configuration (240 total experiments).
        </p>

        {/* Status */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
          <StatusBox label="Status" value={status.toUpperCase()} status={status} />
          <MetricBox label="Completed" value={`${completed} / ${totalExpected}`} />
          <MetricBox label="Pending" value={pending.toString()} />
          <MetricBox label="Policy Version" value="2.0.0" />
        </div>

        {completed === 0 && (
          <div className="border-2 border-dashed border-slate-300 rounded-lg p-6 text-center">
            <p className="text-sm text-slate-600 mb-2">No validation experiments have been run yet.</p>
            <p className="text-xs text-slate-400 mb-4">Execute the validation suite to populate results.</p>
            <code className="block text-xs bg-slate-50 border border-slate-200 rounded p-3 text-slate-700">
              go run ./cmd/owp validate-si2 --output validation/si2
            </code>
          </div>
        )}
      </Card>

      {/* Experiment Matrix */}
      <Card title="Experiment Matrix">
        <table className="w-full text-xs border-collapse">
          <thead>
            <tr className="border-b border-slate-200">
              <th className="text-left py-2 px-2 text-slate-500 uppercase">Domain</th>
              <th className="text-left py-2 px-2 text-slate-500 uppercase">Instance</th>
              <th className="text-left py-2 px-2 text-slate-500 uppercase">Algorithm</th>
              <th className="text-left py-2 px-2 text-slate-500 uppercase">Rules</th>
              <th className="text-left py-2 px-2 text-slate-500 uppercase">Hybrid</th>
              <th className="text-left py-2 px-2 text-slate-500 uppercase">Learned</th>
            </tr>
          </thead>
          <tbody>
            {[
              { domain: 'CVRP', instance: 'A-n32-k5', algorithm: 'SA' },
              { domain: 'CVRP', instance: 'A-n32-k5', algorithm: 'Portfolio' },
              { domain: 'JSS', instance: 'la01', algorithm: 'Tabu' },
              { domain: 'JSS', instance: 'la01', algorithm: 'Portfolio' },
              { domain: 'VRPTW', instance: 'C101', algorithm: 'SA' },
              { domain: 'VRPTW', instance: 'C101', algorithm: 'Portfolio' },
              { domain: 'NRP', instance: 'n012w8', algorithm: 'SA' },
              { domain: 'NRP', instance: 'n012w8', algorithm: 'Portfolio' },
            ].map((row, i) => {
              const prefix = `si2-${row.domain.toLowerCase()}-${row.instance}-${row.algorithm.toLowerCase()}`;
              const rulesCount = validationRuns.filter(id => id.startsWith(`${prefix}-rules`)).length;
              const hybridCount = validationRuns.filter(id => id.startsWith(`${prefix}-hybrid`)).length;
              const learnedCount = validationRuns.filter(id => id.startsWith(`${prefix}-learned`)).length;
              return (
                <tr key={i} className="border-b border-slate-100">
                  <td className="py-1.5 px-2 font-medium text-slate-700">{row.domain}</td>
                  <td className="py-1.5 px-2 text-slate-600">{row.instance}</td>
                  <td className="py-1.5 px-2 text-slate-600">{row.algorithm}</td>
                  <td className="py-1.5 px-2"><RunCount count={rulesCount} target={10} /></td>
                  <td className="py-1.5 px-2"><RunCount count={hybridCount} target={10} /></td>
                  <td className="py-1.5 px-2"><RunCount count={learnedCount} target={10} /></td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </Card>

      {/* How to reproduce */}
      <Card title="Reproducibility">
        <p className="text-xs text-slate-500 mb-3">
          One command executes the complete validation suite:
        </p>
        <code className="block text-xs bg-slate-50 border border-slate-200 rounded p-3 text-slate-700 mb-3">
          go run ./cmd/owp validate-si2 --output validation/si2
        </code>
        <p className="text-xs text-slate-500">
          Seeds: 42, 123, 555, 777, 999, 1001, 2022, 3033, 4044, 5055
        </p>
      </Card>
    </div>
  );
}

function MetricBox({ label, value }: { label: string; value: string }) {
  return (
    <div className="border border-slate-200 rounded-lg p-3">
      <p className="text-xs text-slate-500 uppercase tracking-wider">{label}</p>
      <p className="text-lg font-bold text-slate-800 mt-1">{value}</p>
    </div>
  );
}

function StatusBox({ label, value, status }: { label: string; value: string; status: string }) {
  const colors: Record<string, string> = {
    pending: 'border-amber-200 bg-amber-50',
    running: 'border-blue-200 bg-blue-50',
    completed: 'border-emerald-200 bg-emerald-50',
  };
  const textColors: Record<string, string> = {
    pending: 'text-amber-700',
    running: 'text-blue-700',
    completed: 'text-emerald-700',
  };
  return (
    <div className={`border rounded-lg p-3 ${colors[status] || ''}`}>
      <p className="text-xs text-slate-500 uppercase tracking-wider">{label}</p>
      <p className={`text-lg font-bold mt-1 ${textColors[status] || 'text-slate-800'}`}>{value}</p>
    </div>
  );
}

function RunCount({ count, target }: { count: number; target: number }) {
  if (count === 0) return <span className="text-slate-400">—</span>;
  if (count >= target) return <span className="text-emerald-600 font-medium">{count}/{target} ✓</span>;
  return <span className="text-blue-600">{count}/{target}</span>;
}
