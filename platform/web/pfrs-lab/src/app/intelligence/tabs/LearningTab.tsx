'use client';

import { useEffect, useState } from 'react';
import Card from '@/components/Card';

export default function LearningTab() {
  const [content, setContent] = useState<'loading' | 'ready' | 'empty'>('loading');

  useEffect(() => {
    // Signal that this tab is ready (content loaded from existing page via redirect)
    setContent('ready');
  }, []);

  return (
    <Card title="Learning — Worker Behaviour Telemetry">
      <p className="text-xs text-gray-400 mb-3">
        How search workers behave across runs. Training data for the ML model.
        One row per completed worker or search run.
      </p>
      <p className="text-xs text-gray-500">
        Full interactive dashboard available at{' '}
        <a href="/learning" className="text-blue-400 hover:underline">/learning</a>
      </p>
    </Card>
  );
}
