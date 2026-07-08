import { loadRunSummary, loadImprovements, loadDiscoveries } from '@/lib/data-loader';
import RunPageShell from '@/features/runs/RunPageShell';
import SearchCharts from '@/app/search/SearchCharts';
import GlobalBestTimeline from '@/app/search/GlobalBestTimeline';
import DiscoveryTimeline from '@/app/search/DiscoveryTimeline';

export const dynamic = 'force-dynamic';

export default async function RunSearchPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;

  try {
    const [d, improvements, discoveries] = await Promise.all([
      loadRunSummary(id),
      loadImprovements(id),
      loadDiscoveries(id),
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
    const empty = chartData.length === 0 && improvements.length === 0 && discoveries.length === 0;

    return (
      <RunPageShell
        title="Search Progress"
        empty={empty}
        emptyMessage="No search data available for this run."
      >
        <div>
          {chartData.length > 0 && <SearchCharts data={chartData} />}
          {improvements.length > 0 && <GlobalBestTimeline events={improvements} />}
          {discoveries.length > 0 && <DiscoveryTimeline records={discoveries} />}
        </div>
      </RunPageShell>
    );
  } catch (err) {
    return (
      <RunPageShell title="Search Progress" error={String(err)}>
        {null}
      </RunPageShell>
    );
  }
}
