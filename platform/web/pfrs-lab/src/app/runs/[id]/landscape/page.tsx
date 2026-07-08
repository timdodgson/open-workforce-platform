import { loadDiversity, loadDiscoveries, loadTree } from '@/lib/data-loader';
import RunPageShell from '@/features/runs/RunPageShell';
import FitnessLandscape from './FitnessLandscape';

export const dynamic = 'force-dynamic';

interface Props {
  params: Promise<{ id: string }>;
}

export default async function LandscapePage({ params }: Props) {
  const { id } = await params;
  try {
    const [diversity, discoveries, tree] = await Promise.all([
      loadDiversity(id),
      loadDiscoveries(id),
      loadTree(id),
    ]);

    const empty = diversity.length === 0 && discoveries.length === 0;
    return (
      <RunPageShell
        title="Fitness Landscape"
        empty={empty}
        emptyMessage="Insufficient telemetry for landscape reconstruction. Requires diversity.csv and discoveries.csv from a beam search run."
      >
        <FitnessLandscape diversity={diversity} discoveries={discoveries} tree={tree} />
      </RunPageShell>
    );
  } catch (err) {
    return (
      <RunPageShell title="Fitness Landscape" error={String(err)}>
        {null}
      </RunPageShell>
    );
  }
}
