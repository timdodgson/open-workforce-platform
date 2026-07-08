import { loadDiscoveries, loadWorkerLifecycles, loadRunMetadata } from '@/lib/data-loader';
import RunPageShell from '@/features/runs/RunPageShell';
import TemperatureLandscape from './TemperatureLandscape';

export const dynamic = 'force-dynamic';

export default async function TemperaturePage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;

  try {
    const [discoveries, workers, metadata] = await Promise.all([
      loadDiscoveries(id),
      loadWorkerLifecycles(id),
      loadRunMetadata(id),
    ]);

    const mode = metadata?.mode?.toLowerCase() || '';
    if (mode === 'lahc') {
      return (
        <RunPageShell
          title="Temperature Landscape"
          empty
          emptyMessage="This page is only available for Simulated Annealing runs. Current algorithm: LAHC (no temperature)."
        >
          {null}
        </RunPageShell>
      );
    }

    const empty = discoveries.length === 0 && workers.length === 0;
    return (
      <RunPageShell
        title="Temperature Landscape"
        empty={empty}
        emptyMessage="No temperature data available."
      >
        <TemperatureLandscape discoveries={discoveries} workers={workers} metadata={metadata} />
      </RunPageShell>
    );
  } catch (err) {
    return (
      <RunPageShell title="Temperature Landscape" error={String(err)}>
        {null}
      </RunPageShell>
    );
  }
}
