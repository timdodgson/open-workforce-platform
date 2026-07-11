import type { Metadata } from 'next';
import ReproducePage from '@/features/landing/ReproducePage';

export const metadata: Metadata = {
  title: 'Cite & Reproduce',
  description:
    'Academic citation guide and student reproduction path for PFRS Lab — browse live results, reproduce one run locally, or replicate validation suites.',
};

export default function ReproduceRoute() {
  return <ReproducePage />;
}
