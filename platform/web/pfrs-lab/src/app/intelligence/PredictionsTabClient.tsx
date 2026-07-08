'use client';

import Card from '@/components/Card';
import PredictionExplorer from '../predictions/PredictionExplorer';
import WhatIfLab from '../what-if/WhatIfLab';
import { usePredictions } from '../predictions/usePredictions';

export function PredictionsTabClient() {
  const { data, loading, loadingMore, error, hasMore, loadMore } = usePredictions();

  if (loading) {
    return <Card title="Predictions"><p className="text-sm text-gray-400 p-6 text-center">Loading…</p></Card>;
  }
  if (error || !data?.predictions.length) {
    return (
      <Card title="Predictions">
        <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500 text-xs">
          No prediction data. Run worker_model.predict to shard predictions per run.
        </div>
      </Card>
    );
  }
  return (
    <>
      <PredictionExplorer data={data} />
      {hasMore && (
        <div className="mt-3 text-center">
          <button
            type="button"
            onClick={loadMore}
            disabled={loadingMore}
            className="text-xs px-4 py-2 rounded bg-gray-800 text-blue-400 hover:bg-gray-700 disabled:opacity-50"
          >
            {loadingMore ? 'Loading…' : `Load more (${data.predictions.length} loaded)`}
          </button>
        </div>
      )}
    </>
  );
}

export function WhatIfTabClient() {
  const { data, loading, error } = usePredictions(500);

  if (loading) {
    return <Card title="What-If Lab"><p className="text-sm text-gray-400 p-6 text-center">Loading…</p></Card>;
  }
  if (error || !data?.predictions.length) {
    return (
      <Card title="What-If Lab">
        <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500 text-xs">
          No prediction data for simulation.
        </div>
      </Card>
    );
  }
  return <WhatIfLab predictions={data.predictions} />;
}
