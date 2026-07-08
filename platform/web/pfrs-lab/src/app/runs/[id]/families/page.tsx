import { loadTree } from '@/lib/data-loader';
import RunPageShell from '@/features/runs/RunPageShell';
import FamilyEvolution from './FamilyEvolution';

export const dynamic = 'force-dynamic';

export default async function FamiliesPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;

  try {
    const tree = await loadTree(id);
    return (
      <RunPageShell
        title="Family Evolution"
        empty={tree.length === 0}
        emptyMessage="No tree data available. Run a PFRS beam search with beam width > 1 to generate tree.csv."
      >
        <FamilyEvolution tree={tree} />
      </RunPageShell>
    );
  } catch (err) {
    return (
      <RunPageShell title="Family Evolution" error={String(err)}>
        {null}
      </RunPageShell>
    );
  }
}
