import { loadDiscoveries } from '@/lib/data-loader';
import RunPageShell from '@/features/runs/RunPageShell';
import SearchMap from './SearchMap';

export const dynamic = 'force-dynamic';

export default async function MapPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;

  try {
    const discoveries = await loadDiscoveries(id);
    return (
      <RunPageShell
        title="Search Map"
        empty={discoveries.length === 0}
        emptyMessage="No discovery data. Run a PFRS beam search to generate discoveries.csv."
      >
        <SearchMap discoveries={discoveries} />
      </RunPageShell>
    );
  } catch (err) {
    return (
      <RunPageShell title="Search Map" error={String(err)}>
        {null}
      </RunPageShell>
    );
  }
}
