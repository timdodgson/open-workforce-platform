import Card from '@/components/Card';
import { getStorageProvider } from '@/lib/storage';
import WhatIfLab from './WhatIfLab';

export const dynamic = 'force-dynamic';

export interface WhatIfPrediction {
  index: number;
  run_id: string;
  problem_type: string;
  instance: string;
  algorithm: string;
  seed: number;
  week: number;
  depth: number;
  actual: {
    improved: boolean;
    produced_global_best: boolean;
    improvement_amount: number;
    roi: number;
  };
  predicted: {
    p_improved: number;
    p_global_best: number;
    expected_improvement: number;
    expected_roi: number;
  };
  error: {
    improvement: number;
    roi: number;
  };
  feature_values: Record<string, number>;
  explanation: string;
}

export default async function WhatIfPage() {
  const storage = getStorageProvider();
  const content = await storage.readRootFile('worker_predictions.json');

  if (!content) {
    return (
      <Card title="What-If Lab">
        <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500">
          <p className="mb-2">No prediction data available yet.</p>
          <p className="text-xs">
            Generate predictions first:
          </p>
          <pre className="mt-3 text-left text-[10px] bg-gray-800 rounded p-3 inline-block">
{`cd platform/ml
python -m worker_model.predict \\
  --data-dir <runs-dir> \\
  --output worker_predictions.json`}
          </pre>
        </div>
      </Card>
    );
  }

  let data: { predictions: WhatIfPrediction[] };
  try {
    data = JSON.parse(content);
  } catch {
    return (
      <Card title="What-If Lab">
        <div className="border-2 border-dashed border-red-700 rounded-lg p-8 text-center text-red-400">
          <p>Failed to parse worker_predictions.json</p>
        </div>
      </Card>
    );
  }

  return <WhatIfLab predictions={data.predictions} />;
}
