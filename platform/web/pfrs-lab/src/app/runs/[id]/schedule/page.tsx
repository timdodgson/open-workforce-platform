import { loadRoster, loadRunSummary, loadWeeks } from '@/lib/data-loader';
import RunPageShell from '@/features/runs/RunPageShell';
import ScheduleViewer from './ScheduleViewer';

export const dynamic = 'force-dynamic';

export default async function SchedulePage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  try {
    const [roster, summary, weeks] = await Promise.all([
      loadRoster(id),
      loadRunSummary(id),
      loadWeeks(id),
    ]);

    const emptyMessage = weeks.length > 0 && roster.length === 0
      ? 'This run has week summaries but no roster.json — schedule grid needs a full PFRS beam run (tune-pfrs) with roster export. SI benchmark uploads often omit roster data.'
      : 'No roster data available. Run tune-pfrs with --pfrs-run-label to generate roster.json, or owp solve nrp for a single-week roster in solution.json.';

    return (
      <RunPageShell
        title="Schedule Viewer"
        empty={roster.length === 0}
        emptyMessage={emptyMessage}
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
