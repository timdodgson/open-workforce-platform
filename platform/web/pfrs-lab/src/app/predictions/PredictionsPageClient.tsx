'use client';

import Card from '@/components/Card';
import PredictionExplorer from './PredictionExplorer';
import { usePredictions } from './usePredictions';

export default function PredictionsPageClient() {
  const { data, loading, error, hasMore } = usePredictions(2000);

  if (loading) {
    return (
      <Card title="Worker Prediction Explorer">
        <p className="text-sm text-gray-400 p-8 text-center">Loading predictions…</p>
      </Card>
    );
  }

  if (error || !data) {
    return (
      <Card title="Worker Prediction Explorer">
        <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500">
          <p className="mb-2">No prediction data available.</p>
          <p className="text-xs">
            Run <code className="text-blue-400">python -m worker_model.predict</code> to generate per-run shards.
          </p>
          {error && <p className="text-xs text-red-400 mt-2">{error}</p>}
        </div>
      </Card>
    );
  }

  return (
    <div>
      {hasMore && (
        <p className="text-[10px] text-gray-500 mb-2">
          Showing {data.predictions.length} of {data.total_predictions} workers (paginated API).
        </p>
      )}
      <PredictionExplorer data={data} />
    </div>
  );
}
