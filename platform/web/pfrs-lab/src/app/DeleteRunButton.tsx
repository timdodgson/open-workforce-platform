'use client';
import { useState } from 'react';
import { useRouter } from 'next/navigation';

export default function DeleteRunButton({ runId }: { runId: string }) {
  const router = useRouter();
  const [confirming, setConfirming] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function handleDelete() {
    setDeleting(true);
    setError(null);

    try {
      const res = await fetch(`/api/runs/${runId}`, { method: 'DELETE' });

      if (!res.ok) {
        const body = await res.json().catch(() => ({ error: 'Unknown error' }));
        setError(body.error || `Failed (${res.status})`);
        setDeleting(false);
        return;
      }

      // Success — refresh the server component data.
      router.refresh();
    } catch (err) {
      setError('Network error');
    }

    setDeleting(false);
    setConfirming(false);
  }

  if (error) {
    return (
      <div className="flex items-center gap-2" onClick={e => e.preventDefault()}>
        <span className="text-[10px] text-red-400">{error}</span>
        <button onClick={() => { setError(null); setConfirming(false); }}
          className="text-[10px] bg-gray-700 hover:bg-gray-600 text-gray-300 px-2 py-0.5 rounded">
          OK
        </button>
      </div>
    );
  }

  if (confirming) {
    return (
      <div className="flex items-center gap-2" onClick={e => e.preventDefault()}>
        <span className="text-[10px] text-red-400">Delete?</span>
        <button onClick={handleDelete} disabled={deleting}
          className="text-[10px] bg-red-600 hover:bg-red-500 text-white px-2 py-0.5 rounded disabled:opacity-50">
          {deleting ? '...' : 'Yes'}
        </button>
        <button onClick={() => setConfirming(false)} disabled={deleting}
          className="text-[10px] bg-gray-700 hover:bg-gray-600 text-gray-300 px-2 py-0.5 rounded">
          No
        </button>
      </div>
    );
  }

  return (
    <button onClick={(e) => { e.preventDefault(); e.stopPropagation(); setConfirming(true); }}
      className="text-gray-600 hover:text-red-400 transition-colors p-1 text-sm"
      title={`Delete ${runId}`}>
      🗑
    </button>
  );
}
