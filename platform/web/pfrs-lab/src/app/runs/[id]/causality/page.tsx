import { loadDiscoveries, loadImprovements, loadWorkerLifecycles, loadPlateaus, loadTree } from '@/lib/data-loader';
import RunPageShell from '@/features/runs/RunPageShell';
import CausalityExplorer from './CausalityExplorer';

export const dynamic = 'force-dynamic';

export default async function CausalityPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;

  try {
    const [discoveries, improvements, workers, plateaus, tree] = await Promise.all([
      loadDiscoveries(id),
      loadImprovements(id),
      loadWorkerLifecycles(id),
      loadPlateaus(id),
      loadTree(id),
    ]);
    return (
      <RunPageShell
        title="Causality Explorer"
        empty={discoveries.length === 0 && workers.length === 0}
        emptyMessage="No telemetry data available. Run a PFRS beam search to generate data."
      >
        <CausalityExplorer
          discoveries={discoveries}
          improvements={improvements}
          workers={workers}
          plateaus={plateaus}
          tree={tree}
        />
      </RunPageShell>
    );
  } catch (err) {
    return (
      <RunPageShell title="Causality Explorer" error={String(err)}>
        {null}
      </RunPageShell>
    );
  }
}
