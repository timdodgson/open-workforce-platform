import type { Metadata } from 'next';
import GettingStartedView from './GettingStartedView';

const TITLE = 'Getting started — Quick start & CLI';
const DESCRIPTION =
  'Five-minute Quick start on CVRPLIB A-n32-k5, then deeper examples for NRP beam and a full CLI / PFRS parameter reference.';

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
