import { loadTree, loadRunSummary } from '@/lib/data-loader';
import Card from '@/components/Card';
import TreeView from '@/app/tree/TreeView';

export const dynamic = 'force-dynamic';

export default async function RunTreePage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const [nodes, summary] = await Promise.all([loadTree(id), loadRunSummary(id)]);

  if (nodes.length === 0) {
    return (
      <Card title="Search Tree">
        <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500">
          <p>No tree data for this run.</p>
        </div>
      </Card>
    );
  }

  return <TreeView nodes={nodes} />;
}
