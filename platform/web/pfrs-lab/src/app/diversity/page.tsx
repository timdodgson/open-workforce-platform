import { loadDiversity } from '@/lib/data-loader';
import Card from '@/components/Card';
import DiversityCharts from './DiversityCharts';

export default async function DiversityPage() {
  const records = await loadDiversity();

  if (records.length === 0) {
    return (
      <div>
        <Card title="Beam Diversity">
          <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500">
            <p className="mb-2">Diversity data not available.</p>
            <p className="text-xs">
              Run <code className="bg-gray-800 px-1 rounded">owp tune-pfrs --pfrs-beam-width 5 --pfrs-beam-seeds 42,101,202</code> to generate diversity.csv.
            </p>
          </div>
        </Card>
      </div>
    );
  }

  return (
    <div>
      <DiversityCharts records={records} />
    </div>
  );
}
