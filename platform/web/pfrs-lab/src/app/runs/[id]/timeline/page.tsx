import { loadDiscoveries, loadImprovements, loadWorkerLifecycles, loadPlateaus, loadRunSummary } from '@/lib/data-loader';
import Card from '@/components/Card';
import RunTimeline from './RunTimeline';

export const dynamic = 'force-dynamic';

export default async function TimelinePage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;

  let discoveries, improvements, workers, plateaus, summary;
  try {
    [discoveries, improvements, workers, plateaus, summary] = await Promise.all([
      loadDiscoveries(id),
      loadImprovements(id),
      loadWorkerLifecycles(id),
      loadPlateaus(id),
      loadRunSummary(id),
    ]);
  } catch (err) {
    return (
      <Card title="Error">
        <p className="text-red-400 text-sm">Failed to load data: {String(err)}</p>
      </Card>
    );
  }

  if (workers.length === 0 && discoveries.length === 0) {
    return (
      <Card title="Run Timeline">
        <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500">
          <p className="mb-2">No timeline data available.</p>
          <p className="text-xs">Run a PFRS beam search to generate telemetry.</p>
        </div>
      </Card>
    );
  }

  return (
    <RunTimeline
      discoveries={discoveries}
      improvements={improvements}
      workers={workers}
      plateaus={plateaus}
      weeks={summary.weeks}
    />
  );
}
