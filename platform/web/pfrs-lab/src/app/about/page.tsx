import type { Metadata } from 'next';
import AboutPage from '@/features/landing/AboutPage';

export const metadata: Metadata = {
  title: 'About Tim Dodgson',
  description:
    'Tim Dodgson — Principal Software Engineer. Journey from City & Guilds mechanic and electronics repair through networking and a First-class BSc to CDL Principal; builder of PFRS Lab.',
  alternates: { canonical: '/about' },
  openGraph: {
    title: 'About Tim Dodgson | PFRS Lab',
    description:
      'From trades and component-level repair to Principal SWE and PFRS Lab — career journey, leadership, certifications, and measurable optimisation research.',
    url: '/about',
  },
};

export default function AboutRoute() {
  return <AboutPage />;
}
