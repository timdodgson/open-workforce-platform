import { loadTree, loadRunSummary } from '@/lib/data-loader';
import RunPageShell from '@/features/runs/RunPageShell';
import TreeView from '@/app/tree/TreeView';

export const dynamic = 'force-dynamic';

export default async function RunTreePage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;

  try {
    const [nodes] = await Promise.all([loadTree(id), loadRunSummary(id)]);

    return (
      <RunPageShell
        title="Search Tree"
        empty={nodes.length === 0}
        emptyMessage="No tree data for this run."
      >
        <TreeView nodes={nodes} />
      </RunPageShell>
    );
  } catch (err) {
    return <RunPageShell title="Search Tree" error={String(err)}>{null}</RunPageShell>;
  }
}
