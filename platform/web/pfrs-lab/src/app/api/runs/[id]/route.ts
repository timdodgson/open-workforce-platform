import { NextResponse } from 'next/server';
import { revalidatePath } from 'next/cache';
import { getStorageProvider } from '@/lib/storage';

export async function DELETE(request: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;

  if (id.includes('..') || id.includes('/') || id.includes('\\')) {
    return NextResponse.json({ error: 'Invalid run ID' }, { status: 400 });
  }

  try {
    await getStorageProvider().hideRun(id);
    revalidatePath('/');
    return NextResponse.json({ success: true });
  } catch (err) {
    console.error(`Failed to hide run '${id}':`, err);
    return NextResponse.json({ error: String(err) }, { status: 500 });
  }
}
