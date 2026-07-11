import type { Metadata } from 'next';
import LandingPage from '@/features/landing/LandingPage';

export const metadata: Metadata = {
  title: 'Research',
  description:
    'Technical deep dive — algorithms, Search Intelligence, beam search, explainability, and validation evidence for PFRS Lab.',
};

export const dynamic = 'force-dynamic';

export default function ResearchPage() {
  return <LandingPage runs={[]} />;
}
