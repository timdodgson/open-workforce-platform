import { loadRunSummary, loadImprovements, loadDiscoveries } from '@/lib/data-loader';
import Card from '@/components/Card';
import SearchCharts from '@/app/search/SearchCharts';
import GlobalBestTimeline from '@/app/search/GlobalBestTimeline';
import DiscoveryTimeline from '@/app/search/DiscoveryTimeline';

export const dynamic = 'force-dynamic';

export default async function RunSearchPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;

  let d, improvements, discoveries;
  try {
    [d, improvements, discoveries] = await Promise.all([
      loadRunSummary(id),
      loadImprovements(id),
      loadDiscoveries(id),
    ]);
  } catch (err) {
    return (
      <Card title="Error">
        <p className="text-red-400 text-sm">Failed to load data: {String(err)}</p>
      </Card>
    );
  }

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
      {chartData.length > 0 && <SearchCharts data={chartData} />}
      {improvements.length > 0 && <GlobalBestTimeline events={improvements} />}
      {discoveries.length > 0 && <DiscoveryTimeline records={discoveries} />}
      {chartData.length === 0 && improvements.length === 0 && discoveries.length === 0 && (
        <Card title="No Data">
          <p className="text-gray-500 text-sm">No search data available for this run.</p>
        </Card>
      )}
    </div>
  );
}
