import { loadTree, loadDiscoveries, loadRunSummary } from '@/lib/data-loader';
import RunPageShell from '@/features/runs/RunPageShell';
import InsightsPanel from './InsightsPanel';

export const dynamic = 'force-dynamic';

export default async function InsightsPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  try {
    const [nodes, discoveries, summary] = await Promise.all([
      loadTree(id),
      loadDiscoveries(id),
      loadRunSummary(id),
    ]);

    return (
      <RunPageShell
        title="Research Insights"
        empty={nodes.length === 0}
        emptyMessage="No tree data available for analysis."
      >
        <InsightsPanel nodes={nodes} discoveries={discoveries} summary={summary} runId={id} />
      </RunPageShell>
    );
  } catch (err) {
    return (
      <RunPageShell title="Research Insights" error={String(err)}>
        {null}
      </RunPageShell>
    );
  }
}
