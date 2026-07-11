'use client';

import AppHeader from '@/components/AppHeader';
import Sidebar from '@/components/Sidebar';
import { useRunNavigation } from '@/features/runs/useRunNavigation';

export default function AppShell({ children }: { children: React.ReactNode }) {
  const { runId, runMode } = useRunNavigation();

  return (
    <div className="min-h-screen flex flex-col">
      <AppHeader />
      <div className="flex flex-1 pt-14 min-h-0">
        <Sidebar runId={runId} runMode={runMode} />
        <main className="flex-1 min-w-0 p-6 overflow-x-hidden ml-56">{children}</main>
      </div>
    </div>
  );
}
