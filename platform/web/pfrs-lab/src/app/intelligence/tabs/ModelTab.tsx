'use client';

import Card from '@/components/Card';

export default function ModelTab() {
  return (
    <Card title="Model — Worker Value Prediction">
      <p className="text-xs text-gray-400 mb-3">
        How the trained model predicts worker value. Feature importance, decision tree
        structure, and model accuracy metrics.
      </p>
      <p className="text-xs text-gray-500">
        Full interactive dashboard available at{' '}
        <a href="/feature-importance" className="text-blue-400 hover:underline">/feature-importance</a>
      </p>
    </Card>
  );
}
