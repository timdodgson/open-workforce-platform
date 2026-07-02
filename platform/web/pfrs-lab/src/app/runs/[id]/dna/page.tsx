import { loadRunSummary, loadDiscoveries, loadWorkerLifecycles, loadTree, loadDiversity } from '@/lib/data-loader';
import Card from '@/components/Card';
import SearchDNA from './SearchDNA';

export const dynamic = 'force-dynamic';

export default async function DNAPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;

  let summary, discoveries, workers, tree, diversity;
  try {
    [summary, discoveries, workers, tree, diversity] = await Promise.all([
      loadRunSummary(id),
      loadDiscoveries(id),
      loadWorkerLifecycles(id),
      loadTree(id),
      loadDiversity(id),
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
      <Card title="Search DNA">
        <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500">
          <p>No telemetry available for DNA analysis.</p>
        </div>
      </Card>
    );
  }

  return (
    <SearchDNA
      summary={summary}
      discoveries={discoveries}
      workers={workers}
      tree={tree}
      diversity={diversity}
    />
  );
}
