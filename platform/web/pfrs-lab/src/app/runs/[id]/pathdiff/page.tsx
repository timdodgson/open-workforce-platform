import { loadTree } from '@/lib/data-loader';
import Card from '@/components/Card';
import BeamPathDiff from './BeamPathDiff';

export const dynamic = 'force-dynamic';

export default async function PathDiffPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;

  let tree;
  try {
    tree = await loadTree(id);
  } catch (err) {
    return (
      <Card title="Error">
        <p className="text-red-400 text-sm">Failed to load data: {String(err)}</p>
      </Card>
    );
  }

  if (tree.length < 2) {
    return (
      <Card title="Beam Path Diff">
        <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500">
          <p className="mb-2">Need at least 2 paths for comparison.</p>
          <p className="text-xs">Run with beam width &gt; 1.</p>
        </div>
      </Card>
    );
  }

  return <BeamPathDiff tree={tree} />;
}
