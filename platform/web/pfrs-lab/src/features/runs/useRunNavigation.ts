'use client';

import { usePathname } from 'next/navigation';
import { useEffect, useState } from 'react';
import { useRunNav } from './RunNavContext';
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

/** Resolves current run id + sidebar mode from layout context, RunMeta, then API fallback. */
export function useRunNavigation(): RunNavigation {
  const pathname = usePathname();
  const { nav } = useRunNav();
  const runId = runIdFromPath(pathname);
  const [runMode, setRunMode] = useState<RunMode | null>(null);

  const contextMode = runId && nav.runId === runId ? nav.runMode : null;

  useEffect(() => {
    if (!runId) {
      setRunMode(null);
      return;
    }

    if (contextMode) {
      setRunMode(contextMode);
      return;
    }

    const fromMeta = modeFromRunMeta(runId);
    if (fromMeta) {
      setRunMode(fromMeta);
      return;
    }

    let cancelled = false;
    fetch(`/api/runs/${encodeURIComponent(runId)}/meta`)
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
  }, [runId, pathname, contextMode]);

  return { runId, runMode: contextMode ?? runMode };
}
