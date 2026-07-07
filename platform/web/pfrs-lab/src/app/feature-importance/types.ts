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
