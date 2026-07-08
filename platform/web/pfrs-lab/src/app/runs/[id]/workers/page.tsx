import { loadWorkerLifecycles, loadDiscoveries } from '@/lib/data-loader';
import RunPageShell from '@/features/runs/RunPageShell';
import WorkerAnalysis from './WorkerAnalysis';

export const dynamic = 'force-dynamic';

export default async function WorkersPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;

  try {
    const [workers, discoveries] = await Promise.all([
      loadWorkerLifecycles(id),
      loadDiscoveries(id),
    ]);

    return (
      <RunPageShell
        title="Worker Analysis"
        empty={workers.length === 0}
        emptyMessage="No worker data. Run a PFRS beam search to generate workers.csv."
      >
        <WorkerAnalysis workers={workers} discoveries={discoveries} />
      </RunPageShell>
    );
  } catch (err) {
    return <RunPageShell title="Worker Analysis" error={String(err)}>{null}</RunPageShell>;
  }
}
