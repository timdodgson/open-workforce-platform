'use client';

import { useState } from 'react';

export default function RebuildArtifactsButton() {
  const [status, setStatus] = useState<'idle' | 'loading' | 'done' | 'error'>('idle');
  const [message, setMessage] = useState('');

  async function handleRebuild() {
    setStatus('loading');
    setMessage('');
    try {
      const res = await fetch('/api/admin/rebuild-intelligence', { method: 'POST' });
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
    <div className="mt-4 pt-4 border-t border-gray-700">
      <h4 className="text-xs font-semibold text-gray-300 mb-2">Intelligence Artifacts</h4>
      <p className="text-[10px] text-gray-500 mb-2">
        Precompute summary, learning, and policy dashboards to S3 for faster /intelligence loads.
      </p>
      <button
        type="button"
        onClick={handleRebuild}
        disabled={status === 'loading'}
        className="text-xs px-3 py-1.5 rounded bg-blue-600 text-white hover:bg-blue-500 disabled:opacity-50"
      >
        {status === 'loading' ? 'Rebuilding…' : 'Rebuild intelligence artifacts'}
      </button>
      {message && (
        <p className={`text-[10px] mt-2 ${status === 'error' ? 'text-red-400' : 'text-emerald-400'}`}>
          {message}
        </p>
      )}
    </div>
  );
}
