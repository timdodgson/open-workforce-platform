import { loadRunSummary, loadTree, loadDiversity, loadDiscoveries, loadWorkerLifecycles } from '@/lib/data-loader';
import Card from '@/components/Card';
import PublicationExport from './PublicationExport';

export const dynamic = 'force-dynamic';

export default async function ExportPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;

  let summary, tree, diversity, discoveries, workers;
  try {
    [summary, tree, diversity, discoveries, workers] = await Promise.all([
      loadRunSummary(id),
      loadTree(id),
      loadDiversity(id),
      loadDiscoveries(id),
      loadWorkerLifecycles(id),
    ]);
  } catch (err) {
    return (
      <Card title="Error">
        <p className="text-red-400 text-sm">Failed to load data: {String(err)}</p>
      </Card>
    );
  }

  return (
    <PublicationExport
      runId={id} summary={summary} tree={tree}
      diversity={diversity} discoveries={discoveries} workers={workers}
    />
  );
}
