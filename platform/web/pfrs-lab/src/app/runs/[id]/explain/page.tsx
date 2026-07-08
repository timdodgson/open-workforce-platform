import { loadRunSummary, loadDiscoveries, loadWorkerLifecycles, loadTree, loadDiversity, loadPlateaus } from '@/lib/data-loader';
import RunPageShell from '@/features/runs/RunPageShell';
import ExplainRun from './ExplainRun';

export const dynamic = 'force-dynamic';

export default async function ExplainPage({ params }: { params: Promise<{ id: string }> }) {
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
        title="Explain Run"
        empty={summary.weeks.length === 0}
        emptyMessage="No telemetry available."
      >
        <ExplainRun
          runId={id}
          summary={summary}
          discoveries={discoveries}
          workers={workers}
          tree={tree}
          diversity={diversity}
          plateaus={plateaus}
        />
      </RunPageShell>
    );
  } catch (err) {
    return (
      <RunPageShell title="Explain Run" error={String(err)}>
        {null}
      </RunPageShell>
    );
  }
}
