import { loadTree } from '@/lib/data-loader';
import RunPageShell from '@/features/runs/RunPageShell';
import BeamPathDiff from './BeamPathDiff';

export const dynamic = 'force-dynamic';

export default async function PathDiffPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;

  try {
    const tree = await loadTree(id);
    return (
      <RunPageShell
        title="Beam Path Diff"
        empty={tree.length < 2}
        emptyMessage="Need at least 2 paths for comparison. Run with beam width > 1."
      >
        <BeamPathDiff tree={tree} />
      </RunPageShell>
    );
  } catch (err) {
    return (
      <RunPageShell title="Beam Path Diff" error={String(err)}>
        {null}
      </RunPageShell>
    );
  }
}
