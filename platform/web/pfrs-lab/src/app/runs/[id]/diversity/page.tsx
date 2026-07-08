import { loadDiversity } from '@/lib/data-loader';
import RunPageShell from '@/features/runs/RunPageShell';
import DiversityCharts from '@/app/diversity/DiversityCharts';

export const dynamic = 'force-dynamic';

export default async function RunDiversityPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  try {
    const records = await loadDiversity(id);
    return (
      <RunPageShell
        title="Beam Diversity"
        empty={records.length === 0}
        emptyMessage="No diversity data for this run."
      >
        <DiversityCharts records={records} />
      </RunPageShell>
    );
  } catch (err) {
    return (
      <RunPageShell title="Beam Diversity" error={String(err)}>
        {null}
      </RunPageShell>
    );
  }
}
