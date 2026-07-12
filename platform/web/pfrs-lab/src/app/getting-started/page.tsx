import type { Metadata } from 'next';
import GettingStartedView from './GettingStartedView';

const TITLE = 'Getting started';
const DESCRIPTION =
  'Install Go for concurrent PFRS workers, run worked examples across NRP, CVRP, JSS, and VRPTW, and use a dependency-aware CLI switch reference.';

export const metadata: Metadata = {
  title: TITLE,
  description: DESCRIPTION,
  alternates: { canonical: '/getting-started' },
  openGraph: {
    title: `${TITLE} | PFRS Lab`,
    description: DESCRIPTION,
    url: '/getting-started',
  },
};

export default function GettingStartedPage() {
  return <GettingStartedView />;
}
