import { loadRunSummary, loadDiscoveries, loadWorkerLifecycles, loadTree, loadDiversity, loadPlateaus } from '@/lib/data-loader';
import Card from '@/components/Card';
import QualityReport from './QualityReport';

export const dynamic = 'force-dynamic';

export default async function ReportPage({ params }: { params: Promise<{ id: string }> }) {
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
      <Card title="Quality Report">
        <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500">
          <p>No data available for report generation.</p>
        </div>
      </Card>
    );
  }

  return (
    <QualityReport
      summary={summary} discoveries={discoveries} workers={workers}
      tree={tree} diversity={diversity} plateaus={plateaus} runId={id}
    />
  );
}
