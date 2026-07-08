import { loadDiscoveries, loadImprovements, loadWorkerLifecycles, loadPlateaus, loadRunSummary } from '@/lib/data-loader';
import RunPageShell from '@/features/runs/RunPageShell';
import RunTimeline from './RunTimeline';

export const dynamic = 'force-dynamic';

export default async function TimelinePage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;

  try {
    const [discoveries, improvements, workers, plateaus, summary] = await Promise.all([
      loadDiscoveries(id),
      loadImprovements(id),
      loadWorkerLifecycles(id),
      loadPlateaus(id),
      loadRunSummary(id),
    ]);

    const empty = workers.length === 0 && discoveries.length === 0;
    return (
      <RunPageShell
        title="Run Timeline"
        empty={empty}
        emptyMessage="No timeline data available. Run a PFRS beam search to generate telemetry."
      >
        <RunTimeline
          discoveries={discoveries}
          improvements={improvements}
          workers={workers}
          plateaus={plateaus}
          weeks={summary.weeks}
        />
      </RunPageShell>
    );
  } catch (err) {
    return (
      <RunPageShell title="Run Timeline" error={String(err)}>
        {null}
      </RunPageShell>
    );
  }
}
