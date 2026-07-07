'use client';

import Card from '@/components/Card';

export default function DecisionsTab() {
  return (
    <Card title="Decision Analysis — Historical Rule Analysis">
      <p className="text-xs text-gray-400 mb-3 leading-relaxed">
        This page documents why the original worker decision rules evolved into the current
        Search Intelligence system. It provides forensic analysis of early rule-engine
        predictions versus actual outcomes.
      </p>
      <p className="text-xs text-gray-400 mb-3 leading-relaxed">
        Current Search Intelligence v3 includes WorkerAssist, SearchAssist, PortfolioAssist,
        Learned Portfolio Allocation, and Adaptive Mode. These later improvements are
        evaluated on the <strong className="text-blue-400">Assist Validation</strong> tab.
      </p>
      <p className="text-xs text-gray-500">
        Full interactive dashboard available at{' '}
        <a href="/decisions" className="text-blue-400 hover:underline">/decisions</a>
      </p>
    </Card>
  );
}
