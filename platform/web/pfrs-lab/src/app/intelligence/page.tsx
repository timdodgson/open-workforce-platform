import IntelligenceShell from './IntelligenceShell';
import type { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Search Intelligence',
  description: 'AI advisory system for optimisation. Monitors search behaviour and allocates compute safely and measurably.',
};

export const dynamic = 'force-dynamic';

export type { IntelligenceData } from '@/lib/intelligence';

export default function IntelligencePage() {
  return <IntelligenceShell />;
}
