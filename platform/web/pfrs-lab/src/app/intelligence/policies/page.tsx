import { getStorageProvider } from '@/lib/storage';
import Card from '@/components/Card';
import type { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Policy Dashboard',
  description: 'Policy performance, versioning, calibration, regret, and drift analysis.',
};

// Do not ISR/prerender — S3 reads exceed OpenNext build timeouts.
export const dynamic = 'force-dynamic';

// --- Types ---

interface PolicyVersion {
  id: string;
  version: string;
  domain: string;
  decision_type: string;
  algorithm: string;
  status: string;
  created_at: string;
  promoted_at?: string;
  retired_at?: string;
  training_samples: number;
  training_date: string;
  features: string[];
  offline_accuracy: number;
  shadow_accuracy: number;
  production_accuracy: number;
  production_runs: number;
  regret_vs_rules: number;
  drift_detected: boolean;
  model_path: string;
  rolled_back_from?: string;
  rollback_reason?: string;
}

interface EvalRow {
  domain: string;
  policyId: string;
  policyVersion: string;
  policyType: string;
  confidence: number;
  correct: boolean;
  regret: number;
  predictionError: number;
}

// --- Parsers ---

function parseEvalCSV(content: string): EvalRow[] {
  const lines = content.trim().split('\n');
  if (lines.length < 2) return [];
  const rows: EvalRow[] = [];
  for (let i = 1; i < lines.length; i++) {
    const f = lines[i].split(',');
    if (f.length < 18) continue;
    rows.push({
      domain: f[3],
      policyId: f[6],
      policyVersion: f[7],
      policyType: f[8],
      confidence: parseFloat(f[10]) || 0,
      correct: f[14] === '1',
      regret: parseFloat(f[17]) || 0,
      predictionError: parseFloat(f[13]) || 0,
    });
  }
  return rows;
}

// --- Page ---

export default async function PolicyDashboardPage() {
  const storage = getStorageProvider();

  // Load registry.
  const registryContent = await storage.readRootFile('policy_registry.json');
  let versions: PolicyVersion[] = [];
  if (registryContent) {
    try {
      const parsed = JSON.parse(registryContent);
      versions = parsed.versions || [];
    } catch { /* graceful */ }
  }

  // Load evaluation data from all runs.
  const runIds = await storage.listRuns();
  const evalRows: EvalRow[] = [];
  for (const runId of runIds) {
    const content = await storage.readFile(runId, 'policy_evaluation.csv');
    if (content) evalRows.push(...parseEvalCSV(content));
  }

  const activeVersions = versions.filter(v => v.status === 'active');
  const hasData = versions.length > 0 || evalRows.length > 0;

  if (!hasData) {
    return (
      <Card title="Policy Dashboard">
        <div className="border-2 border-dashed border-slate-300 rounded-lg p-8 text-center text-slate-500">
          <p className="mb-2">No policy data available.</p>
          <p className="text-xs">Run experiments with <code className="text-blue-600">--worker-decision-mode adaptive</code> and a policy directory to generate data.</p>
        </div>
      </Card>
    );
  }

  // Compute metrics from eval rows.
  const totalDecisions = evalRows.length;
  const correctCount = evalRows.filter(r => r.correct).length;
  const accuracy = totalDecisions > 0 ? correctCount / totalDecisions : 0;
  const totalRegret = evalRows.reduce((s, r) => s + r.regret, 0);
  const meanRegret = totalDecisions > 0 ? totalRegret / totalDecisions : 0;

  // Fallback rate.
  const fallbackCount = evalRows.filter(r => r.policyType === 'rule').length;
  const learnedCount = evalRows.filter(r => r.policyType === 'learned' || r.policyType === 'hybrid').length;
  const fallbackRate = totalDecisions > 0 ? (fallbackCount / totalDecisions) * 100 : 0;

  // Calibration buckets.
  const calibration = Array.from({ length: 10 }, (_, i) => {
    const lo = i * 0.1;
    const hi = lo + 0.1;
    const bucket = evalRows.filter(r => r.confidence >= lo && r.confidence < hi);
    const bucketCorrect = bucket.filter(r => r.correct).length;
    return {
      range: `${(lo * 100).toFixed(0)}–${(hi * 100).toFixed(0)}%`,
      total: bucket.length,
      correct: bucketCorrect,
      actual: bucket.length > 0 ? bucketCorrect / bucket.length : 0,
      expected: lo + 0.05,
    };
  }).filter(b => b.total > 0);

  // Per-domain metrics.
  const domains = [...new Set(evalRows.map(r => r.domain))];
  const domainMetrics = domains.map(d => {
    const rows = evalRows.filter(r => r.domain === d);
    const correct = rows.filter(r => r.correct).length;
    const regret = rows.reduce((s, r) => s + r.regret, 0);
    return {
      domain: d,
      decisions: rows.length,
      accuracy: rows.length > 0 ? correct / rows.length : 0,
      regret,
      meanRegret: rows.length > 0 ? regret / rows.length : 0,
    };
  });

  return (
    <div className="space-y-6">
      {/* Current Production Policies */}
      <Card title="Active Policies">
        {activeVersions.length === 0 ? (
          <p className="text-sm text-slate-500">No active policies. All decisions use rule-based fallback.</p>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {activeVersions.map(v => (
              <div key={`${v.id}-${v.version}`} className="border border-slate-200 rounded-lg p-4">
                <div className="flex items-baseline justify-between mb-2">
                  <span className="text-sm font-semibold text-slate-800">{v.id}</span>
                  <span className="text-xs text-emerald-600 font-medium bg-emerald-50 px-2 py-0.5 rounded">{v.status}</span>
                </div>
                <div className="text-xs text-slate-500 space-y-1">
                  <p>Version: <span className="text-slate-700 font-medium">{v.version}</span></p>
                  <p>Domain: <span className="text-slate-700">{v.domain}</span> · Type: <span className="text-slate-700">{v.decision_type}</span></p>
                  <p>Trained on: <span className="text-slate-700">{v.training_samples} samples</span></p>
                  <p>Accuracy: <span className="text-slate-700">{v.production_accuracy > 0 ? `${(v.production_accuracy * 100).toFixed(1)}%` : 'pending'}</span></p>
                  <p>Runs: <span className="text-slate-700">{v.production_runs}</span></p>
                  {v.drift_detected && <p className="text-amber-600 font-medium">⚠ Drift detected</p>}
                </div>
              </div>
            ))}
          </div>
        )}
      </Card>

      {/* Policy Confidence & Accuracy */}
      {totalDecisions > 0 && (
        <Card title="Policy Performance">
          <div className="grid grid-cols-2 md:grid-cols-5 gap-4 mb-4">
            <MetricBox label="Decisions" value={totalDecisions.toLocaleString()} />
            <MetricBox label="Accuracy" value={`${(accuracy * 100).toFixed(1)}%`} />
            <MetricBox label="Mean Regret" value={meanRegret.toFixed(2)} />
            <MetricBox label="Fallback Rate" value={`${fallbackRate.toFixed(1)}%`} />
            <MetricBox label="Learned Used" value={learnedCount.toLocaleString()} />
          </div>
        </Card>
      )}

      {/* Calibration */}
      {calibration.length > 0 && (
        <Card title="Confidence Calibration">
          <p className="text-xs text-slate-500 mb-3">
            Does confidence match reality? Perfect calibration: 80% confidence → 80% correct.
          </p>
          <table className="w-full text-xs border-collapse">
            <thead>
              <tr className="border-b border-slate-200">
                <th className="text-left py-2 px-3 text-slate-500 uppercase">Confidence</th>
                <th className="text-left py-2 px-3 text-slate-500 uppercase">Decisions</th>
                <th className="text-left py-2 px-3 text-slate-500 uppercase">Correct</th>
                <th className="text-left py-2 px-3 text-slate-500 uppercase">Actual %</th>
                <th className="text-left py-2 px-3 text-slate-500 uppercase">Expected %</th>
                <th className="text-left py-2 px-3 text-slate-500 uppercase">Error</th>
              </tr>
            </thead>
            <tbody>
              {calibration.map(b => (
                <tr key={b.range} className="border-b border-slate-100">
                  <td className="py-1.5 px-3 text-slate-700 font-medium">{b.range}</td>
                  <td className="py-1.5 px-3 text-slate-600">{b.total}</td>
                  <td className="py-1.5 px-3 text-slate-600">{b.correct}</td>
                  <td className="py-1.5 px-3 text-slate-600">{(b.actual * 100).toFixed(0)}%</td>
                  <td className="py-1.5 px-3 text-slate-600">{(b.expected * 100).toFixed(0)}%</td>
                  <td className={`py-1.5 px-3 font-medium ${Math.abs(b.actual - b.expected) > 0.15 ? 'text-red-600' : 'text-slate-600'}`}>
                    {(Math.abs(b.actual - b.expected) * 100).toFixed(1)}%
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </Card>
      )}

      {/* Per-Domain Performance */}
      {domainMetrics.length > 0 && (
        <Card title="Per-Domain Performance">
          <table className="w-full text-sm border-collapse">
            <thead>
              <tr className="border-b border-slate-200">
                <th className="text-left py-2 px-3 text-xs text-slate-500 uppercase">Domain</th>
                <th className="text-left py-2 px-3 text-xs text-slate-500 uppercase">Decisions</th>
                <th className="text-left py-2 px-3 text-xs text-slate-500 uppercase">Accuracy</th>
                <th className="text-left py-2 px-3 text-xs text-slate-500 uppercase">Total Regret</th>
                <th className="text-left py-2 px-3 text-xs text-slate-500 uppercase">Mean Regret</th>
              </tr>
            </thead>
            <tbody>
              {domainMetrics.map(d => (
                <tr key={d.domain} className="border-b border-slate-100">
                  <td className="py-2 px-3 font-semibold text-slate-700">{d.domain.toUpperCase()}</td>
                  <td className="py-2 px-3 text-slate-600">{d.decisions}</td>
                  <td className="py-2 px-3 text-slate-600">{(d.accuracy * 100).toFixed(1)}%</td>
                  <td className="py-2 px-3 text-slate-600">{d.regret.toFixed(1)}</td>
                  <td className="py-2 px-3 text-slate-600">{d.meanRegret.toFixed(2)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </Card>
      )}

      {/* Version History */}
      {versions.length > 0 && (
        <Card title="Version History">
          <table className="w-full text-xs border-collapse">
            <thead>
              <tr className="border-b border-slate-200">
                <th className="text-left py-2 px-2 text-slate-500 uppercase">Policy</th>
                <th className="text-left py-2 px-2 text-slate-500 uppercase">Version</th>
                <th className="text-left py-2 px-2 text-slate-500 uppercase">Domain</th>
                <th className="text-left py-2 px-2 text-slate-500 uppercase">Status</th>
                <th className="text-left py-2 px-2 text-slate-500 uppercase">Samples</th>
                <th className="text-left py-2 px-2 text-slate-500 uppercase">Offline</th>
                <th className="text-left py-2 px-2 text-slate-500 uppercase">Shadow</th>
                <th className="text-left py-2 px-2 text-slate-500 uppercase">Production</th>
                <th className="text-left py-2 px-2 text-slate-500 uppercase">Regret</th>
              </tr>
            </thead>
            <tbody>
              {versions.map((v, i) => (
                <tr key={i} className="border-b border-slate-50">
                  <td className="py-1.5 px-2 text-slate-700 font-medium">{v.id}</td>
                  <td className="py-1.5 px-2 text-slate-600">{v.version}</td>
                  <td className="py-1.5 px-2 text-slate-600">{v.domain}</td>
                  <td className="py-1.5 px-2">
                    <StatusBadge status={v.status} />
                  </td>
                  <td className="py-1.5 px-2 text-slate-600">{v.training_samples}</td>
                  <td className="py-1.5 px-2 text-slate-600">{v.offline_accuracy > 0 ? `${(v.offline_accuracy * 100).toFixed(0)}%` : '—'}</td>
                  <td className="py-1.5 px-2 text-slate-600">{v.shadow_accuracy > 0 ? `${(v.shadow_accuracy * 100).toFixed(0)}%` : '—'}</td>
                  <td className="py-1.5 px-2 text-slate-600">{v.production_accuracy > 0 ? `${(v.production_accuracy * 100).toFixed(0)}%` : '—'}</td>
                  <td className={`py-1.5 px-2 ${v.regret_vs_rules < 0 ? 'text-emerald-600' : v.regret_vs_rules > 0 ? 'text-red-600' : 'text-slate-500'}`}>
                    {v.regret_vs_rules !== 0 ? v.regret_vs_rules.toFixed(1) : '—'}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </Card>
      )}

      {/* Training History */}
      {versions.length > 0 && (
        <Card title="Training History">
          <div className="space-y-3">
            {versions.filter(v => v.features && v.features.length > 0).map((v, i) => (
              <div key={i} className="border border-slate-100 rounded-lg p-3">
                <div className="flex items-baseline gap-2 mb-1">
                  <span className="text-xs font-semibold text-slate-700">{v.id} v{v.version}</span>
                  <span className="text-xs text-slate-400">{v.training_date ? new Date(v.training_date).toLocaleDateString() : '—'}</span>
                </div>
                <p className="text-xs text-slate-500">
                  {v.training_samples} samples · {v.features.length} features: <span className="text-slate-600">{v.features.join(', ')}</span>
                </p>
              </div>
            ))}
          </div>
        </Card>
      )}
    </div>
  );
}

// --- Components ---

function MetricBox({ label, value }: { label: string; value: string }) {
  return (
    <div className="border border-slate-200 rounded-lg p-3">
      <p className="text-xs text-slate-500 uppercase tracking-wider">{label}</p>
      <p className="text-lg font-bold text-slate-800 mt-1">{value}</p>
    </div>
  );
}

function StatusBadge({ status }: { status: string }) {
  const styles: Record<string, string> = {
    active: 'bg-emerald-50 text-emerald-700 border-emerald-200',
    shadow: 'bg-blue-50 text-blue-700 border-blue-200',
    training: 'bg-amber-50 text-amber-700 border-amber-200',
    retired: 'bg-slate-100 text-slate-500 border-slate-200',
  };
  const cls = styles[status] || styles.retired;
  return <span className={`text-[10px] font-semibold px-1.5 py-0.5 rounded border ${cls}`}>{status}</span>;
}
