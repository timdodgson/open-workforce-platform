import Card from '@/components/Card';

export default function ComparePage() {
  return (
    <div>
      <Card title="Compare Runs">
        <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500">
          <p>Run comparison will be added in a future iteration.</p>
          <p className="text-xs mt-3 text-gray-600">
            Planned structure: <code className="bg-gray-800 px-1 rounded">data/runs/&lt;run-id&gt;/results.csv</code>
          </p>
        </div>
      </Card>
    </div>
  );
}
