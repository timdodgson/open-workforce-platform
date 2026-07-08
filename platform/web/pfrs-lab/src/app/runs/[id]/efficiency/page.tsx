import { loadWorkerLifecycles, loadDiscoveries, loadRunSummary } from '@/lib/data-loader';
import RunPageShell from '@/features/runs/RunPageShell';
import EfficiencyDashboard from './EfficiencyDashboard';

export const dynamic = 'force-dynamic';

export default async function EfficiencyPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;

  try {
    const [workers, discoveries, summary] = await Promise.all([
      loadWorkerLifecycles(id),
      loadDiscoveries(id),
      loadRunSummary(id),
    ]);
    return (
      <RunPageShell
        title="Efficiency Dashboard"
        empty={workers.length === 0}
        emptyMessage="No worker data available. Run a PFRS beam search to generate telemetry."
      >
        <EfficiencyDashboard workers={workers} discoveries={discoveries} summary={summary} />
      </RunPageShell>
    );
  } catch (err) {
    return (
      <RunPageShell title="Efficiency Dashboard" error={String(err)}>
        {null}
      </RunPageShell>
    );
  }
}
