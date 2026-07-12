import type { Metadata } from 'next';
import GettingStartedView from './GettingStartedView';

const TITLE = 'Getting started — CLI & parameters';
const DESCRIPTION =
  'Install Go for concurrent PFRS workers, run worked examples across NRP, CVRP, JSS, and VRPTW, and use the full CLI / PFRS parameter reference.';

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
