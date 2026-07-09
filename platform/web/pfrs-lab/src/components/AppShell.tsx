'use client';

import { usePathname } from 'next/navigation';
import { useEffect, useState } from 'react';
import Sidebar from '@/components/Sidebar';
import type { RunMode } from '@/features/runs/run-mode';

export default function AppShell({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const [runMode, setRunMode] = useState<RunMode | null>(null);
  const [runId, setRunId] = useState<string | null>(null);

  useEffect(() => {
    const match = pathname.match(/^\/runs\/([^/]+)/);
    const id = match ? match[1] : null;
    setRunId(id);

    if (!id) {
      setRunMode(null);
      return;
    }

    const meta = document.getElementById('run-meta');
    if (meta?.dataset.runId === id && meta.dataset.runMode) {
      setRunMode(meta.dataset.runMode as RunMode);
      return;
    }

    fetch(`/api/runs/${id}/meta`)
      .then((res) => (res.ok ? res.json() : null))
      .then((data) => setRunMode((data?.mode as RunMode) ?? 'pfrs'))
      .catch(() => setRunMode('pfrs'));
  }, [pathname]);

  return (
    <>
      <Sidebar runId={runId} runMode={runMode} />
      <main className="ml-56 p-6 max-w-[1200px]">{children}</main>
    </>
  );
}
