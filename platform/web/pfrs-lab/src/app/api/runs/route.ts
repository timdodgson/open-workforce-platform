import { NextResponse } from 'next/server';
import { loadRunMetadata } from '@/lib/data-loader';

export async function GET() {
  const meta = await loadRunMetadata();
  return NextResponse.json({ runs: meta ? [{ id: 'latest', ...meta }] : [] });
}
