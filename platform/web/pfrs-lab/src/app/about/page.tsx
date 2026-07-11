import type { Metadata } from 'next';
import AboutPage from '@/features/landing/AboutPage';

export const metadata: Metadata = {
  title: 'About Tim Dodgson',
  description:
    'Executive summary and portfolio context for PFRS Lab — Principal Software Engineer, multi-domain optimisation, Search Intelligence, 320+ validated runs.',
};

export default function AboutRoute() {
  return <AboutPage />;
}
