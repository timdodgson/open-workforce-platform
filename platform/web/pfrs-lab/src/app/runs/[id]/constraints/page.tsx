import { loadRunSummary, loadRoster } from '@/lib/data-loader';
import Card from '@/components/Card';
import ConstraintAnalysis from './ConstraintAnalysis';

export const dynamic = 'force-dynamic';

export default async function ConstraintsPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;

  let summary, roster;
  try {
    [summary, roster] = await Promise.all([
      loadRunSummary(id),
      loadRoster(id),
    ]);
  } catch (err) {
    return (
      <Card title="Error">
        <p className="text-red-400 text-sm">Failed to load data: {String(err)}</p>
      </Card>
    );
  }

  return (
    <ConstraintAnalysis
      weeks={summary.weeks}
      totalPenalty={summary.totalPenalty}
      roster={roster}
      numWeeks={summary.numWeeks}
    />
  );
}
