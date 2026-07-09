'use client';

import { useState } from 'react';
import Card from '@/components/Card';

export default function RebuildArtifactsButton() {
  const [status, setStatus] = useState<'idle' | 'loading' | 'done' | 'error'>('idle');
  const [message, setMessage] = useState('');

  async function handleRebuild() {
    setStatus('loading');
    setMessage('');
    try {
      const res = await fetch('/api/admin/rebuild-intelligence', {
        method: 'POST',
        credentials: 'include',
      });
      const json = await res.json();
      if (!res.ok) throw new Error(json.error || `HTTP ${res.status}`);
      setStatus('done');
      setMessage(`Wrote ${json.paths?.length ?? 3} artifacts at ${json.generatedAt}`);
    } catch (e) {
      setStatus('error');
      setMessage(e instanceof Error ? e.message : 'Rebuild failed');
    }
  }

  return (
    <Card title="Intelligence Artifacts">
      <p className="text-xs text-gray-400 mb-4">
        Precompute summary, learning, and policy dashboards to S3 for faster /intelligence page loads.
        Run this after uploading new runs or retraining policies.
      </p>
      <button
        type="button"
        onClick={handleRebuild}
        disabled={status === 'loading'}
        className="text-sm px-4 py-2 rounded bg-blue-600 text-white hover:bg-blue-500 disabled:opacity-50 font-medium"
      >
        {status === 'loading' ? 'Rebuilding…' : 'Rebuild intelligence artifacts'}
      </button>
      {message && (
        <p className={`text-xs mt-3 ${status === 'error' ? 'text-red-400' : 'text-emerald-400'}`}>
          {message}
        </p>
      )}
    </Card>
  );
}
