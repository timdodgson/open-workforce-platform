import { NextRequest, NextResponse } from 'next/server';
import {
  loadIntelligenceSection,
  type IntelligenceSection,
} from '@/lib/intelligence-data';

export const dynamic = 'force-dynamic';

const SECTIONS = new Set<IntelligenceSection>([
  'summary', 'learning', 'decisions', 'model', 'assist', 'policies',
  'continuous-learning', 'promotion', 'counterfactual',
]);

/** Sectioned intelligence API — avoids Lambda 6MB response limit on /intelligence. */
export async function GET(request: NextRequest) {
  const raw = request.nextUrl.searchParams.get('section') || 'summary';
  if (!SECTIONS.has(raw as IntelligenceSection)) {
    return NextResponse.json({ error: 'Invalid section' }, { status: 400 });
  }

  const data = await loadIntelligenceSection(raw as IntelligenceSection);
  return NextResponse.json(data);
}
