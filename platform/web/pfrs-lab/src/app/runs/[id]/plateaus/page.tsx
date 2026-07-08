import { loadPlateaus, loadRunSummary } from '@/lib/data-loader';
import RunPageShell from '@/features/runs/RunPageShell';
import PlateauAtlas from './PlateauAtlas';

export const dynamic = 'force-dynamic';

export default async function PlateausPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;

  try {
    const [plateaus, summary] = await Promise.all([
      loadPlateaus(id),
      loadRunSummary(id),
    ]);

    return (
      <RunPageShell
        title="Plateau Atlas"
        empty={plateaus.length === 0}
        emptyMessage="No plateau data. Run a PFRS beam search to generate plateaus.csv."
      >
        <PlateauAtlas plateaus={plateaus} numWeeks={summary.numWeeks} />
      </RunPageShell>
    );
  } catch (err) {
    return <RunPageShell title="Plateau Atlas" error={String(err)}>{null}</RunPageShell>;
  }
}
