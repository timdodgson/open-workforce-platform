'use client';

import Card from '@/components/Card';

export default function WhatIfTab() {
  return (
    <Card title="What-If Lab — Counterfactual Simulation">
      <p className="text-xs text-gray-400 mb-3">
        Simulate alternative decisions and predict outcomes. Explore what would happen
        with different confidence thresholds, budget allocations, or skip policies.
      </p>
      <p className="text-xs text-gray-500">
        Full interactive simulation available at{' '}
        <a href="/what-if" className="text-blue-400 hover:underline">/what-if</a>
      </p>
    </Card>
  );
}
