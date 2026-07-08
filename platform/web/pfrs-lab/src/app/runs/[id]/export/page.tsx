import { loadRunSummary, loadTree, loadDiversity, loadDiscoveries, loadWorkerLifecycles } from '@/lib/data-loader';
import RunPageShell from '@/features/runs/RunPageShell';
import PublicationExport from './PublicationExport';

export const dynamic = 'force-dynamic';

export default async function ExportPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;

  try {
    const [summary, tree, diversity, discoveries, workers] = await Promise.all([
      loadRunSummary(id),
      loadTree(id),
      loadDiversity(id),
      loadDiscoveries(id),
      loadWorkerLifecycles(id),
    ]);
    return (
      <RunPageShell title="Publication Export">
        <PublicationExport
          runId={id}
          summary={summary}
          tree={tree}
          diversity={diversity}
          discoveries={discoveries}
          workers={workers}
        />
      </RunPageShell>
    );
  } catch (err) {
    return (
      <RunPageShell title="Publication Export" error={String(err)}>
        {null}
      </RunPageShell>
    );
  }
}
