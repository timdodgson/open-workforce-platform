import { NextResponse } from 'next/server';
import { getStorageProvider } from '@/lib/storage';

export async function DELETE(request: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;

  // Sanitise: prevent path traversal.
  if (id.includes('..') || id.includes('/') || id.includes('\\')) {
    return NextResponse.json({ error: 'Invalid run ID' }, { status: 400 });
  }

  try {
    await getStorageProvider().hideRun(id);
    return NextResponse.json({ success: true });
  } catch (err) {
    return NextResponse.json({ error: String(err) }, { status: 500 });
  }
}
