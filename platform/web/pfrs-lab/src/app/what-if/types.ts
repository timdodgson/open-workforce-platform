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
