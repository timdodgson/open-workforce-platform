import type { Metadata } from 'next';
import './globals.css';
import Sidebar from '@/components/Sidebar';

export const metadata: Metadata = {
  title: 'PFRS Research Lab',
  description: 'Parallel Feasible Roster Search — Performance Analysis',
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body className="bg-gray-950 text-gray-200 min-h-screen">
        <Sidebar />
        <main className="ml-56 p-6 max-w-[1200px]">
          {children}
        </main>
      </body>
    </html>
  );
}
