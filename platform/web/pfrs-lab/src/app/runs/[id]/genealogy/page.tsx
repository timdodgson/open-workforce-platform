import { loadTree, loadWorkerLifecycles } from '@/lib/data-loader';
import RunPageShell from '@/features/runs/RunPageShell';
import BranchGenealogy from './BranchGenealogy';

export const dynamic = 'force-dynamic';

export default async function GenealogyPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;

  try {
    const [tree, workers] = await Promise.all([
      loadTree(id),
      loadWorkerLifecycles(id),
    ]);
    return (
      <RunPageShell
        title="Branch Genealogy"
        empty={tree.length === 0}
        emptyMessage="No tree data available. Run a PFRS beam search with beam width > 1."
      >
        <BranchGenealogy tree={tree} workers={workers} />
      </RunPageShell>
    );
  } catch (err) {
    return (
      <RunPageShell title="Branch Genealogy" error={String(err)}>
        {null}
      </RunPageShell>
    );
  }
}
