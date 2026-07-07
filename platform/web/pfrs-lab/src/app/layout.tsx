import type { Metadata } from 'next';
import './globals.css';
import Sidebar from '@/components/Sidebar';

export const metadata: Metadata = {
  title: 'PFRS Lab',
  description: 'Multi-domain optimisation research platform with Search Intelligence',
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
