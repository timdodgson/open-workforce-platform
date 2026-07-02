import { loadPlateaus, loadRunSummary } from '@/lib/data-loader';
import Card from '@/components/Card';
import PlateauAtlas from './PlateauAtlas';

export const dynamic = 'force-dynamic';

export default async function PlateausPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;

  let plateaus, summary;
  try {
    [plateaus, summary] = await Promise.all([
      loadPlateaus(id),
      loadRunSummary(id),
    ]);
  } catch (err) {
    return (
      <Card title="Error">
        <p className="text-red-400 text-sm">Failed to load data: {String(err)}</p>
      </Card>
    );
  }

  if (plateaus.length === 0) {
    return (
      <Card title="Plateau Atlas">
        <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500">
          <p className="mb-2">No plateau data available.</p>
          <p className="text-xs">Run a PFRS beam search to generate plateaus.csv.</p>
        </div>
      </Card>
    );
  }

  return <PlateauAtlas plateaus={plateaus} numWeeks={summary.numWeeks} />;
}
