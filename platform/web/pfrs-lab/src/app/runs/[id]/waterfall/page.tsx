import { loadRunSummary } from '@/lib/data-loader';
import RunPageShell from '@/features/runs/RunPageShell';
import PenaltyWaterfall from './PenaltyWaterfall';

export const dynamic = 'force-dynamic';

export default async function WaterfallPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;

  try {
    const summary = await loadRunSummary(id);
    return (
      <RunPageShell
        title="Penalty Waterfall"
        empty={summary.weeks.length === 0}
        emptyMessage="No results data available."
      >
        <PenaltyWaterfall weeks={summary.weeks} totalPenalty={summary.totalPenalty} />
      </RunPageShell>
    );
  } catch (err) {
    return (
      <RunPageShell title="Penalty Waterfall" error={String(err)}>
        {null}
      </RunPageShell>
    );
  }
}
