import { NextResponse } from 'next/server';
import { getLatestRunId, loadTree } from '@/lib/data-loader';

export async function GET() {
  const runId = await getLatestRunId();
  if (!runId) return NextResponse.json({ available: false, nodes: [] });
  const nodes = await loadTree(runId);
  return NextResponse.json({ available: nodes.length > 0, nodes });
}
