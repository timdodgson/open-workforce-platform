'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';

const NAV = [
  { href: '/research', label: 'Research' },
  { href: '/reproduce', label: 'Reproduce' },
  { href: '/about', label: 'About' },
] as const;

export default function SiteShell({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();

  return (
    <div className="min-h-screen flex flex-col bg-gray-950 text-gray-100">
      <header className="sticky top-0 z-50 border-b border-gray-800/80 bg-gray-950/90 backdrop-blur-md">
        <div className="max-w-6xl mx-auto px-6 h-16 flex items-center justify-between gap-4">
          <Link href="/" className="min-w-0 group">
            <p className="text-sm font-bold text-blue-400 group-hover:text-blue-300 transition-colors">PFRS Lab</p>
            <p className="text-[10px] text-gray-500">Adaptive Optimisation Research</p>
          </Link>

          <nav className="hidden sm:flex items-center gap-6">
            {NAV.map(({ href, label }) => (
              <Link
                key={href}
                href={href}
                className={`text-sm transition-colors ${
                  pathname === href ? 'text-gray-100 font-medium' : 'text-gray-400 hover:text-gray-200'
                }`}
              >
                {label}
              </Link>
            ))}
            <a
              href="https://github.com/timdodgson/open-workforce-platform"
              target="_blank"
              rel="noopener noreferrer"
              className="text-sm text-gray-400 hover:text-gray-200 transition-colors"
            >
              GitHub
            </a>
          </nav>

          <Link
            href="/lab"
            className="shrink-0 text-sm font-semibold px-4 py-2 rounded-lg bg-blue-600 text-white hover:bg-blue-500 transition-colors"
          >
            Open Lab →
          </Link>
        </div>
      </header>

      <main className="flex-1">{children}</main>

      <footer className="border-t border-gray-800 mt-16">
        <div className="max-w-6xl mx-auto px-6 py-10 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div>
            <p className="text-sm font-medium text-gray-300">PFRS Lab</p>
            <p className="text-xs text-gray-500 mt-1">
              Open research platform · Built by Tim Dodgson
            </p>
          </div>
          <div className="flex flex-wrap gap-4 text-xs text-gray-500">
            <Link href="/research" className="hover:text-gray-300">Research depth</Link>
            <Link href="/reproduce" className="hover:text-gray-300">Cite &amp; reproduce</Link>
            <Link href="/about" className="hover:text-gray-300">About</Link>
            <Link href="/about#summary" className="hover:text-gray-300">Executive summary</Link>
            <Link href="/lab/byod" className="hover:text-gray-300">BYOD registry</Link>
            <Link href="/benchmarks" className="hover:text-gray-300">Benchmarks</Link>
            <a
              href="https://github.com/timdodgson/open-workforce-platform"
              target="_blank"
              rel="noopener noreferrer"
              className="hover:text-gray-300"
            >
              Source code
            </a>
          </div>
        </div>
      </footer>
    </div>
  );
}
