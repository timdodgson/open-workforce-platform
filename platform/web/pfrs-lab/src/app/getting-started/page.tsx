import type { Metadata } from 'next';
import GettingStartedView from './GettingStartedView';

export const metadata: Metadata = {
  title: 'Getting Started',
  description:
    'Install Go for concurrent PFRS workers, run worked examples across NRP/CVRP/JSS/VRPTW, and use the full CLI switch reference with dependencies.',
};

export default function GettingStartedPage() {
  return <GettingStartedView />;
}
