import { loadDiscoveries, loadImprovements, loadTree, loadWorkerLifecycles } from '@/lib/data-loader';
import RunPageShell from '@/features/runs/RunPageShell';
import ReplayPlayer from './ReplayPlayer';

export const dynamic = 'force-dynamic';

export default async function ReplayPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;

  try {
    const [discoveries, improvements, tree, workers] = await Promise.all([
      loadDiscoveries(id),
      loadImprovements(id),
      loadTree(id),
      loadWorkerLifecycles(id),
    ]);
    return (
      <RunPageShell
        title="Search Replay"
        empty={discoveries.length === 0 && improvements.length === 0}
        emptyMessage="No replay data available. Run a PFRS beam search to generate telemetry."
      >
        <ReplayPlayer
          discoveries={discoveries}
          improvements={improvements}
          tree={tree}
          workers={workers}
        />
      </RunPageShell>
    );
  } catch (err) {
    return (
      <RunPageShell title="Search Replay" error={String(err)}>
        {null}
      </RunPageShell>
    );
  }
}
