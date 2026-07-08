'use client';

import Card from '@/components/Card';
import PredictionExplorer from '../predictions/PredictionExplorer';
import WhatIfLab from '../what-if/WhatIfLab';
import { usePredictions } from '../predictions/usePredictions';

export function PredictionsTabClient() {
  const { data, loading, error } = usePredictions(2000);

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
  return <PredictionExplorer data={data} />;
}

export function WhatIfTabClient() {
  const { data, loading, error } = usePredictions(2000);

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
