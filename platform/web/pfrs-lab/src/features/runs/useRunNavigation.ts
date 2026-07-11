'use client';

import { usePathname } from 'next/navigation';
import { useEffect, useState } from 'react';
import type { RunMode } from './run-mode';

export interface RunNavigation {
  runId: string | null;
  runMode: RunMode | null;
}

function runIdFromPath(pathname: string): string | null {
  const match = pathname.match(/^\/runs\/([^/]+)/);
  return match ? match[1] : null;
}

function modeFromRunMeta(runId: string): RunMode | null {
  if (typeof document === 'undefined') return null;
  const meta = document.getElementById('run-meta');
  if (meta?.dataset.runId === runId && meta.dataset.runMode) {
    return meta.dataset.runMode as RunMode;
  }
  return null;
}

/** Resolves current run id + sidebar mode from SSR RunMeta, with one API fallback. */
export function useRunNavigation(): RunNavigation {
  const pathname = usePathname();
  const [runMode, setRunMode] = useState<RunMode | null>(null);
  const runId = runIdFromPath(pathname);

  useEffect(() => {
    if (!runId) {
      setRunMode(null);
      return;
    }

    const fromMeta = modeFromRunMeta(runId);
    if (fromMeta) {
      setRunMode(fromMeta);
      return;
    }

    let cancelled = false;
    fetch(`/api/runs/${runId}/meta`)
      .then((res) => (res.ok ? res.json() : null))
      .then((data) => {
        if (!cancelled) setRunMode((data?.mode as RunMode) ?? 'pfrs');
      })
      .catch(() => {
        if (!cancelled) setRunMode('pfrs');
      });

    return () => {
      cancelled = true;
    };
  }, [runId, pathname]);

  return { runId, runMode };
}
