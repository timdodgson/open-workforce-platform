import { getStorageProvider } from '@/lib/storage';
import Card from '@/components/Card';
import type { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Counterfactual Learning',
  description: 'Policy regret analysis, counterfactual improvements, and training opportunities.',
};

export const dynamic = 'force-dynamic';

interface CounterfactualRow {
  timestamp: string;
  runId: string;
  decisionType: string;
  domain: string;
  instance: string;
  algorithm: string;
  actualAction: string;
  confidence: number;
  policyId: string;
  policyVersion: string;
  policyType: string;
  expectedValue: number;
  actualOutcome: number;
  outcomeMetric: string;
  regret: number;
  bestAlternative: string;
}

function parseCounterfactualCSV(content: string): CounterfactualRow[] {
  const lines = content.trim().split('\n');
  if (lines.length < 2) return [];
  const rows: CounterfactualRow[] = [];
  for (let i = 1; i < lines.length; i++) {
    const f = lines[i].split(',');
    if (f.length < 25) continue;
    rows.push({
      timestamp: f[0],
      runId: f[1],
      decisionType: f[2],
      domain: f[3],
      instance: f[4],
      algorithm: f[5],
      actualAction: f[6],
      confidence: parseFloat(f[7]) || 0,
      policyId: f[8],
      policyVersion: f[9],
      policyType: f[10],
      expectedValue: parseFloat(f[11]) || 0,
      actualOutcome: parseFloat(f[21]) || 0,
      outcomeMetric: f[22],
      regret: parseFloat(f[23]) || 0,
      bestAlternative: f[24],
    });
  }
  return rows;
}

export default async function CounterfactualPage() {
  const storage = getStorageProvider();
  const runIds = await storage.listRuns();

  const allRows: CounterfactualRow[] = [];
  for (const runId of runIds) {
    const content = await storage.readFile(runId, 'counterfactual_learning.csv');
    if (content) {
      allRows.push(...parseCounterfactualCSV(content));
    }
  }

  if (allRows.length === 0) {
    return (
      <Card title="Counterfactual Learning">
        <div className="border-2 border-dashed border-slate-300 rounded-lg p-8 text-center text-slate-500">
          <p className="mb-2">No counterfactual data available.</p>
          <p className="text-xs">Run experiments with <code className="text-blue-600">--worker-decision-mode adaptive</code> to generate counterfactual records.</p>
        </div>
      </Card>
    );
  }

  // Compute summary metrics.
  const totalDecisions = allRows.length;
  const totalRegret = allRows.reduce((sum, r) => sum + r.regret, 0);
  const meanRegret = totalRegret / totalDecisions;
  const positiveRegret = allRows.filter(r => r.regret > 0);
  const regretRate = (positiveRegret.length / totalDecisions) * 100;
  const trainingOpportunities = positiveRegret.length;

  // Group by domain.
  const byDomain = new Map<string, CounterfactualRow[]>();
  for (const row of allRows) {
    const existing = byDomain.get(row.domain) || [];
    existing.push(row);
    byDomain.set(row.domain, existing);
  }

  // Top regret decisions.
  const topRegret = [...allRows]
    .sort((a, b) => b.regret - a.regret)
    .slice(0, 10);

  return (
    <div className="space-y-6">
      <Card title="Counterfactual Learning">
        <p className="text-sm text-slate-600 mb-4">
          Every policy decision records what was done and what alternatives existed.
          Regret measures the gap between actual outcome and the best alternative.
          Positive regret = training opportunity.
        </p>

        {/* Summary metrics */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
          <MetricBox label="Total Decisions" value={totalDecisions.toLocaleString()} />
          <MetricBox label="Mean Regret" value={meanRegret.toFixed(2)} />
          <MetricBox label="Regret Rate" value={`${regretRate.toFixed(1)}%`} />
          <MetricBox label="Training Opportunities" value={trainingOpportunities.toLocaleString()} />
        </div>
      </Card>

      {/* Per-domain breakdown */}
      <Card title="Policy Regret by Domain">
        <table className="w-full text-sm border-collapse">
          <thead>
            <tr className="border-b border-slate-200">
              <th className="text-left py-2 px-3 text-xs font-semibold text-slate-500 uppercase">Domain</th>
              <th className="text-left py-2 px-3 text-xs font-semibold text-slate-500 uppercase">Decisions</th>
              <th className="text-left py-2 px-3 text-xs font-semibold text-slate-500 uppercase">Total Regret</th>
              <th className="text-left py-2 px-3 text-xs font-semibold text-slate-500 uppercase">Mean Regret</th>
              <th className="text-left py-2 px-3 text-xs font-semibold text-slate-500 uppercase">Regret Rate</th>
            </tr>
          </thead>
          <tbody>
            {Array.from(byDomain.entries()).map(([domain, rows]) => {
              const domainRegret = rows.reduce((s, r) => s + r.regret, 0);
              const domainMean = domainRegret / rows.length;
              const domainRate = (rows.filter(r => r.regret > 0).length / rows.length) * 100;
              return (
                <tr key={domain} className="border-b border-slate-100">
                  <td className="py-2 px-3 font-semibold text-slate-700">{domain.toUpperCase()}</td>
                  <td className="py-2 px-3 text-slate-600">{rows.length}</td>
                  <td className="py-2 px-3 text-slate-600">{domainRegret.toFixed(1)}</td>
                  <td className="py-2 px-3 text-slate-600">{domainMean.toFixed(2)}</td>
                  <td className="py-2 px-3 text-slate-600">{domainRate.toFixed(1)}%</td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </Card>

      {/* Top regret decisions (training opportunities) */}
      {topRegret.length > 0 && (
        <Card title="Top Training Opportunities">
          <p className="text-xs text-slate-500 mb-3">
            Decisions with highest regret — strongest signal for policy improvement.
          </p>
          <table className="w-full text-xs border-collapse">
            <thead>
              <tr className="border-b border-slate-200">
                <th className="text-left py-2 px-2 text-slate-500 uppercase">Run</th>
                <th className="text-left py-2 px-2 text-slate-500 uppercase">Domain</th>
                <th className="text-left py-2 px-2 text-slate-500 uppercase">Action</th>
                <th className="text-left py-2 px-2 text-slate-500 uppercase">Better</th>
                <th className="text-left py-2 px-2 text-slate-500 uppercase">Regret</th>
                <th className="text-left py-2 px-2 text-slate-500 uppercase">Policy</th>
              </tr>
            </thead>
            <tbody>
              {topRegret.map((row, i) => (
                <tr key={i} className="border-b border-slate-50">
                  <td className="py-1.5 px-2 text-slate-600 truncate max-w-[150px]">{row.runId}</td>
                  <td className="py-1.5 px-2 text-slate-600">{row.domain}</td>
                  <td className="py-1.5 px-2 text-slate-700 font-medium">{row.actualAction}</td>
                  <td className="py-1.5 px-2 text-emerald-600">{row.bestAlternative || '—'}</td>
                  <td className="py-1.5 px-2 text-red-600 font-medium">{row.regret.toFixed(1)}</td>
                  <td className="py-1.5 px-2 text-slate-500">{row.policyId} v{row.policyVersion}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </Card>
      )}
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
