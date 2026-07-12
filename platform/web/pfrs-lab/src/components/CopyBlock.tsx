'use client';

import { useState } from 'react';

/** Copyable shell / CLI snippet for lab and marketing pages. */
export default function CopyBlock({ text }: { text: string }) {
  const [copied, setCopied] = useState(false);
  return (
    <div className="relative group">
      <pre className="text-[11px] leading-relaxed text-gray-200 bg-gray-950 border border-gray-800 rounded-lg p-3 overflow-x-auto whitespace-pre-wrap font-mono">
        {text}
      </pre>
      <button
        type="button"
        onClick={async () => {
          await navigator.clipboard.writeText(text);
          setCopied(true);
          window.setTimeout(() => setCopied(false), 1500);
        }}
        className="absolute top-2 right-2 text-[10px] px-2 py-1 rounded border border-gray-700 bg-gray-900 text-gray-400 hover:text-gray-200"
      >
        {copied ? 'Copied' : 'Copy'}
      </button>
    </div>
  );
}
