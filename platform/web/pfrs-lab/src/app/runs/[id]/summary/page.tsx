import RunPageShell from '@/features/runs/RunPageShell';
import { renderRunSummary } from '@/features/runs/summary/renderRunSummary';

export const dynamic = 'force-dynamic';

export default async function RunSummaryPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;

  try {
    return (
      <RunPageShell title="Run Summary">
        {await renderRunSummary(id)}
      </RunPageShell>
    );
  } catch (err) {
    return (
      <RunPageShell title="Run Summary" error={String(err)}>
        {null}
      </RunPageShell>
    );
  }
}
