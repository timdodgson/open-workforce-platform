import { loadDiversity, loadDiscoveries, loadTree } from '@/lib/data-loader';
import Card from '@/components/Card';
import FitnessLandscape from './FitnessLandscape';

export const dynamic = 'force-dynamic';

interface Props {
  params: Promise<{ id: string }>;
}

export default async function LandscapePage({ params }: Props) {
  const { id } = await params;
  const [diversity, discoveries, tree] = await Promise.all([
    loadDiversity(id),
    loadDiscoveries(id),
    loadTree(id),
  ]);

  if (diversity.length === 0 && discoveries.length === 0) {
    return (
      <Card title="Fitness Landscape">
        <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500">
          <p className="mb-2">Insufficient telemetry for landscape reconstruction.</p>
          <p className="text-xs">Requires diversity.csv and discoveries.csv from a beam search run.</p>
        </div>
      </Card>
    );
  }

  return <FitnessLandscape diversity={diversity} discoveries={discoveries} tree={tree} />;
}
