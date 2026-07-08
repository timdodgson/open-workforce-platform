import { loadWorkerLifecycles, loadDiscoveries } from '@/lib/data-loader';
import RunPageShell from '@/features/runs/RunPageShell';
import WorkerArchetypes from './WorkerArchetypes';

export const dynamic = 'force-dynamic';

export default async function ArchetypesPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;

  try {
    const [workers, discoveries] = await Promise.all([
      loadWorkerLifecycles(id),
      loadDiscoveries(id),
    ]);

    return (
      <RunPageShell
        title="Worker Archetypes"
        empty={workers.length === 0}
        emptyMessage="No worker data. Run a PFRS beam search to generate workers.csv."
      >
        <WorkerArchetypes workers={workers} discoveries={discoveries} />
      </RunPageShell>
    );
  } catch (err) {
    return (
      <RunPageShell title="Worker Archetypes" error={String(err)}>
        {null}
      </RunPageShell>
    );
  }
}
