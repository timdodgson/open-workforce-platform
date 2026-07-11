import type { Metadata } from 'next';
import AboutPage from '@/features/landing/AboutPage';

export const metadata: Metadata = {
  title: 'About',
  description: 'The story behind PFRS Lab — from university dissertation to production AI-native optimisation research.',
};

export default function AboutRoute() {
  return <AboutPage />;
}
