import { NextResponse } from 'next/server';
import { getAuthProvider } from '@/lib/auth/provider';
import { buildIntelligenceArtifacts } from '@/lib/intelligence-data';

export const dynamic = 'force-dynamic';
export const maxDuration = 300;

/** Admin-only: scan runs and write intelligence_summary / learning / policy artifacts to S3. */
export async function POST() {
  const auth = getAuthProvider();
  if (!(await auth.isAdmin())) {
    return NextResponse.json({ error: 'Forbidden' }, { status: 403 });
  }

  try {
    const result = await buildIntelligenceArtifacts();
    return NextResponse.json(result);
  } catch (e) {
    const message = e instanceof Error ? e.message : 'Rebuild failed';
    return NextResponse.json({ error: message }, { status: 500 });
  }
}
