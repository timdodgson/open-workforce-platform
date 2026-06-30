import { NextResponse } from 'next/server';
import { loadRunSummary } from '@/lib/data-loader';

export async function GET() {
  const summary = await loadRunSummary();
  return NextResponse.json(summary);
}
