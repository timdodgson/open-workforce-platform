import { loadDiversity } from '@/lib/data-loader';
import Card from '@/components/Card';
import DiversityCharts from '@/app/diversity/DiversityCharts';

export const dynamic = 'force-dynamic';

export default async function RunDiversityPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const records = await loadDiversity(id);

  if (records.length === 0) {
    return (
      <Card title="Beam Diversity">
        <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500">
          <p>No diversity data for this run.</p>
        </div>
      </Card>
    );
  }

  return <DiversityCharts records={records} />;
}
