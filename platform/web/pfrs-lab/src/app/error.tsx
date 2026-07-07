'use client';

interface ErrorProps {
  error: Error & { digest?: string };
  reset: () => void;
}

export default function GlobalError({ error, reset }: ErrorProps) {
  return (
    <div className="flex flex-col items-center justify-center py-20 px-6 text-center">
      <h2 className="text-lg font-semibold text-slate-800 mb-2">Something went wrong</h2>
      <p className="text-sm text-slate-500 mb-4 max-w-md">
        {error.message || 'An unexpected error occurred while loading this page.'}
      </p>
      {error.digest && (
        <p className="text-xs text-slate-400 mb-4 font-mono">Error ID: {error.digest}</p>
      )}
      <button
        onClick={reset}
        className="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded hover:bg-blue-500 transition-colors"
      >
        Try again
      </button>
    </div>
  );
}
