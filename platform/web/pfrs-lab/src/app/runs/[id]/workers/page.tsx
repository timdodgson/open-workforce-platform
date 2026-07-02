import { loadWorkerLifecycles, loadDiscoveries } from '@/lib/data-loader';
import Card from '@/components/Card';
import WorkerAnalysis from './WorkerAnalysis';

export const dynamic = 'force-dynamic';

export default async function WorkersPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;

  let workers, discoveries;
  try {
    [workers, discoveries] = await Promise.all([
      loadWorkerLifecycles(id),
      loadDiscoveries(id),
    ]);
  } catch (err) {
    return (
      <Card title="Error">
        <p className="text-red-400 text-sm">Failed to load data: {String(err)}</p>
      </Card>
    );
  }

  if (workers.length === 0) {
    return (
      <Card title="Worker Analysis">
        <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500">
          <p className="mb-2">No worker data available.</p>
          <p className="text-xs">Run a PFRS beam search to generate workers.csv.</p>
        </div>
      </Card>
    );
  }

  return <WorkerAnalysis workers={workers} discoveries={discoveries} />;
}
