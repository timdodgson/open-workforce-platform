import { NextResponse } from 'next/server';
import { loadWeeks } from '@/lib/data-loader';

export async function GET() {
  const weeks = await loadWeeks();
  return NextResponse.json({ weeks });
}
