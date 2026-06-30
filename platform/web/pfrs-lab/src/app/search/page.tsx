import { loadRunSummary, loadImprovements, loadDiscoveries } from '@/lib/data-loader';
import Card from '@/components/Card';
import SearchCharts from './SearchCharts';
import GlobalBestTimeline from './GlobalBestTimeline';
import DiscoveryTimeline from './DiscoveryTimeline';

function fmt(n: number): string {
  return n.toLocaleString('en-US');
}

export default async function SearchPage() {
  const [d, improvements, discoveries] = await Promise.all([
    loadRunSummary(),
    loadImprovements(),
    loadDiscoveries(),
  ]);

  const chartData = d.weeks.map((w, i) => ({
    week: `W${w.week}`,
    penalty: w.finalPenalty,
    cumulative: d.cumulativePenalties[i],
    contribution: d.totalPenalty > 0 ? +(w.finalPenalty / d.totalPenalty * 100).toFixed(1) : 0,
    efficiencyPerM: w.candidates > 0 ? +(w.improvement / (w.candidates / 1_000_000)).toFixed(1) : 0,
    workers: w.workersStarted,
    candidates: w.candidates,
  }));

  return (
    <div>
      <SearchCharts data={chartData} />

      <GlobalBestTimeline events={improvements} />

      {/* Discovery Timeline — the core instrumentation section */}
      <DiscoveryTimeline records={discoveries} />

      <Card title="Workers vs Candidates">
        <table className="w-full text-xs">
          <thead>
            <tr className="text-gray-500 uppercase">
              <th className="text-left p-2">Week</th>
              <th className="text-right p-2">Workers</th>
              <th className="text-right p-2">Candidates</th>
              <th className="text-right p-2">Cands/Worker</th>
              <th className="text-right p-2">Penalty</th>
              <th className="text-right p-2">Improvement</th>
              <th className="text-right p-2">Eff (pen/M cands)</th>
            </tr>
          </thead>
          <tbody>
            {d.weeks.map(w => (
              <tr key={w.week} className="border-t border-gray-800">
                <td className="p-2">{w.week}</td>
                <td className="text-right p-2">{w.workersStarted}</td>
                <td className="text-right p-2">{fmt(w.candidates)}</td>
                <td className="text-right p-2">{w.workersStarted > 0 ? fmt(Math.round(w.candidates / w.workersStarted)) : '—'}</td>
                <td className="text-right p-2">{fmt(w.finalPenalty)}</td>
                <td className="text-right p-2">{w.improvement}</td>
                <td className="text-right p-2">{w.candidates > 0 ? (w.improvement / (w.candidates / 1_000_000)).toFixed(1) : '—'}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </Card>
    </div>
  );
}
