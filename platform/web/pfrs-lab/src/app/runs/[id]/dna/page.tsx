import { loadRunSummary, loadDiscoveries, loadWorkerLifecycles, loadTree, loadDiversity } from '@/lib/data-loader';
import RunPageShell from '@/features/runs/RunPageShell';
import SearchDNA from './SearchDNA';

export const dynamic = 'force-dynamic';

export default async function DNAPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;

  try {
    const [summary, discoveries, workers, tree, diversity] = await Promise.all([
      loadRunSummary(id),
      loadDiscoveries(id),
      loadWorkerLifecycles(id),
      loadTree(id),
      loadDiversity(id),
    ]);
    return (
      <RunPageShell
        title="Search DNA"
        empty={summary.weeks.length === 0}
        emptyMessage="No telemetry available for DNA analysis."
      >
        <SearchDNA
          summary={summary}
          discoveries={discoveries}
          workers={workers}
          tree={tree}
          diversity={diversity}
        />
      </RunPageShell>
    );
  } catch (err) {
    return (
      <RunPageShell title="Search DNA" error={String(err)}>
        {null}
      </RunPageShell>
    );
  }
}
