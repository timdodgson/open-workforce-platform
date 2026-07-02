import { loadRunSummary } from '@/lib/data-loader';
import Card from '@/components/Card';
import PenaltyWaterfall from './PenaltyWaterfall';

export const dynamic = 'force-dynamic';

export default async function WaterfallPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;

  let summary;
  try {
    summary = await loadRunSummary(id);
  } catch (err) {
    return (
      <Card title="Error">
        <p className="text-red-400 text-sm">Failed to load data: {String(err)}</p>
      </Card>
    );
  }

  if (summary.weeks.length === 0) {
    return (
      <Card title="Penalty Waterfall">
        <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500">
          <p>No results data available.</p>
        </div>
      </Card>
    );
  }

  return <PenaltyWaterfall weeks={summary.weeks} totalPenalty={summary.totalPenalty} />;
}
