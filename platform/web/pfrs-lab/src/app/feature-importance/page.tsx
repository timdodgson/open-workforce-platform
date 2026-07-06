import Card from '@/components/Card';
import { getStorageProvider } from '@/lib/storage';
import FeatureImportanceDashboard from './FeatureImportanceDashboard';

export const dynamic = 'force-dynamic';

export interface WorkerModel {
  version: string;
  description: string;
  training_samples: number;
  test_samples: number;
  features: string[];
  models: {
    improved: ModelResult;
    produced_global_best: ModelResult;
    improvement_amount: ModelResult;
    roi: ModelResult;
  };
  data_summary: {
    total_records: number;
    improvement_rate: number;
    global_best_rate: number;
    mean_improvement: number;
    mean_roi: number;
  };
}

export interface ModelResult {
  target: string;
  type: 'classifier' | 'regressor';
  max_depth: number;
  n_train: number;
  n_test: number;
  metrics: Record<string, number>;
  feature_importance: Record<string, number>;
  confusion_matrix?: {
    tn: number;
    fp: number;
    fn: number;
    tp: number;
  };
  classification_report?: Record<string, unknown>;
  tree_text: string;
}

export default async function FeatureImportancePage() {
  const storage = getStorageProvider();

  // Try loading worker_model.json from root storage.
  const content = await storage.readRootFile('worker_model.json');

  if (!content) {
    return (
      <Card title="Feature Importance — Worker Value Model">
        <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500">
          <p className="mb-2">No trained model available yet.</p>
          <p className="text-xs">
            Train the Worker Value Model to generate <code className="text-blue-400">worker_model.json</code>:
          </p>
          <pre className="mt-3 text-left text-[10px] bg-gray-800 rounded p-3 inline-block">
{`cd platform/ml
pip install -e .
python -m worker_model.train --data-dir <runs-dir> --output worker_model.json`}
          </pre>
          <p className="text-xs mt-3">
            Then place the output in the data root or upload to S3.
          </p>
        </div>
      </Card>
    );
  }

  let model: WorkerModel;
  try {
    model = JSON.parse(content) as WorkerModel;
  } catch {
    return (
      <Card title="Feature Importance — Worker Value Model">
        <div className="border-2 border-dashed border-red-700 rounded-lg p-8 text-center text-red-400">
          <p>Failed to parse worker_model.json</p>
        </div>
      </Card>
    );
  }

  return <FeatureImportanceDashboard model={model} />;
}
