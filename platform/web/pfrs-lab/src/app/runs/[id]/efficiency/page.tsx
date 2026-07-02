import { loadWorkerLifecycles, loadDiscoveries, loadRunSummary } from '@/lib/data-loader';
import Card from '@/components/Card';
import EfficiencyDashboard from './EfficiencyDashboard';

export const dynamic = 'force-dynamic';

export default async function EfficiencyPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;

  let workers, discoveries, summary;
  try {
    [workers, discoveries, summary] = await Promise.all([
      loadWorkerLifecycles(id),
      loadDiscoveries(id),
      loadRunSummary(id),
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
      <Card title="Efficiency Dashboard">
        <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500">
          <p className="mb-2">No worker data available.</p>
          <p className="text-xs">Run a PFRS beam search to generate telemetry.</p>
        </div>
      </Card>
    );
  }

  return <EfficiencyDashboard workers={workers} discoveries={discoveries} summary={summary} />;
}
