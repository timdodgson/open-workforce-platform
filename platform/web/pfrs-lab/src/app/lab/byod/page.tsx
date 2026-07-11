import type { Metadata } from 'next';
import ByodLabView from './ByodLabView';

export const metadata: Metadata = {
  title: 'BYOD / BYOA',
  description: 'Registered solvers, extension contract, and owp-sdk reference for bring-your-own-domain integrations.',
};

export default function ByodLabPage() {
  return <ByodLabView />;
}
