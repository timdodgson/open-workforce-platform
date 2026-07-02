import { loadDiscoveries, loadImprovements, loadWorkerLifecycles, loadPlateaus, loadTree } from '@/lib/data-loader';
import Card from '@/components/Card';
import CausalityExplorer from './CausalityExplorer';

export const dynamic = 'force-dynamic';

export default async function CausalityPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;

  let discoveries, improvements, workers, plateaus, tree;
  try {
    [discoveries, improvements, workers, plateaus, tree] = await Promise.all([
      loadDiscoveries(id),
      loadImprovements(id),
      loadWorkerLifecycles(id),
      loadPlateaus(id),
      loadTree(id),
    ]);
  } catch (err) {
    return (
      <Card title="Error">
        <p className="text-red-400 text-sm">Failed to load data: {String(err)}</p>
      </Card>
    );
  }

  if (discoveries.length === 0 && workers.length === 0) {
    return (
      <Card title="Causality Explorer">
        <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500">
          <p className="mb-2">No telemetry data available.</p>
          <p className="text-xs">Run a PFRS beam search to generate data.</p>
        </div>
      </Card>
    );
  }

  return (
    <CausalityExplorer
      discoveries={discoveries}
      improvements={improvements}
      workers={workers}
      plateaus={plateaus}
      tree={tree}
    />
  );
}
