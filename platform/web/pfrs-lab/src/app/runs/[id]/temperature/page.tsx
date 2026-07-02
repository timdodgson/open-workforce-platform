import { loadDiscoveries, loadWorkerLifecycles, loadRunMetadata } from '@/lib/data-loader';
import Card from '@/components/Card';
import TemperatureLandscape from './TemperatureLandscape';

export const dynamic = 'force-dynamic';

export default async function TemperaturePage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;

  let discoveries, workers, metadata;
  try {
    [discoveries, workers, metadata] = await Promise.all([
      loadDiscoveries(id),
      loadWorkerLifecycles(id),
      loadRunMetadata(id),
    ]);
  } catch (err) {
    return (
      <Card title="Error">
        <p className="text-red-400 text-sm">Failed to load data: {String(err)}</p>
      </Card>
    );
  }

  // Hide for non-SA algorithms.
  const mode = metadata?.mode?.toLowerCase() || '';
  if (mode === 'lahc') {
    return (
      <Card title="Temperature Landscape">
        <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500">
          <p className="mb-2">This page is only available for Simulated Annealing runs.</p>
          <p className="text-xs">Current algorithm: LAHC (no temperature)</p>
        </div>
      </Card>
    );
  }

  if (discoveries.length === 0 && workers.length === 0) {
    return (
      <Card title="Temperature Landscape">
        <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500">
          <p>No temperature data available.</p>
        </div>
      </Card>
    );
  }

  return <TemperatureLandscape discoveries={discoveries} workers={workers} metadata={metadata} />;
}
