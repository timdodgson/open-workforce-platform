import { loadRunSummary, loadRoster } from '@/lib/data-loader';
import RunPageShell from '@/features/runs/RunPageShell';
import ConstraintAnalysis from './ConstraintAnalysis';

export const dynamic = 'force-dynamic';

export default async function ConstraintsPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;

  try {
    const [summary, roster] = await Promise.all([
      loadRunSummary(id),
      loadRoster(id),
    ]);
    return (
      <RunPageShell title="Constraints">
        <ConstraintAnalysis
          weeks={summary.weeks}
          totalPenalty={summary.totalPenalty}
          roster={roster}
          numWeeks={summary.numWeeks}
        />
      </RunPageShell>
    );
  } catch (err) {
    return (
      <RunPageShell title="Constraints" error={String(err)}>
        {null}
      </RunPageShell>
    );
  }
}
