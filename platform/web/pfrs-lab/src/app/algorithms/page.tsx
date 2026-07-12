import type { Metadata } from 'next';
import AlgorithmsPage from '@/features/landing/AlgorithmsPage';

const TITLE = 'Metaheuristic algorithms compared';
const DESCRIPTION =
  'Compare Simulated Annealing, LAHC, Tabu Search, Genetic Algorithms, and portfolio mode — strengths, trade-offs, and when PFRS Lab uses each on real OR benchmarks.';

export const metadata: Metadata = {
  title: TITLE,
  description: DESCRIPTION,
  alternates: { canonical: '/algorithms' },
  openGraph: {
    title: `${TITLE} | PFRS Lab`,
    description: DESCRIPTION,
    url: '/algorithms',
  },
  twitter: {
    title: `${TITLE} | PFRS Lab`,
    description: DESCRIPTION,
  },
};

export default function AlgorithmsRoute() {
  return <AlgorithmsPage />;
}
