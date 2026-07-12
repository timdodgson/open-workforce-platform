import type { Metadata } from 'next';
import DomainsPage from '@/features/landing/DomainsPage';

const TITLE = 'Optimisation domains that map to real operations';
const DESCRIPTION =
  'Nurse rostering, capacitated vehicle routing, VRPTW, and job shop scheduling — real-world stakes, why they are NP-hard, and the published benchmarks PFRS Lab measures against.';

export const metadata: Metadata = {
  title: TITLE,
  description: DESCRIPTION,
  alternates: { canonical: '/domains' },
  openGraph: {
    title: `${TITLE} | PFRS Lab`,
    description: DESCRIPTION,
    url: '/domains',
  },
  twitter: {
    title: `${TITLE} | PFRS Lab`,
    description: DESCRIPTION,
  },
};

export default function DomainsRoute() {
  return <DomainsPage />;
}
