import IntelligenceShell from './IntelligenceShell';
import { loadIntelligenceData } from '@/lib/intelligence-data';
import type { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Search Intelligence',
  description: 'AI advisory system for optimisation. Monitors search behaviour and allocates compute safely and measurably.',
};

// Must be dynamic: ISR cached empty data at build time when no runs exist in CI/deploy.
export const dynamic = 'force-dynamic';

export type { IntelligenceData } from '@/lib/intelligence-data';

export default async function IntelligencePage() {
  const data = await loadIntelligenceData();
  return <IntelligenceShell data={data} />;
}
