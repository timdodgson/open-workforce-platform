import { loadRoster, loadRunSummary } from '@/lib/data-loader';
import Card from '@/components/Card';
import ScheduleViewer from './ScheduleViewer';

export const dynamic = 'force-dynamic';

export default async function SchedulePage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const [roster, summary] = await Promise.all([loadRoster(id), loadRunSummary(id)]);

  if (roster.length === 0) {
    return (
      <Card title="Schedule Viewer">
        <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500">
          <p className="mb-2">No roster data available.</p>
          <p className="text-xs">Re-run the optimiser to generate roster.json.</p>
        </div>
      </Card>
    );
  }

  return <ScheduleViewer roster={roster} summary={summary} />;
}
