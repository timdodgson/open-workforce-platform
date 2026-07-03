import { NextResponse } from 'next/server';
import { getStorageProvider } from '@/lib/storage';

export async function GET(request: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;

  if (id.includes('..') || id.includes('/') || id.includes('\\')) {
    return NextResponse.json({ error: 'Invalid run ID' }, { status: 400 });
  }

  const storage = getStorageProvider();
  const content = await storage.readFile(id, 'run.json');

  if (!content) {
    return NextResponse.json({ mode: 'pfrs' }); // Default to PFRS if no metadata.
  }

  try {
    const meta = JSON.parse(content);
    return NextResponse.json({ mode: meta.mode ?? 'pfrs' });
  } catch {
    return NextResponse.json({ mode: 'pfrs' });
  }
}
