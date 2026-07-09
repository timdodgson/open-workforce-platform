import { loadRunMetadata } from '@/lib/data-loader';
import { RunProvider } from '@/features/runs/RunContext';
import { deriveRunMode } from '@/features/runs/run-mode';
import RunMeta from '@/components/RunMeta';

export const dynamic = 'force-dynamic';

export default async function RunLayout({
  children,
  params,
}: {
  children: React.ReactNode;
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  const metadata = await loadRunMetadata(id);
  const mode = deriveRunMode(metadata);

  return (
    <RunProvider runId={id} metadata={metadata} mode={mode}>
      <RunMeta runId={id} mode={mode} />
      {children}
    </RunProvider>
  );
}
