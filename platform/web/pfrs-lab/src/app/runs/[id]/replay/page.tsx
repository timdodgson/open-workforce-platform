import { loadDiscoveries, loadImprovements, loadTree, loadWorkerLifecycles } from '@/lib/data-loader';
import Card from '@/components/Card';
import ReplayPlayer from './ReplayPlayer';

export const dynamic = 'force-dynamic';

export default async function ReplayPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;

  let discoveries, improvements, tree, workers;
  try {
    [discoveries, improvements, tree, workers] = await Promise.all([
      loadDiscoveries(id),
      loadImprovements(id),
      loadTree(id),
      loadWorkerLifecycles(id),
    ]);
  } catch (err) {
    return (
      <Card title="Error">
        <p className="text-red-400 text-sm">Failed to load data: {String(err)}</p>
      </Card>
    );
  }

  if (discoveries.length === 0 && improvements.length === 0) {
    return (
      <Card title="Search Replay">
        <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500">
          <p className="mb-2">No replay data available.</p>
          <p className="text-xs">Run a PFRS beam search to generate telemetry.</p>
        </div>
      </Card>
    );
  }

  return (
    <ReplayPlayer
      discoveries={discoveries}
      improvements={improvements}
      tree={tree}
      workers={workers}
    />
  );
}
