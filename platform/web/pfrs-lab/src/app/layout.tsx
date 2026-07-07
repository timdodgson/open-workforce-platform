import type { Metadata } from 'next';
import './globals.css';
import Sidebar from '@/components/Sidebar';

export const metadata: Metadata = {
  title: 'PFRS Lab',
  description: 'Multi-domain optimisation research platform with Search Intelligence',
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body className="bg-gray-950 text-gray-200 min-h-screen dark">
        <ThemeScript />
        <Sidebar />
        <main className="ml-56 p-6 max-w-[1200px]">
          {children}
        </main>
      </body>
    </html>
  );
}

function ThemeScript() {
  // Inline script to prevent flash — reads localStorage before paint.
  return (
    <script
      dangerouslySetInnerHTML={{
        __html: `(function(){try{var t=localStorage.getItem('pfrs-theme');if(t==='light'){document.documentElement.classList.remove('dark');document.documentElement.classList.add('light');document.body.style.background='#f8fafc';document.body.style.color='#1e293b'}}catch(e){}})()`,
      }}
    />
  );
}
