import { NextResponse } from 'next/server';
import { listRunsAsync } from '@/lib/data-loader';

export async function GET() {
  const runs = await listRunsAsync();
  return NextResponse.json({
    runs: runs.map((run) => ({
      id: run.id,
      timestamp: run.timestamp,
      ...run.metadata,
    })),
  });
}
