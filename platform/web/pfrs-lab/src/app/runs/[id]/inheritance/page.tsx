import { loadTree } from '@/lib/data-loader';
import Card from '@/components/Card';
import InheritanceCharts from './InheritanceCharts';

export const dynamic = 'force-dynamic';

export default async function InheritancePage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const nodes = await loadTree(id);

  if (nodes.length === 0) {
    return (
      <Card title="Inheritance Analysis">
        <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500">
          <p>No tree data available for this run.</p>
        </div>
      </Card>
    );
  }

  return <InheritanceCharts nodes={nodes} />;
}
