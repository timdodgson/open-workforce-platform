import { listRunsAsync } from '@/lib/data-loader';
import Card from '@/components/Card';
import RunList from '../RunList';

export const dynamic = 'force-dynamic';

export default async function AllRunsPage() {
  const runs = await listRunsAsync();

  if (runs.length === 0) {
    return (
      <Card title="All Runs">
        <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500">
          <p className="mb-2">No runs yet.</p>
          <p className="text-xs">Run experiments with <code className="text-blue-400">--run-label</code> to populate this page.</p>
        </div>
      </Card>
    );
  }

  return <RunList runs={runs} />;
}
