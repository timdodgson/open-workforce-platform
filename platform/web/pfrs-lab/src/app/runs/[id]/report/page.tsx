import { loadRunSummary, loadDiscoveries, loadWorkerLifecycles, loadTree, loadDiversity, loadPlateaus } from '@/lib/data-loader';
import RunPageShell from '@/features/runs/RunPageShell';
import QualityReport from './QualityReport';

export const dynamic = 'force-dynamic';

export default async function ReportPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;

  try {
    const [summary, discoveries, workers, tree, diversity, plateaus] = await Promise.all([
      loadRunSummary(id),
      loadDiscoveries(id),
      loadWorkerLifecycles(id),
      loadTree(id),
      loadDiversity(id),
      loadPlateaus(id),
    ]);

    return (
      <RunPageShell
        title="Quality Report"
        empty={summary.weeks.length === 0}
        emptyMessage="No data available for report generation."
      >
        <QualityReport
          summary={summary}
          discoveries={discoveries}
          workers={workers}
          tree={tree}
          diversity={diversity}
          plateaus={plateaus}
          runId={id}
        />
      </RunPageShell>
    );
  } catch (err) {
    return (
      <RunPageShell title="Quality Report" error={String(err)}>
        {null}
      </RunPageShell>
    );
  }
}
