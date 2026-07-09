import type { RunMode } from '@/features/runs/run-mode';

/** SSR hint for Sidebar — avoids /api/runs/[id]/meta on every run sub-page navigation. */
export default function RunMeta({ runId, mode }: { runId: string; mode: RunMode }) {
  return <div id="run-meta" data-run-id={runId} data-run-mode={mode} hidden />;
}
