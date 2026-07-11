'use client';

import { createContext, useContext, useLayoutEffect, useMemo, useState } from 'react';
import type { RunMode } from './run-mode';

export interface RunNavState {
  runId: string | null;
  runMode: RunMode | null;
}

interface RunNavContextValue {
  nav: RunNavState;
  setRunNav: (nav: RunNavState) => void;
}

const RunNavContext = createContext<RunNavContextValue | null>(null);

export function RunNavProvider({ children }: { children: React.ReactNode }) {
  const [nav, setRunNav] = useState<RunNavState>({ runId: null, runMode: null });
  const value = useMemo(() => ({ nav, setRunNav }), [nav]);
  return <RunNavContext.Provider value={value}>{children}</RunNavContext.Provider>;
}

export function useRunNav() {
  const ctx = useContext(RunNavContext);
  if (!ctx) {
    return {
      nav: { runId: null, runMode: null } satisfies RunNavState,
      setRunNav: () => {},
    };
  }
  return ctx;
}

/** Sync server-derived run mode from /runs/[id]/layout into AppShell sidebar. */
export function RunNavSetter({ runId, mode }: { runId: string; mode: RunMode }) {
  const { setRunNav } = useRunNav();

  useLayoutEffect(() => {
    setRunNav({ runId, runMode: mode });
    return () => setRunNav({ runId: null, runMode: null });
  }, [runId, mode, setRunNav]);

  return null;
}
