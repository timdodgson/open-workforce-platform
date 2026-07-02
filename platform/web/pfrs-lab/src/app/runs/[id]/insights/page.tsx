import { loadTree, loadDiscoveries, loadRunSummary } from '@/lib/data-loader';
import Card from '@/components/Card';
import InsightsPanel from './InsightsPanel';

export const dynamic = 'force-dynamic';

export default async function InsightsPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const [nodes, discoveries, summary] = await Promise.all([
    loadTree(id),
    loadDiscoveries(id),
    loadRunSummary(id),
  ]);

  if (nodes.length === 0) {
    return (
      <Card title="Research Insights">
        <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500">
          <p>No tree data available for analysis.</p>
        </div>
      </Card>
    );
  }

  return <InsightsPanel nodes={nodes} discoveries={discoveries} summary={summary} runId={id} />;
}
