import { NextResponse } from 'next/server';
import { getLatestRunId, loadWeeks } from '@/lib/data-loader';

export async function GET() {
  const runId = await getLatestRunId();
  if (!runId) return NextResponse.json({ weeks: [] });
  const weeks = await loadWeeks(runId);
  return NextResponse.json({ weeks });
}
