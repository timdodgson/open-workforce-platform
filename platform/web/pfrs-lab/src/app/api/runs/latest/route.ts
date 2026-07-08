import { NextResponse } from 'next/server';
import { getLatestRunId, loadRunSummary } from '@/lib/data-loader';

export async function GET() {
  const runId = await getLatestRunId();
  if (!runId) return NextResponse.json({ error: 'No runs found' }, { status: 404 });
  const summary = await loadRunSummary(runId);
  return NextResponse.json(summary);
}
