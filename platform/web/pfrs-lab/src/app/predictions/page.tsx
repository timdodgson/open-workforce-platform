import Card from '@/components/Card';
import { getStorageProvider } from '@/lib/storage';
import PredictionExplorer from './PredictionExplorer';

export const dynamic = 'force-dynamic';

export interface WorkerPrediction {
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
  decision_path: { feature: string; condition: string; threshold: number; value: number }[];
  feature_contributions: Record<string, number>;
  feature_values: Record<string, number>;
  explanation: string;
}

export interface PredictionsData {
  version: string;
  total_predictions: number;
  predictions: WorkerPrediction[];
}

export default async function PredictionsPage() {
  const storage = getStorageProvider();
  const content = await storage.readRootFile('worker_predictions.json');

  if (!content) {
    return (
      <Card title="Worker Prediction Explorer">
        <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500">
          <p className="mb-2">No prediction data available.</p>
          <p className="text-xs">
            Train the ML model and generate predictions to populate this page.
            See README for commands.
          </p>
        </div>
      </Card>
    );
  }

  let data: PredictionsData;
  try {
    data = JSON.parse(content) as PredictionsData;
  } catch {
    return (
      <Card title="Worker Prediction Explorer">
        <div className="border-2 border-dashed border-red-700 rounded-lg p-8 text-center text-red-400">
          <p>Failed to parse worker_predictions.json</p>
        </div>
      </Card>
    );
  }

  return <PredictionExplorer data={data} />;
}
