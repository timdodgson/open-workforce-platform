import { NextResponse } from 'next/server';
import { existsSync, rmSync } from 'fs';
import path from 'path';

const RUNS_DIR = path.join(process.cwd(), 'data', 'runs');

export async function DELETE(request: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;

  // Sanitise: prevent path traversal.
  if (id.includes('..') || id.includes('/') || id.includes('\\')) {
    return NextResponse.json({ error: 'Invalid run ID' }, { status: 400 });
  }

  const runDir = path.join(RUNS_DIR, id);
  if (!existsSync(runDir)) {
    return NextResponse.json({ error: 'Run not found' }, { status: 404 });
  }

  rmSync(runDir, { recursive: true, force: true });
  return NextResponse.json({ success: true });
}
