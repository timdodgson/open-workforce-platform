'use client';

import { createContext, useContext } from 'react';
import type { RunMetadata } from '@/lib/types';
import type { RunMode } from './run-mode';

export interface RunContextValue {
  runId: string;
  metadata: RunMetadata | null;
  mode: RunMode;
}

const RunContext = createContext<RunContextValue | null>(null);

export function RunProvider({
  runId,
  metadata,
  mode,
  children,
}: RunContextValue & { children: React.ReactNode }) {
  return (
    <RunContext.Provider value={{ runId, metadata, mode }}>
      {children}
    </RunContext.Provider>
  );
}

export function useRunContext(): RunContextValue | null {
  return useContext(RunContext);
}
