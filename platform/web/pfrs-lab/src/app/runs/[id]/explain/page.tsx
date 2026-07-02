import { loadRunSummary, loadDiscoveries, loadWorkerLifecycles, loadTree, loadDiversity, loadPlateaus } from '@/lib/data-loader';
import Card from '@/components/Card';
import ExplainRun from './ExplainRun';

export const dynamic = 'force-dynamic';

export default async function ExplainPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;

  let summary, discoveries, workers, tree, diversity, plateaus;
  try {
    [summary, discoveries, workers, tree, diversity, plateaus] = await Promise.all([
      loadRunSummary(id),
      loadDiscoveries(id),
      loadWorkerLifecycles(id),
      loadTree(id),
      loadDiversity(id),
      loadPlateaus(id),
    ]);
  } catch (err) {
    return (
      <Card title="Error">
        <p className="text-red-400 text-sm">Failed to load data: {String(err)}</p>
      </Card>
    );
  }

  if (summary.weeks.length === 0) {
    return (
      <Card title="Explain Run">
        <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500">
          <p>No telemetry available.</p>
        </div>
      </Card>
    );
  }

  return (
    <ExplainRun runId={id} summary={summary} discoveries={discoveries}
      workers={workers} tree={tree} diversity={diversity} plateaus={plateaus} />
  );
}
