import { loadTree } from '@/lib/data-loader';
import Card from '@/components/Card';
import FamilyEvolution from './FamilyEvolution';

export const dynamic = 'force-dynamic';

export default async function FamiliesPage({ params }: { params: Promise<{ id: string }> }) {
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

  if (tree.length === 0) {
    return (
      <Card title="Family Evolution">
        <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500">
          <p className="mb-2">No tree data available.</p>
          <p className="text-xs">Run a PFRS beam search with beam width &gt; 1 to generate tree.csv.</p>
        </div>
      </Card>
    );
  }

  return <FamilyEvolution tree={tree} />;
}
