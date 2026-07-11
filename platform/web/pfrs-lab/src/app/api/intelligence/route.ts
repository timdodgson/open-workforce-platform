import { NextRequest, NextResponse } from 'next/server';
import {
  DEFAULT_PAGE_LIMIT,
  loadIntelligenceSection,
  type IntelligenceSection,
} from '@/lib/intelligence';

export const dynamic = 'force-dynamic';

const SECTIONS = new Set<IntelligenceSection>([
  'summary', 'learning', 'decisions', 'model', 'assist', 'policies',
  'continuous-learning', 'promotion', 'counterfactual',
]);

const CACHE_MAX_AGE: Partial<Record<IntelligenceSection, number>> = {
  summary: 120,
  promotion: 120,
  model: 120,
  'continuous-learning': 90,
  counterfactual: 90,
  policies: 60,
  assist: 60,
  learning: 45,
  decisions: 45,
};

/** Sectioned intelligence API — avoids Lambda 6MB response limit on /intelligence. */
export async function GET(request: NextRequest) {
  const raw = request.nextUrl.searchParams.get('section') || 'summary';
  if (!SECTIONS.has(raw as IntelligenceSection)) {
    return NextResponse.json({ error: 'Invalid section' }, { status: 400 });
  }

  const section = raw as IntelligenceSection;
  const offset = Math.max(0, Number(request.nextUrl.searchParams.get('offset') || 0));
  const limit = Math.min(
    500,
    Math.max(1, Number(request.nextUrl.searchParams.get('limit') || DEFAULT_PAGE_LIMIT)),
  );

  const data = await loadIntelligenceSection(section, undefined, { offset, limit });
  const maxAge = CACHE_MAX_AGE[section] ?? 60;
  return NextResponse.json(data, {
    headers: { 'Cache-Control': `public, max-age=${maxAge}, stale-while-revalidate=30` },
  });
}
