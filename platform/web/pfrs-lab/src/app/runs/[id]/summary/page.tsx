import { renderRunSummary } from '@/features/runs/summary/renderRunSummary';

export const dynamic = 'force-dynamic';

export default async function RunSummaryPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  return renderRunSummary(id);
}
