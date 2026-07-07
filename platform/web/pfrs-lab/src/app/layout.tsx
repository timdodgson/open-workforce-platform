import type { Metadata } from 'next';
import './globals.css';
import Sidebar from '@/components/Sidebar';

const BASE_URL = process.env.NEXT_PUBLIC_BASE_URL || 'https://pfrs-lab.com';

export const metadata: Metadata = {
  metadataBase: new URL(BASE_URL),
  title: {
    default: 'PFRS Lab — Adaptive Optimisation Research',
    template: '%s | PFRS Lab',
  },
  description:
    'Multi-domain optimisation research platform. Solves NRP, CVRP, JSS, and VRPTW with metaheuristic algorithms. Search Intelligence reduces compute by 40–73% with zero quality loss.',
  keywords: [
    'optimisation', 'metaheuristics', 'simulated annealing', 'tabu search',
    'nurse rostering', 'vehicle routing', 'job shop scheduling',
    'search intelligence', 'combinatorial optimisation', 'NP-hard',
  ],
  authors: [{ name: 'Tim Dodgson' }],
  openGraph: {
    type: 'website',
    locale: 'en_GB',
    siteName: 'PFRS Lab',
    title: 'PFRS Lab — Adaptive Optimisation Research',
    description:
      'Multi-domain optimisation research platform with Search Intelligence. 4 domains, 5 algorithms, 320+ validated runs.',
  },
  twitter: {
    card: 'summary_large_image',
    title: 'PFRS Lab — Adaptive Optimisation Research',
    description:
      'Multi-domain optimisation research platform with Search Intelligence.',
  },
  robots: {
    index: true,
    follow: true,
  },
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body className="bg-white text-gray-900 min-h-screen">
        <Sidebar />
        <main className="ml-56 p-6 max-w-[1200px]">
          {children}
        </main>
      </body>
    </html>
  );
}
