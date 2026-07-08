import { loadTree } from '@/lib/data-loader';
import RunPageShell from '@/features/runs/RunPageShell';
import InheritanceCharts from './InheritanceCharts';

export const dynamic = 'force-dynamic';

export default async function InheritancePage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  try {
    const nodes = await loadTree(id);
    return (
      <RunPageShell
        title="Inheritance Analysis"
        empty={nodes.length === 0}
        emptyMessage="No tree data available for this run."
      >
        <InheritanceCharts nodes={nodes} />
      </RunPageShell>
    );
  } catch (err) {
    return (
      <RunPageShell title="Inheritance Analysis" error={String(err)}>
        {null}
      </RunPageShell>
    );
  }
}
