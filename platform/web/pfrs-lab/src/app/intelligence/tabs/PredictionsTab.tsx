'use client';

import Card from '@/components/Card';

export default function PredictionsTab() {
  return (
    <Card title="Predictions — Predicted vs Actual">
      <p className="text-xs text-gray-400 mb-3">
        Per-worker predictions from the trained ML model compared against actual outcomes.
        Shows prediction accuracy, confidence calibration, and error analysis.
      </p>
      <p className="text-xs text-gray-500">
        Full interactive dashboard available at{' '}
        <a href="/predictions" className="text-blue-400 hover:underline">/predictions</a>
      </p>
    </Card>
  );
}
