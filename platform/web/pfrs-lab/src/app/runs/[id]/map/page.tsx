import { loadDiscoveries } from '@/lib/data-loader';
import Card from '@/components/Card';
import SearchMap from './SearchMap';

export const dynamic = 'force-dynamic';

export default async function MapPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;

  let discoveries;
  try {
    discoveries = await loadDiscoveries(id);
  } catch (err) {
    return (
      <Card title="Error">
        <p className="text-red-400 text-sm">Failed to load data: {String(err)}</p>
      </Card>
    );
  }

  if (discoveries.length === 0) {
    return (
      <Card title="Search Map">
        <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500">
          <p className="mb-2">No discovery data available.</p>
          <p className="text-xs">Run a PFRS beam search to generate discoveries.csv.</p>
        </div>
      </Card>
    );
  }

  return <SearchMap discoveries={discoveries} />;
}
