import { loadRoster, loadRunSummary } from '@/lib/data-loader';
import RunPageShell from '@/features/runs/RunPageShell';
import ScheduleViewer from './ScheduleViewer';

export const dynamic = 'force-dynamic';

export default async function SchedulePage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  try {
    const [roster, summary] = await Promise.all([loadRoster(id), loadRunSummary(id)]);
    return (
      <RunPageShell
        title="Schedule Viewer"
        empty={roster.length === 0}
        emptyMessage="No roster data available. Re-run the optimiser to generate roster.json."
      >
        <ScheduleViewer roster={roster} summary={summary} />
      </RunPageShell>
    );
  } catch (err) {
    return (
      <RunPageShell title="Schedule Viewer" error={String(err)}>
        {null}
      </RunPageShell>
    );
  }
}
