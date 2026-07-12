import type { Metadata } from 'next';
import ReferencesPage from '@/features/landing/ReferencesPage';

const TITLE = 'Standards and references for optimisation research';
const DESCRIPTION =
  'OR societies (OR Society, INFORMS, EURO), INRC-II, CVRPLIB, Solomon VRPTW, job-shop archives, HiGHS, and COIN-OR — standards this optimisation research sits against.';

export const metadata: Metadata = {
  title: TITLE,
  description: DESCRIPTION,
  alternates: { canonical: '/references' },
  openGraph: {
    title: `${TITLE} | PFRS Lab`,
    description: DESCRIPTION,
    url: '/references',
  },
  twitter: {
    title: `${TITLE} | PFRS Lab`,
    description: DESCRIPTION,
  },
};

export default function ReferencesRoute() {
  return <ReferencesPage />;
}
