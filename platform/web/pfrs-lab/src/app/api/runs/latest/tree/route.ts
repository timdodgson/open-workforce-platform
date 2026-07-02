import { NextResponse } from 'next/server';
import { loadTree } from '@/lib/data-loader';

export async function GET() {
  const nodes = await loadTree();
  return NextResponse.json({ available: nodes.length > 0, nodes });
}
