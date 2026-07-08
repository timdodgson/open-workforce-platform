'use client';

import Card from '@/components/Card';
import type { CounterfactualSummary } from '@/lib/types/intelligence';

export default function CounterfactualTab({ summary }: { summary: CounterfactualSummary | null }) {
  if (!summary || summary.totalDecisions === 0) {
    return (
      <Card title="Counterfactual Learning">
        <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500 text-xs">
          No counterfactual data. Run with <code className="text-blue-400">--worker-decision-mode adaptive</code> to generate records.
        </div>
      </Card>
    );
  }

  return (
    <div className="space-y-4">
      <Card title="Counterfactual Learning">
        <p className="text-xs text-gray-400 mb-4">
          Regret measures the gap between actual outcome and the best alternative. Positive regret = training opportunity.
        </p>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
          <Metric label="Decisions" value={summary.totalDecisions.toLocaleString()} />
          <Metric label="Mean Regret" value={summary.meanRegret.toFixed(2)} />
          <Metric label="Regret Rate" value={`${summary.regretRate.toFixed(1)}%`} />
          <Metric label="Training Ops" value={summary.trainingOpportunities.toLocaleString()} />
        </div>
      </Card>

      {summary.byDomain.length > 0 && (
        <Card title="Regret by Domain">
          <table className="w-full text-[10px]">
            <thead>
              <tr className="text-gray-500 uppercase border-b border-gray-700">
                <th className="text-left p-1.5">Domain</th>
                <th className="text-right p-1.5">Decisions</th>
                <th className="text-right p-1.5">Mean Regret</th>
                <th className="text-right p-1.5">Rate</th>
              </tr>
            </thead>
            <tbody>
              {summary.byDomain.map((d) => (
                <tr key={d.domain} className="border-b border-gray-800">
                  <td className="p-1.5 font-medium text-gray-300">{d.domain.toUpperCase()}</td>
                  <td className="p-1.5 text-right">{d.decisions}</td>
                  <td className="p-1.5 text-right">{d.meanRegret.toFixed(2)}</td>
                  <td className="p-1.5 text-right">{d.regretRate.toFixed(1)}%</td>
                </tr>
              ))}
            </tbody>
          </table>
        </Card>
      )}

      {summary.topRegret.length > 0 && (
        <Card title="Top Training Opportunities">
          <table className="w-full text-[10px]">
            <thead>
              <tr className="text-gray-500 uppercase border-b border-gray-700">
                <th className="text-left p-1.5">Run</th>
                <th className="text-left p-1.5">Action</th>
                <th className="text-left p-1.5">Better</th>
                <th className="text-right p-1.5">Regret</th>
              </tr>
            </thead>
            <tbody>
              {summary.topRegret.map((row, i) => (
                <tr key={i} className="border-b border-gray-800">
                  <td className="p-1.5 text-blue-400 truncate max-w-[120px]">{row.runId}</td>
                  <td className="p-1.5">{row.actualAction}</td>
                  <td className="p-1.5 text-emerald-400">{row.bestAlternative || '—'}</td>
                  <td className="p-1.5 text-right text-red-400">{row.regret.toFixed(1)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </Card>
      )}
    </div>
  );
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div className="bg-gray-800 rounded p-3">
      <p className="text-[9px] text-gray-500 uppercase">{label}</p>
      <p className="text-lg font-bold text-gray-200 mt-1">{value}</p>
    </div>
  );
}
